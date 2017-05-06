package main

import (
	"fmt"
	"net/url"
)

func main() {
	test, _ := url.Parse("http://www.liwei.com/test/help?t=1#b=2")
	fmt.Println(test.Scheme)
	fmt.Println(test.Host)
}
