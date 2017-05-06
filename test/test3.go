package main

import (
	"log"
	"spider"
)

func main() {
	//sp, err := spider.NewRuleSpider("http://www.ugirls.com", `img src=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)
	//sp, err := spider.NewRuleSpider("http://www.ugirls.com", `img src=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//sp, err := spider.NewRuleSpider("https://www.taotuba.net", `img data-original=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)
	//sp, err := spider.NewRuleSpider("https://www.taotuba.net", `img src=\"(?P<image>https://img.taotuba.net/taotuba/[a-z]+/[0-9]+\.jpg)\"`)
	//sp, err := spider.NewRuleSpider("https://www.taotuba.net", `img src=\"(?P<image>https\:\/\/img\.taotuba\.net\/taotuba\/[[:world:]]+\/[[:world:]]+\.jpg)\"`)
	//sp, err := spider.NewRuleSpider("http://www.mzitu.com", `img src=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)

	sp, err := spider.NewRuleSpider("https://www.xiezhen.men", `img src=\"(?P<image>https://img.alicdn.com/imgextra/[[:word:]\![:space:]\/]+\.jpg)\"`)
	if err != nil {
		log.Println(err.Error())
	}

	sp.Start()
	done := make(chan string)
	<-done
}
