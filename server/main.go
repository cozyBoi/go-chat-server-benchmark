package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type bson_struct struct {
	Timestamp int64  `bson:"timestamp"`
	RoomId    int    `bson:"room"`
	UserId    string `bson:"id"`
	Msg       string `bson:"msg"`
}

var mg_client *mongo.Client
var coll *mongo.Collection

var upgrader = websocket.Upgrader{
	EnableCompression: true,
} // use default options

var chat_log []string

var conns map[int][]*websocket.Conn //chatroom id => map => conn[]
var chatCache map[int]*queue        //cache 30 chats

func broadcast_msg(conn *websocket.Conn, msg []byte, roomId int) {
	connsArr := conns[roomId]
	for _, curr_conn := range connsArr {
		if curr_conn == conn {
			continue
		}
		curr_conn.WriteMessage(1, msg)
	}
}

func connClose(c *websocket.Conn, rid int) {
	for i, curr_conn := range conns[rid] {
		if curr_conn == c {
			conns[rid] = append(conns[rid][:i], conns[rid][i+1:]...) //... => unpack the slice
		}
	}
	c.Close()
}

func socketHandler(ctx echo.Context) error {
	fmt.Println("get sock")
	cid := "0"
	fmt.Println("get sock2")

	//upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	c, _ := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	roomIdStr := ctx.Param("id")
	roomId, _ := strconv.Atoi(roomIdStr)
	defer connClose(c, roomId)
	_, flag := conns[roomId]
	if !flag {
		conns[roomId] = make([]*websocket.Conn, 0, 10)
		chatCache[roomId] = NewQ()

		opts := options.Find().SetSort(bson.D{{"timestamp", 1}})
		var thirty int64 = 30
		opts.Limit = &thirty

		filter := bson.D{{"room", roomId}}
		cursor, _ := coll.Find(context.TODO(), filter, opts)
		var results []bson.M
		cursor.All(context.TODO(), &results)
		for _, result := range results {
			var bs bson_struct
			bsonBytes, _ := bson.Marshal(result)
			bson.Unmarshal(bsonBytes, &bs)
			chatCache[roomId].Push(buff{Msg: bs.Msg, Sender: bs.UserId})
		}
	}

	conns[roomId] = append(conns[roomId], c)

	for {
		_, msg, err := c.ReadMessage() //get msg type and msg
		if err != nil {
			return err
		}
		if chatCache[roomId].IsFull() {
			chatCache[roomId].Pop()
		}
		str_msg := string(msg)
		newChat := buff{Msg: str_msg, Sender: cid}
		docs := bson.D{{"timestamp", time.Now().UnixNano() / int64(time.Millisecond)}, {"room", roomId}, {"id", cid}, {"msg", str_msg}}
		coll.InsertOne(context.TODO(), docs)
		chatCache[roomId].Push(newChat)
		broadcast_msg(c, msg, roomId)
	}
}

var roomNumber int = 5
var chatRooms = []string{"1", "2", "3", "4", "5"}

func roomsHandler(ctx echo.Context) error {
	var curr_error error
	for i, room := range chatRooms {
		if i == roomNumber-1 {
			curr_error = ctx.String(http.StatusOK, room)
		} else {
			curr_error = ctx.String(http.StatusOK, room+",")
		}
	}
	return curr_error
}

func sendPrevChats(ctx echo.Context) error {
	var curr_error error
	roomIdStr := ctx.Param("id")
	roomId, _ := strconv.Atoi(roomIdStr)

	if chatCache[roomId] == nil {
		return ctx.NoContent(http.StatusOK)
	} else if chatCache[roomId].size == 0 {
		return ctx.NoContent(http.StatusOK)
	}

	for i := 0; i < chatCache[roomId].size; i++ {
		curr_idx := (chatCache[roomId].front + i) % 30
		fmt.Println(chatCache[roomId].buf[curr_idx])
		curr_error = ctx.JSON(http.StatusOK, &chatCache[roomId].buf[curr_idx])
	}
	return curr_error
}

func changeRoomHandler(ctx echo.Context) error {
	return ctx.File("assets/comm.html")
}

func roomsCreate(ctx echo.Context) error {
	roomNumber++
	chatRooms = append(chatRooms, strconv.Itoa(roomNumber))
	return ctx.NoContent(http.StatusOK)
}

func main() {
	conns = make(map[int][]*websocket.Conn)
	chatCache = make(map[int]*queue)
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mg_client = client
	coll = mg_client.Database("go-chat-server").Collection("chat-info")
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	e := echo.New()

	e.Static("/", "assets")
	e.Static("/rooms", "assets")

	e.File("/", "assets/main.html")
	e.File("/rooms/:id", "assets/comm.html")

	e.GET("/rooms/:id/ws", socketHandler)
	e.GET("/rooms/:id/chats", sendPrevChats)
	e.GET("/rooms", roomsHandler)
	e.GET("/cookie", writeCookie)

	e.POST("/rooms", roomsCreate)

	e.Start(":9100")
}
