package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:9100", "http service address")

var done chan bool
var client_number int = 200

func testFunc(id int) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go func() {
		defer close(done)
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			//fmt.Printf("get from id %d\n", id)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func main() {
	done = make(chan bool)

	for i := 0; i < client_number; i++ {
		go testFunc(i)
	}
	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	go func() {
		for t := range ticker.C {
			fmt.Println(t)
			done <- true
			return
		}
	}()

	var wait sync.WaitGroup
	wait.Add(client_number)
	wait.Wait()
}
