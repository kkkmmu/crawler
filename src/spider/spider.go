package spider

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"
)

type Filter func(match string) bool

type ResultCleaner func(domain, result string) string

type LinkGenerator func(page string, document string) ([]string, error)

type Spider interface {
	Spide(page string)
	RegisterFilter(filter Filter)                  //@liwei: Should be change to setfilter
	RegisterLinkGenerator(generator LinkGenerator) //@liwei: Should be change to setfilter
	Filter(chan string) chan string
	Start() chan string
}

func defaultFilter(in string) bool {
	if strings.HasSuffix(in, "css") || strings.HasSuffix(in, "js") || strings.HasSuffix(in, "asp") || strings.HasSuffix(in, "jsp") || strings.HasSuffix(in, "xml") {
		log.Println(" ", in, " is filtered by defaultFilter")
		return true
	}
	log.Println(in, " passed the default filter!")
	return false
}

func defaultCleaner(root, result string) string {
	if strings.HasPrefix(result, "/") {
		return root + "/" + result
	}
	return result
}

func defaultLinkGenerator(page string, document string) ([]string, error) {
	re, err := regexp.Compile(`href=\"(?P<link>[[:word:]\-_#%\$\^&=:\~/\.\?]+)\"`)
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
		if strings.HasPrefix(link, "/") || strings.HasPrefix(link, "./") {
			link = u.Scheme + "://" + u.Host + link
		}
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			if !strings.Contains(link, "js") && !strings.Contains(link, "css") && !strings.Contains(link, "jpg") && !strings.Contains(link, "png") && !strings.Contains(link, "gif") && !strings.Contains(link, "jpeg") && !strings.Contains(link, "xml") && !strings.Contains(link, "less") && !strings.Contains(link, "php") && !strings.Contains(link, "aspx") {
				/* The root already processed. */
				if link != u.Scheme+"://"+u.Host && link != u.Scheme+"://"+u.Host+"/" {
					/* We do not go out of this site */
					if strings.Contains(link, u.Scheme+"://"+u.Host) {
						links = append(links, link)
					}
				}
			}
		}
	}

	return links, nil
}
