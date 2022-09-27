package main

import "net/http"

func testFunc() {
	http.Get("http://localhost:9100")
	http.Get("http://localhost:9100/cookie")
	http.Get("http://localhost:9100/rooms")
	http.Get("http://localhost:9100/rooms/1/ws")
	for {
		//websocket thing
	}
}

func main() {
	go testFunc()
}
