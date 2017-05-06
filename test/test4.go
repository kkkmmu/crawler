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
	/*
		sp, err := CreateNewSpider("https://www.taotuba.net", "div.post-thumbnail>a", "href")
		if err != nil {
			log.Println(err.Error())
		}
		go sp.Start()
	*/

	/* Please refer to the CSS selector document to get the right selector*/
	//sp1, err := CreateNewSpider("https://www.taotuba.net/huayan/6139.html", "div.picsbox>p.img_jz>a", "href")
	//sp1, err := CreateNewSpider("https://www.taotuba.net/huayan/6139.html", "div.picsboxcenter>p.img_jz>a>img", "src")
	//sp1, err := CreateNewSpider("https://www.taotuba.net/huayan/6139.html", "div.picsboxcenter>p.img_jz>a>img", "src")
	//sp1, err := CreateNewSpider("http://www.163.com", "a>img", "src")
	//sp1, err := CreateNewSpider("http://www.163.com", "a", "href")
	//sp1, err := CreateNewSpider("http://www.24meinv.me/2017/5-4/tuimo26621_16.html", "div#imgshow>a>img", "src")
	//sp1, err := CreateNewSpider("https://www.xiezhen.men/beautyleg-1440/", "article.article-content>p>img", "src")
	//sp1, err := CreateNewSpider("http://www.qingdouke.com/pc/album/948.html", "div.scanpic>div.scanpic_r>ul>li>img", "src")

	sp1, err := CreateNewSpider("http://m.xxxiao.com/71983/%E6%B8%85%E7%BA%AF%E5%8F%AF%E8%80%90-%E5%82%B2%E5%A8%87%E8%90%8C%E8%90%8C-%E5%BE%A1%E5%A5%B3%E9%83%8E%E5%A5%B3%E4%BB%86%E4%B8%80%E5%A4%A7%E7%89%87%E9%9B%AA%E7%99%BD%E7%BE%8E%E4%B8%8D%E8%83%9C-22", "div.entry-content>div>div>a>img", "src")
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
