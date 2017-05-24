package main

import (
	"log"
	"spider"
)

func main() {
	sp, err := spider.CreateNewDomSpider("http://www.icandygirls.com", "div.post-content>p>a>img", "src")
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
