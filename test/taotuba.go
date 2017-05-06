package main

import (
	"log"
	"spider"
)

func main() {
	//sp, err := CreateNewSpider("https://www.taotuba.net", "div.post-thumbnail>a", "href")
	//sp, err := spider.CreateNewDomSpider("http://www.lesmao.com", "#thread-pic>ul>li>img", "src")
	//sp, err := spider.CreateNewDomSpider("http://www.lesmao.com", "#thread-pic", "")
	sp, err := spider.CreateNewDomSpider("https://www.taotuba.net", "div.picsbox>p.img_jz>a>img", "src")
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
