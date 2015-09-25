package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

const DYEON_URL string = "https://dyeon.net"
const DYEON_BOARD_URL string = "http://dyeon.net/board"
const LOGIN_URL string = "https://dyeon.net/user/login"

// dyeon account
var username, password string

// target information
var year, month, date string

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

func postCheck() {
}

func getPostList(idx int) {
	resp, _ := httpClient.Get(DYEON_BOARD_URL + "?page=" + strconv.Itoa(idx))
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	t := doc.Find("tbody").First()
	t.Find("tr").Not(".ad").Not(".notice").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Find(".subject").Find("a").Attr("href")
		if exists {
			if strings.HasPrefix(link, "http://") {
        fmt.Println(link)
			}
		}
	})
}

func getPost(id string) bool {
	return false
}

func main() {
	// args
	flag.StringVar(&username, "u", "", "dyeon id")
	flag.StringVar(&password, "p", "", "dyeon password")
	flag.StringVar(&year, "y", "", "target year")
	flag.StringVar(&month, "m", "", "target month")
	flag.StringVar(&date, "d", "", "target date")
	flag.Parse()

	// try login
	color.Yellow("[!] Login")
	if login() {
		// if Success
		color.Green("[!] Login : PASS")
		getPostList(1)
	} else {
		// if Fail
		color.Red("[!] Login : FAIL")
	}
}
