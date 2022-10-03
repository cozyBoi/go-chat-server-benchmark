package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Client struct {
	conn *websocket.Conn
	get  chan []byte
}

var broadcast_cnt chan int

var send_msg_cnt chan int

var upgrader = websocket.Upgrader{} // use default options

var conns []*Client

var C_chan chan *Client

func broadcast_msg(conn *websocket.Conn, msg []byte) int {
	var ret int
	for _, curr_conn := range conns {
		if curr_conn.conn == conn {
			continue
		}
		curr_conn.get <- msg
		ret++
	}
	return ret
}

func serve_ws(ctx echo.Context) error {
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
	currClnt.get = make(chan []byte)
	C_chan <- currClnt
	fmt.Println("[init lat]", time.Since(init_s))

	go func() {
		for newMsg := range currClnt.get {
			currClnt.conn.WriteMessage(1, newMsg)
			//fmt.Println("writing")
			s_cnt++
		}
	}()

	for {
		_, message, err := currClnt.conn.ReadMessage() //get msg type and msg
		//fmt.Println("Reading", message)
		if err != nil {
			log.Println("read:", err)
			break
		}

		prs_s = time.Now()
		b_cnt += broadcast_msg(c, message)
		fmt.Println("[broad lat]", time.Since(prs_s))
	}
	return err
}

func conn_mng() {
	fmt.Println("conn_manage start")
	for new_clnt := range C_chan {
		conns = append(conns, new_clnt)
	}
}

func initFunc() {
	C_chan = make(chan *Client)
	broadcast_cnt = make(chan int)
	send_msg_cnt = make(chan int)
}

func main() {
	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

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
