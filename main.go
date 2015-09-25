package main

import (
	"github.com/garyburd/redigo/redis"
	"net/http"
	"net/http/cookiejar"
)

const LOGIN_URL string = "https://dyeon.net/user/login"

var r = connectRedis()

func connectRedis() redis.Conn {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	return c
}

func login(email string, password string) {

}

func getPostList(idx int) {

}

func getPost(id string) bool {

}

func main() {

}
