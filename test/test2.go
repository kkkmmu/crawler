package main

import (
	"fmt"
	"spider"
)

func main() {
	sp, err := spider.NewRuleSpider("http://ditto.dasannetworks.com", `img src=\"(?P<img>[[:word:]/\-_\?\&=\.%\*\#:]+)\"`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for i := range sp.Spide() {
		fmt.Println(i)
	}
}
