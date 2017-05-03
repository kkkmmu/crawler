package main

import (
	"fmt"
	"spider"
)

func main() {
	sp, err := spider.NewRuleSpider("http://ditto.dasannetworks.com", `title=\"(?P<t>[[:word:]]+)\"`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for i := range sp.Spide() {
		fmt.Println(i)
	}
}
