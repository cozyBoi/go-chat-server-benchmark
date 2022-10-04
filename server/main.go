package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	conn *websocket.Conn
	get  chan Payload
}

type ConnInfo struct {
	conn   Client
	roomid int
}

type Payload struct {
	UserId int    `json:userid`
	RoomId int    `json:roomid`
	Msg    string `json:msg`
}

var broadcast_cnt chan int

var send_msg_cnt chan int

var upgrader = websocket.Upgrader{} // use default options

//var conns []*Client

var connMap map[int][]*Client

var C_chan chan ConnInfo

var mg_client *mongo.Client

func broadcast_msg(conn *websocket.Conn, pld Payload, roomid int) int {
	var ret int
	for _, curr_conn := range connMap[roomid] {
		if curr_conn.conn == conn {
			continue
		}
		curr_conn.get <- pld
		ret++
	}
	return ret
}

func serve_ws(ctx echo.Context) error {
	coll := mg_client.Database("go-chat-server").Collection("chat-info")
	var init_s, prs_s time.Time
	init_s = time.Now()
	var b_cnt int
	var s_cnt int

	c, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Print("upgrade:", err)
		return err
	}

	defer func() {
		c.Close()
		broadcast_cnt <- b_cnt
		send_msg_cnt <- s_cnt
	}()

	currClnt := new(Client)
	currClnt.conn = c
	currClnt.get = make(chan Payload)

	//Need WebServe Chat Room Information Handshaking OR GET from DB

	//*** get chat room info ***
	//get chat room num
	_, msg, _ := c.ReadMessage()
	roomNumber, _ := strconv.Atoi(string(msg))
	//get chat room lists
	for i := 0; i < roomNumber; i++ {
		_, msg, _ := c.ReadMessage()
		roomIdx, _ := strconv.Atoi(string(msg))
		C_chan <- ConnInfo{conn: *currClnt, roomid: roomIdx}
	}

	fmt.Println("[init lat]", time.Since(init_s))

	go func() {
		for newPld := range currClnt.get {
			pld, _ := json.Marshal(Payload{UserId: 0, RoomId: newPld.RoomId, Msg: newPld.Msg})
			currClnt.conn.WriteJSON(pld)
			s_cnt++
		}
	}()

	for {
		var newPld Payload
		//var newJson []byte
		currClnt.conn.ReadJSON(&newPld) //get msg type and msg

		docs := bson.D{{"room", newPld.RoomId}, {"id", newPld.UserId}, {"msg", newPld.Msg}}
		coll.InsertOne(context.TODO(), docs)

		if err != nil {
			log.Println("read:", err)
			break
		}

		prs_s = time.Now()
		b_cnt += broadcast_msg(c, newPld, newPld.RoomId)
		fmt.Println("[broad lat]", time.Since(prs_s))
	}
	return err
}

func conn_mng() {
	fmt.Println("conn_manage start")
	for new_clnt := range C_chan {
		if _, ok := connMap[new_clnt.roomid]; ok {
			connMap[new_clnt.roomid] = append(connMap[new_clnt.roomid], &new_clnt.conn)
		} else {
			connMap[new_clnt.roomid] = make([]*Client, 0, 10)
			connMap[new_clnt.roomid] = append(connMap[new_clnt.roomid], &new_clnt.conn)
		}
	}
}

func initFunc() {
	C_chan = make(chan ConnInfo)
	broadcast_cnt = make(chan int)
	send_msg_cnt = make(chan int)
}

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mg_client = client

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()
	connMap = make(map[int][]*Client)

	var t_b_cnt int
	var t_s_cnt int

	initFunc()

	e := echo.New()
	e.GET("/ws", serve_ws)

	go conn_mng()

	go func() {
		for {
			select {
			case ret := <-broadcast_cnt:
				t_b_cnt += ret
			case ret := <-send_msg_cnt:
				t_s_cnt += ret
			}
		}
	}()

	go func() {
		for t := range ticker.C {
			fmt.Println(t)
			fmt.Println("total broadcast: ", t_b_cnt)
			fmt.Println("total sending: ", t_s_cnt)
		}
	}()
	log.Fatal(e.Start(":9100"))
}
