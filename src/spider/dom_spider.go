package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"sync"
)

type Spider struct {
	doc       *goquery.Document
	root      string //root
	selector  string //Please refer to the CSS selector docuemnt to get the right selector
	attribute string
	client    *http.Client
}

func CreateNewSpider(root, selector, attribute string) (*Spider, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(root)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Cannot Open page: " + root)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("Cannot create Docuement by Response")
	}

	return &Spider{
		doc:       doc,
		root:      root,
		client:    client,
		selector:  selector,
		attribute: attribute,
	}, nil
}

func (s *Spider) GetHtml(rule string) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	s.doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			content, _ := sl.Html()
			mu.Lock()
			res = append(res, content)
			mu.Unlock()

		}()
	})

	wg.Wait()
	return res, nil
}

func (s *Spider) GetText(rule string) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	s.doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			res = append(res, sl.Text())
			mu.Unlock()
		}()
	})
	wg.Wait()
	return res, nil
}

func (s *Spider) GetAttr(rule, attr string) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	s.doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
		s, _ := sl.Html()
		log.Println("Matched: ", s)
		wg.Add(1)
		go func() {
			defer wg.Done()
			attr, ok := sl.Attr(attr)
			if ok {
				mu.Lock()
				res = append(res, attr)
				mu.Unlock()
			}
		}()
	})
	wg.Wait()
	return res, nil
}

func (s *Spider) Start() {
	log.Println(s.selector, " ", s.attribute)
	log.Println(s.GetAttr(s.selector, s.attribute))
}

func main() {
	sp, err := CreateNewSpider("https://www.taotuba.net", "div.post-thumbnail>a", "href")
	if err != nil {
		log.Println(err.Error())
	}
	go sp.Start()

	/* Please refer to the CSS selector document to get the right selector*/
	//sp1, err := CreateNewSpider("https://www.taotuba.net/huayan/6139.html", "div.picsbox>p.img_jz>a", "href")
	//sp1, err := CreateNewSpider("https://www.taotuba.net/huayan/6139.html", "div.picsboxcenter>p.img_jz>a>img", "src")
	sp1, err := CreateNewSpider("https://www.taotuba.net/huayan/6139.html", "div.picsboxcenter>p.img_jz>a>img", "src")
	if err != nil {
		log.Println(err.Error())
	}
	go sp1.Start()
	done := make(chan int)
	<-done
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
