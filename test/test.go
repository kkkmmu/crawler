package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

var title = regexp.MustCompile(`title=\"(?P<t>[[:word:]]+)\"`)

func main() {
	resp, err := http.Get("http://ditto.dasannetworks.com/")
	if err != nil {
		fmt.Println("Failed to open ditto")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to get Body")
	}

	fmt.Println(string(body))

	result := title.FindAllStringSubmatch(string(body), -1)
	fmt.Println(result)

	for _, v := range result {
		<-time.Tick(time.Second * 10)

		fmt.Println(v[1])
	}
}
