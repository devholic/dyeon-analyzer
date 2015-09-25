package main

import (
	"flag"
	"github.com/fatih/color"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/html"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const DYEON_URL string = "https://dyeon.net"
const DYEON_BOARD_URL string = "http://dyeon.net/board"
const LOGIN_URL string = "https://dyeon.net/user/login"

// dyeon account
var username, password string

var httpClient = initHttpClient()
var redisClient = connectRedis()

func connectRedis() redis.Conn {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	return c
}

func initHttpClient() *http.Client {
	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: cookieJar,
	}
	return client
}

func login() bool {
	resp, _ := httpClient.PostForm(LOGIN_URL, url.Values{
		"user[login]":    {username},
		"user[password]": {password},
		"to":             {DYEON_URL},
	})
	if resp.StatusCode == 301 {
		return true
	} else {
		return false
	}
}

func postCheck(t html.Token) bool {
	for _, a := range t.Attr {
		if a.Key == "tag" {
			if a.Val == "V" {
				return true
			}
		}
	}
	return false
}

func getPostList(idx int) {
	resp, _ := httpClient.PostForm(LOGIN_URL, url.Values{
		"user[login]":    {username},
		"user[password]": {password},
		"to":             {DYEON_URL},
	})
	defer resp.Body.Close()
	// http://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
	t := html.NewTokenizer(resp.Body)
	for {
		n := t.Next()
		switch {
		case n == html.ErrorToken:
			return
		case n == html.StartTagToken:
			t := t.Token()
			isSpan := t.Data == "tag"
			if isSpan {
				color.Red("[!] List : Tag found")
				continue
			}
			if postCheck(t) {
			}
		}
	}
}

func getPost(id string) bool {
	return false
}

func main() {
	flag.StringVar(&username, "u", "", "dyeon id")
	flag.StringVar(&password, "p", "", "dyeon pw")
	flag.Parse()
	color.Yellow("[!] Login")
	l := login()
	if l {
		color.Green("[!] Login : PASS")
	} else {
		color.Red("[!] Login : FAIL")
	}
}
