package main

import (
	"fmt"
	"spider"
	"time"
)

func main() {
	//sp, err := spider.NewRuleSpider("http://www.ugirls.com", `img src=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)
	//sp, err := spider.NewRuleSpider("http://www.ugirls.com", `img src=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)
	sp, err := spider.NewRuleSpider("https://www.taotuba.net", `img data-original=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)
	//sp, err := spider.NewRuleSpider("http://www.mzitu.com", `img src=\"(?P<image>[[:word:]\./:&%$#\-_=]+\.jpg)\"`)
	if err != nil {
		fmt.Println(err.Error())
	}

	for i := range sp.Start() {
		<-time.Tick(time.Second * 2)
		fmt.Println("Get result: ", i)
		sp.Done <- i
	}
}
