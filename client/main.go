package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Payload struct {
	UserId int    `json:"userid"`
	RoomId int    `json:"roomid"`
	Msg    string `json:"msg"`
}

var addr = flag.String("addr", "localhost:9100", "http service address")

var done chan bool
var room_number_per_user int = 5
var client_number int = 50
var t_room_number int = client_number / room_number_per_user //user randomly placed on 20 room, average user / room => 20

func testFunc(id int, roomList []int) {
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
			var newPld Payload
			var newJson []byte
			c.ReadJSON(newJson)
			json.Unmarshal(newJson, &newPld)
			if err != nil {
				log.Println("read:", err)
				return
			}
		}
	}()

	//*** send chat room info ***
	//send chat room num
	c.WriteMessage(1, []byte(strconv.Itoa(room_number_per_user)))
	//send chat room lists
	for i := 0; i < room_number_per_user; i++ {
		c.WriteMessage(1, []byte(strconv.Itoa(roomList[i])))
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var roomIdx int = 0

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			newPld := &Payload{UserId: 0, RoomId: roomList[roomIdx], Msg: t.String()}
			err := c.WriteJSON(newPld)
			roomIdx++
			if room_number_per_user >= roomIdx {
				roomIdx = 0
			}
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
		//generate 20 random number
		var roomList []int
		for {
			var flag bool = false
			rand_num := rand.Intn(t_room_number)
			if len(roomList) < room_number_per_user {
				for r := range roomList {
					if rand_num == r {
						flag = true
						break
					}
				}

				if flag {
					continue
				}

				roomList = append(roomList, rand_num)
			} else {
				break
			}
		}
		go testFunc(i, roomList)
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
