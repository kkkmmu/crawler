package main

import (
	"log"
	"spider"
)

func main() {
	sp, err := spider.CreateNewDomSpider("http://www.169ee.com", "div.big-pic>div.big_img>p>img", "src")
	if err != nil {
		log.Println(err.Error())
	}

	sp.Reset()
	result := sp.Start()

	go func(r <-chan string) {
		for i := range r {
			log.Println("So finally we get the result: ", i)
		}
	}(result)

	log.Println(<-sp.Done())
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
