package service

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"labix.org/v2/mgo"
	"log"
	"models"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func EchoServer(url string, num int, dburi string) websocket.Handler {
	return func(ws *websocket.Conn) {

		session, err := mgo.Dial(dburi)
		defer session.Close()

		if err != nil {
			panic(err)
		}

		session.SetMode(mgo.Monotonic, true)
		modelCollection := session.DB("meikong").C("model")
		//先删除所有的记录
		modelCollection.RemoveAll(nil)
		MeikongSpider(ws, url, num, modelCollection)
	}
}

func MeikongSpider(ws *websocket.Conn, url string, num int, modelCollection *mgo.Collection) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div.index-module.noborder-module>div.w970>ul.post.small-post").Each(func(i int, s *goquery.Selection) {
		name := s.Find("li>a.nickname").Text()
		name = strings.Replace(name, ".", "", -1)
		click := s.Find("li").Eq(2).Find("span").Text()
		job := s.Find("li").Eq(1).Find("span").Text()
		href, _ := s.Find("div.cover>a").Attr("href")
		num = num + 1
		fmt.Println("模特", num, ":", name)
		fmt.Println("职业:", job)
		fmt.Println("点击量:", click)
		url = "http://www.moko.cc" + href
		fmt.Println("个人主页:", url)

		clicknum, _ := strconv.Atoi(click)

		modelInfo := "模特" + strconv.Itoa(num) + "." + name + "<br/>职业:" + job + "<br/>点击量:" + click + "<br/>个人主页:" + url
		if err = websocket.Message.Send(ws, modelInfo); err != nil {
			log.Println("Can't send")
		}

		images := parseModelInfo(ws, url, name)

		err = modelCollection.Insert(&models.Model{Number: num, Name: name, Job: job, Click: clicknum, Page: url, Address: images})

		if err != nil {
			panic(err)
		} else {
			log.Println("-------------insert data for model", name, "success-------------------")
		}
	})

	node := doc.Find("p.page>a").Last()
	_, exists := node.Attr("hidefocus")
	if exists {
		next, _ := node.Attr("href")
		next = "http://www.moko.cc" + next
		MeikongSpider(ws, next, num, modelCollection)
	}
}

func parseModelInfo(ws *websocket.Conn, url string, name string) []string {

	//利用数组切片来存储图片，避免需要预先定义数组的大小
	var images []string

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	dir := "模特" + string(filepath.Separator) + name
	os.MkdirAll(dir, os.ModeDir)

	doc.Find("div.article>div.pic.dBd").Each(func(i int, s *goquery.Selection) {
		content, _ := s.Find("p.picBox").Html()
		re := regexp.MustCompile("src2=\"(\\S+)\"")
		src := re.FindStringSubmatch(content)[1]
		if strings.Contains(src, "?") {
			src = src[:strings.LastIndex(src, "?")]
		}

		images = append(images, src)
		image := src[strings.LastIndex(src, "/")+1:]
		saveFileFromUrl(src, dir, image)

		if err = websocket.Message.Send(ws, time.Now().Format("2006-01-02 15:04:05")+" create image <b>"+image+"</b> success!<br/><img src=\"模特/"+name+"/"+image+"\" style=\"height:200px;\"/>"); err != nil {
			log.Println("Can't send")
		}
	})

	return images
}

func saveFileFromUrl(url string, dir string, image string) {
	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}

	defer response.Body.Close()
	file, err := os.Create(dir + string(filepath.Separator) + image)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	log.Println("create image", image, "success!")
}
