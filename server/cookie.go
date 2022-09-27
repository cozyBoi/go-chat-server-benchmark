//future work
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

var cookieNum int = 0

func readCookie(c echo.Context) error {
	cookie, err := c.Cookie("cid")
	if err != nil {
		return err
	}
	fmt.Println(cookie.Name)
	fmt.Println(cookie.Value)
	return c.String(http.StatusOK, "read a cookie")
}

func writeCookie(c echo.Context) error {
	cookie, err := c.Cookie("cid")
	if err != nil {
		cookie = new(http.Cookie)
		cookie.Name = "cid"
		cookie.Value = strconv.Itoa(cookieNum)
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookieNum++
		c.SetCookie(cookie)

		return c.String(http.StatusOK, "write a cookie")
	}
	return c.String(http.StatusOK, "cookie already exists")
}
