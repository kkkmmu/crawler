package main

import (
	"log"
	"spider"
	"strings"
)

func main() {
	sp, err := spider.CreateNewDomSpider("http://www.ddqsw.com", "div.img-holder>img", "src")
	if err != nil {
		log.Println(err.Error())
	}

	sp.Reset()
	sp.SetResultCleaner(Cleaner)
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

func Cleaner(root, result string) string {
	if strings.HasPrefix(result, "/") {
		result = root + "/" + result
	}

	if strings.Contains(result, "small") {
		result = strings.Replace(result, "small", "large", -1)
	}
	return result

}
