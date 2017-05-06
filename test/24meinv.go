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
	sp, err := spider.CreateNewDomSpider("http://www.24meinv.me", "div#imgshow>a>img", "src")
	if err != nil {
		log.Println(err.Error())
	}

	sp.Reset()
	sp.SetResultCleaner(resultCleaner)
	sp.SetLinkGenerator(defaultLinkGenerator)
	result := sp.Start()

	go func(r <-chan string) {
		for i := range r {
			log.Println("So finally we get the result: ", i)
		}
	}(result)

	log.Println(<-sp.Done())
}

func resultCleaner(domain, result string) string {
	c := result
	if strings.HasPrefix(result, "/") {
		c = domain + "/" + c
	}

	if strings.Contains(c, "img") {
		c = strings.Replace(c, "img", "pic", -1)
	}

	return c
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func defaultLinkGenerator(page string, document string) ([]string, error) {
	re, err := regexp.Compile(`href=\"(?P<link>[[:word:]\-_#\$\^&=:\~/\.\?]+)\"`)
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
			if !strings.HasSuffix(link, "js") && !strings.HasSuffix(link, "css") && !strings.HasSuffix(link, "jpg") && !strings.HasSuffix(link, "png") && !strings.HasSuffix(link, "gif") && !strings.HasSuffix(link, "jpeg") && !strings.HasSuffix(link, "xml") && !strings.HasSuffix(link, "less") && !strings.HasSuffix(link, "php") {
				/* The root already processed. */
				if link != u.Scheme+"://"+u.Host && link != u.Scheme+"://"+u.Host+"/" {
					/* We do not go out of this site */
					if strings.Contains(link, u.Scheme+"://"+u.Host) && strings.HasSuffix(link, "html") {
						log.Println(link)
						links = append(links, link)
					}
				}
			}
		}
	}

	return links, nil
}
