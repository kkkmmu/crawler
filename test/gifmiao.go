package main

import (
	"log"
	"spider"
)

func main() {
	//sp, err := spider.CreateNewDomSpider("http://www.gifmiao.com", "div.DB_imgSet>div.DB_imgWin>img", "src")
	sp, err := spider.CreateNewDomSpider("http://www.gifmiao.com", "div.ilike-tool>a.ilike-down", "href")
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
