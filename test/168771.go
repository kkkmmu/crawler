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
	sp, err := spider.CreateNewDomSpider("http://www.168771.com/mztp", "div.show_content>p>img", "src")
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
		//log.Println(link)
		if strings.HasPrefix(link, "/") || strings.HasPrefix(link, "./") {
			link = u.Scheme + "://" + u.Host + link
		}

		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			if !strings.Contains(link, "js") && !strings.Contains(link, "css") && !strings.Contains(link, "jpg") && !strings.Contains(link, "png") && !strings.Contains(link, "gif") && !strings.Contains(link, "jpeg") && !strings.Contains(link, "xml") && !strings.Contains(link, "less") && !strings.Contains(link, "php") && !strings.Contains(link, "wp.") && !strings.Contains(link, "wp-json") {
				/* The root already processed. */
				if link != u.Scheme+"://"+u.Host && link != u.Scheme+"://"+u.Host+"/" {
					/* We do not go out of this site */
					if strings.Contains(link, u.Scheme+"://"+u.Host) {
						if strings.Contains(link, "mztp") && strings.HasSuffix(link, "html") {
							links = append(links, link)
						}
					}
				}
			}
		}
	}

	return links, nil
}
