package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/labstack/echo/v4"
)

type Payload struct {
	UserId    int    `json:userid`
	RoomId    int    `json:roomid`
	Msg       string `json:msg`
	issueTime time.Time
}

func dbHandler(c echo.Context) error {
	var getPld Payload
	err := c.Bind(&getPld); if err != nil{
		return c.String(http.StatusBadRequest, "bad request")
	}
	fmt.Println(getPld)
	return c.String(http.StatusOK, "good request")
}


func main() {
	e := echo.New()

	e.POST("/db", dbHandler)

	e.Start(":9200")
}
