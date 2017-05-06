package main

import (
	"log"
	"spider"
	"strings"
)

func main() {
	sp, err := spider.CreateNewDomSpider("http://www.qingdouke.com/pc/index.html", "div.scanpic>div.scanpic_r>ul>li>img", "src")
	if err != nil {
		log.Println(err.Error())
	}

	sp.Reset()
	sp.SetResultCleaner(resultCleaner)
	result := sp.Start()

	go func(r <-chan string) {
		for i := range r {
			log.Println("So finally we get the result: ", i)
		}
	}(result)

	log.Println(<-sp.Done())
}

func resultCleaner(domain, result string) string {
	c := result
	if strings.HasPrefix(result, "/") {
		c = domain + "/" + c
	}

	if strings.Contains(c, "600x900") {
		c = strings.Replace(c, "600x900", "2400x3600", -1)
	}

	return c
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
