package main

import (
	"bytes"
	"crypto/tls"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	//resp, err := http.Get("http://wiki.dasannetworks.com")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	//resp, err := client.Get("https://www.taotuba.net")
	resp, err := client.Get("https://www.taotuba.net/huayan/6104.html")
	if err != nil {
		log.Println(err.Error())
	}

	document, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
	}
	//log.Println(string(document))

	z := html.NewTokenizer(bytes.NewReader(document))

	log.Println(z.Text())
	log.Println(z.TagAttr())
	log.Println(z.TagName())
	log.Println(z.Raw())
	log.Println(z.Token())

	/*
		for {
			tt := z.Next()
			if tt == html.ErrorToken {
				continue
			}
			log.Println(tt)
			log.Println(tt.Type)
			log.Println(tt.Data)
			log.Println(tt.Attr)
			// Process the current token.
		}
	*/

	doc, err := html.Parse(bytes.NewReader(document))
	if err != nil {
		// ...
	}
	var f func(*html.Node)
	attribute := make(map[string]string, 1000)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				attribute[a.Key] = a.Val
			}
			if v, ok := attribute["class"]; ok {
				if v == "picsbox picsboxcenter" {
					log.Println("++++++++++++++++++++++++++++++++++++++++")
					log.Println(n)
					log.Println("----------------------------------------")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			log.Println(c)
			f(c)
		}
	}
	f(doc)
}
