package main

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"spider"
	"strings"
)

func main() {
	sp, err := spider.CreateNewDomSpider("http://www.ttttba.com", "div.context>div#post_content>p>img", "data-lazy-src")
	if err != nil {
		log.Println(err.Error())
	}

	sp.Reset()
	sp.SetLinkGenerator(defaultLinkGenerator)
	result := sp.Start()

	go func(r <-chan string) {
		for i := range r {
			log.Println("So finally we get the result: ", i)
		}
	}(result)

	log.Println(<-sp.Done())
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func defaultLinkGenerator(page string, document string) ([]string, error) {
	re, err := regexp.Compile(`href=\"(?P<link>[[:word:]\-_#\$%\^&=:\~/\.\?]+)\"`)
	if err != nil {
		log.Println("Invalid regexp for fetch link")
		return nil, errors.New("Invalid regexp for fetch link")
	}
	matches := re.FindAllStringSubmatch(document, -1)
	links := make([]string, 0, len(matches))

	u, err := url.Parse(page)
	if err != nil {
		log.Println(" Error happened when paresing: ", page)
		return nil, errors.New("Invalid page url")
	}

	for _, v := range matches {
		link := v[1]
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			if !strings.Contains(link, "js") && !strings.Contains(link, "css") && !strings.Contains(link, "jpg") && !strings.Contains(link, "png") && !strings.Contains(link, "gif") && !strings.Contains(link, "jpeg") && !strings.Contains(link, "xml") && !strings.Contains(link, "less") && !strings.Contains(link, "php") && !strings.Contains(link, "wp.") && !strings.Contains(link, "wp-json") && !strings.Contains(link, "javascript") && !strings.Contains(link, "comment") {
				/* The root already processed. */
				if link != u.Scheme+"://"+u.Host && link != u.Scheme+"://"+u.Host+"/" {
					if !strings.HasSuffix(link, "feed") && !strings.HasSuffix(link, "favorite") && !strings.HasSuffix(link, "nvshenhaoqiao") {
						/* We do not go out of this site */
						if strings.Contains(link, u.Scheme+"://"+u.Host) {
							//							log.Println(link)
							links = append(links, link)
						}
					}
				}
			}
		}
	}

	return links, nil
}
