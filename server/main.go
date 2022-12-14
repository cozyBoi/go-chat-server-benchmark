package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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
	UserId    int    `json:userid`
	RoomId    int    `json:roomid`
	Msg       string `json:msg`
	issueTime time.Time
}

var broadcast_cnt chan int

var send_msg_cnt chan int

var upgrader = websocket.Upgrader{} // use default options

//var conns []*Client

var connMap map[int][]*Client

var C_chan chan ConnInfo

var db_chan chan Payload

func broadcast_msg(conn *websocket.Conn, pld Payload, roomid int) int {
	var ret int
	for _, curr_conn := range connMap[roomid] {
		if curr_conn.conn == conn {
			continue
		}
		pld.issueTime = time.Now()
		curr_conn.get <- pld
		ret++
	}
	return ret
}

func serve_ws(ctx echo.Context) error {
	var b_cnt int
	var s_cnt int

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

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

	go func() {
		for t := range ticker.C {
			fmt.Println(t)
			broadcast_cnt <- b_cnt
			send_msg_cnt <- s_cnt
			b_cnt = 0
			s_cnt = 0
		}
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

	go func() {
		for newPld := range currClnt.get {
			//fmt.Println("[broad lat]", time.Since(newPld.issueTime))
			pld, _ := json.Marshal(Payload{UserId: 0, RoomId: newPld.RoomId, Msg: newPld.Msg})
			currClnt.conn.WriteJSON(pld)
			s_cnt++
		}
	}()

	for {
		var newPld Payload
		//var newJson []byte
		currClnt.conn.ReadJSON(&newPld) //get msg type and msg
		if err != nil {
			log.Println("read:", err)
			break
		}
		newPld.issueTime = time.Now()
		//db_chan <- newPld

		pbytes, _ := json.Marshal(newPld)
		buff := bytes.NewBuffer(pbytes)
		http.Post("http://jupiter03:9200/db", "application/json", buff)

		b_cnt += broadcast_msg(c, newPld, newPld.RoomId)
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
	db_chan = make(chan Payload)
	broadcast_cnt = make(chan int)
	send_msg_cnt = make(chan int)
}

func main() {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/testdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
				println("broad", t_b_cnt)
			case ret := <-send_msg_cnt:
				t_s_cnt += ret
				println("send", t_s_cnt)
			}
		}
	}()

	go func() {
		for newPld := range db_chan {
			//fmt.Println("INSERT INTO chat_info VALUE (" + strconv.Itoa(newPld.UserId) + ", " + strconv.Itoa(newPld.RoomId) + ", " + "\"" + newPld.Msg + "\", " + "\"" + "hi" + "\");")
			//fmt.Println("[db lat]", time.Since(newPld.issueTime))
			_, err := db.Exec("INSERT INTO chat_info VALUE (" + strconv.Itoa(newPld.UserId) + ", " + strconv.Itoa(newPld.RoomId) + ", " + "\"" + newPld.Msg + "\", " + "\"" + "hi" + "\");")
			//result, err := db.Exec("INSERT INTO chat_info VALUE (" + strconv.Itoa(newPld.UserId) + ", " + strconv.Itoa(newPld.RoomId) + ", " + "\"" + newPld.Msg + "\", " + "\"" + "hi" + "\");")
			if err != nil {
				fmt.Println(err)
			}
			//nRow, err := result.RowsAffected()
			//fmt.Println("insert counts: ", nRow)
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
