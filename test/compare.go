package main

import (
	"log"
	"strings"
)

func main() {
	log.Println(strings.Compare("Hello", "hello"))
	log.Println(strings.Compare("hello", "hello"))
	log.Println(("hello" == "hello"))
	log.Println(("Hello" != "hello"))
	log.Println(strings.HasPrefix("/hello", "/"))

	s := "http://image.qingdouke.com//upload/2017/125/48/3D2A4447-3_JKT3C_600x900.jpg"
	log.Println(strings.Contains("http://image.qingdouke.com//upload/2017/125/48/3D2A4447-3_JKT3C_600x900.jpg", "600x900"))
	log.Println(strings.Replace(s, "600x900", "2400x3600", 1))
}
