package main

import (
	"flag"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
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

// parse
var startPage int
var nextPage bool = true
var enterLoop bool = false

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

func initRedis() {
	initRedisHash("post")
	initRedisHash("comment")
}

func initRedisHash(t string) {
	_, err := redisClient.Do("DEL", year+month+date+"_"+t+"Count")
	if err != nil {
		log.Panic(err)
	}
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

func getComment(link string) {
	color.Yellow("[!] Parse link : ", link)
	resp, _ := httpClient.Get(link)
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	c := doc.Find(".comments-container").Find(".comments.first")
	c.Find("li").Each(func(_ int, s *goquery.Selection) {
		// date extract
		d := s.Find(".date").Text()
		if strings.Contains(d, string(year+"-"+month+"-"+date)) {
			user, exists := s.Find(".name").Find("a").Attr("data-id")
			if exists {
				userCommentCount(user)
				setUserName(user, s.Find(".name").Find("a").First().Text())
			}
		}
	})
}

func getPostList(idx int) {
	color.Yellow("[!] Parse page #", idx)
	resp, _ := httpClient.Get(DYEON_BOARD_URL + "?page=" + strconv.Itoa(idx))
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	t := doc.Find("tbody").First()
	t.Find("tr").Not(".ad").Not(".notice").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		// date extract
		d := s.Find(".absolute").Text()
		if strings.Contains(d, year+"년") && strings.Contains(d, month+"월") && strings.Contains(d, date+"일") {
			enterLoop = true
			// link extract
			link, exists := s.Find(".subject").Find("a").Attr("href")
			if exists {
				if strings.HasPrefix(link, "http://") {
					getComment(link)
				}
			}
			// user extract
			user, exists := s.Find(".name span").Find("a").Attr("data-id")
			if exists {
				userWriteCount(user)
				// user name extract
				setUserName(user, s.Find(".name span").Find("a").Text())
			}
			return true
		} else {
			if len(d) <= 11 {
				return true
			} else {
				if enterLoop == false {
					return true
				}
				nextPage = false
				return false
			}
		}
	})
	if nextPage {
		getPostList(idx + 1)
	}
}

func userWriteCount(user string) {
	_, err := redisClient.Do("HINCRBY", year+month+date+"_postCount", user, 1)
	if err != nil {
		log.Panic(err)
	}
}

func userCommentCount(user string) {
	_, err := redisClient.Do("HINCRBY", year+month+date+"_commentCount", user, 1)
	if err != nil {
		log.Panic(err)
	}
}

func calcPoint() {

}

func setUserName(user string, name string) {
	_, err := redisClient.Do("HSET", year+month+date+"_user", user, name)
	if err != nil {
		log.Panic(err)
	}
}

func getUserName(user string) string {
	name, err := redis.String(redisClient.Do("HGET", year+month+date+"_user", user))
	if err != nil {
		log.Panic(err)
		return user
	} else {
		return name
	}
}

func getStatistics() {
	sortStatistics("post")
	sortStatistics("comment")
}

func sortStatistics(t string) {
	// TODO : study sort algorithm
	// sort : http://play.golang.org/p/SAYsU8U17P
	data, err := redis.IntMap(redisClient.Do("HGETALL", year+month+date+"_"+t+"Count"))
	f, err := os.OpenFile("./statistics/"+year+month+date+"_"+t+".das", os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	if err != nil {
		log.Println(err)
	} else {
		n := map[int][]string{}
		var a []int
		for k, v := range data {
			n[v] = append(n[v], k)
		}
		for k := range n {
			a = append(a, k)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(a)))
		for _, k := range a {
			for _, s := range n[k] {
				f.WriteString(getUserName(s) + "\n")
				f.WriteString(strconv.Itoa(k) + "\n")
			}
		}
	}
}

func main() {
	// args
	flag.StringVar(&username, "u", "", "dyeon id")
	flag.StringVar(&password, "p", "", "dyeon password")
	flag.StringVar(&year, "y", "", "target year")
	flag.StringVar(&month, "m", "", "target month")
	flag.StringVar(&date, "d", "", "target date")
	flag.IntVar(&startPage, "s", 1, "start page")
	flag.Parse()
	// check args
	if username == "" || password == "" || year == "" || month == "" || date == "" {
		color.Red("[!] Argument missing")
	} else {
		// init redis
		initRedis()
		// try login
		color.Yellow("[!] Login")
		if login() {
			// if Success
			color.Green("[!] Login : PASS")
			getPostList(startPage)
			getStatistics()
		} else {
			// if Fail
			color.Red("[!] Login : FAIL")
		}
	}
}
