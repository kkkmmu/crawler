package spider

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

const (
	UNPROCESSED = iota
	PROCESSED
	PROCESSFAILED
)

type RuleSpider struct {
	url           string
	pages         chan string
	document      string
	rule          string
	re            *regexp.Regexp
	filter        Filter //@liwei: We only need one filter.
	linkGenerator LinkGenerator
	db            map[string]int
	dbLock        *sync.RWMutex
	Done          chan string
	result        chan string
}

func defualtFilter(in string) bool {
	if strings.HasSuffix(in, "css") || strings.HasSuffix(in, "js") || strings.HasSuffix(in, "asp") || strings.HasSuffix(in, "jsp") {
		fmt.Println(" ", in, " is filtered by defaultFilter")
		return true
	}
	fmt.Println(in, " passed the default filter!")
	return false
}

func defaultLinkGenerator(page string, document string) ([]string, error) {
	re, err := regexp.Compile(`href=\"(?P<link>[[:word:]\-_#\$\^&=:\~/\.]+)\"`)
	if err != nil {
		fmt.Println("Invalid regexp for fetch link")
		return nil, errors.New("Invalid regexp for fetch link")
	}
	matches := re.FindAllStringSubmatch(document, -1)
	links := make([]string, 0, len(matches))

	u, err := url.Parse(page)
	if err != nil {
		fmt.Println(" Error happened when paresing: ", page)
		return nil, errors.New("Invalid page url")
	}

	for _, v := range matches {
		if strings.HasPrefix(v[1], "http://") || strings.HasPrefix(v[1], "https://") {
			if !strings.HasSuffix(v[1], "js") && !strings.HasSuffix(v[1], "css") && !strings.HasSuffix(v[1], "jpg") && !strings.HasSuffix(v[1], "png") && !strings.HasSuffix(v[1], "gif") && !strings.HasSuffix(v[1], "jpeg") {
				/* We do not go out of this site */
				if strings.Contains(v[1], u.Scheme+"://"+u.Host) {
					links = append(links, v[1])
				}
			}
		}
		/*
			else {
				u, err := url.Parse(page)
				if err != nil {
					fmt.Println(" Error happened when paresing: ", page)
					continue
				}

				links = append(links, u.Scheme+"://"+u.Host+"/"+v[1])
			}
		*/
	}

	//fmt.Println("Find new links: ", links)
	return links, nil
}

func NewRuleSpider(url string, rule string) (*RuleSpider, error) {
	if url == "" || rule == "" {
		return nil, errors.New("Invalid url and rule")
	}

	re, err := regexp.Compile(rule)
	if err != nil {
		return nil, errors.New("Invalid rule!")
	}

	return &RuleSpider{
		url:           url,
		rule:          rule,
		re:            re,
		filter:        defualtFilter,
		linkGenerator: defaultLinkGenerator,
		pages:         make(chan string, 2),
		db:            make(map[string]int, 1000000), //@liwei: How to make this more flaxiable.
		dbLock:        &sync.RWMutex{},
		Done:          make(chan string, 2),
		result:        make(chan string),
	}, nil
}

func (rs *RuleSpider) Spide(page string) {
	/* Should be put in Spider */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(page)
	if err != nil {
		fmt.Println("Error happened when get url: ", err.Error())
		return
	}
	defer resp.Body.Close()

	document, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error happend when get reponse body: ", err.Error())
		return
	}

	rs.document = string(document)

	news, err := rs.linkGenerator(rs.url, rs.document)
	if err != nil {
		fmt.Println("Error happened when fetch new link in page: ", rs.url)
		return
	}

	//@liwei: Need Lock
	go func(newlink []string) {
		for _, l := range newlink {
			//fmt.Println("Add new link ", l, " into DB")
			rs.dbLock.Lock()
			_, ok := rs.db[l]
			if !ok {
				//	fmt.Println("Link ", l, " is not in the db")
				fmt.Println("Register New Link: ", l, " into db, total count: ", len(rs.db))
				rs.db[l] = UNPROCESSED
				rs.dbLock.Unlock()
				rs.pages <- l
				continue
			}
			//fmt.Println("Link ", l, " is already in the db")
			rs.dbLock.Unlock()
		}
	}(news)

	matches := rs.re.FindAllStringSubmatch(rs.document, -1)
	raw := make(chan string)
	go func(raw chan string) {
		for _, v := range matches {
			raw <- v[1]
		}
	}(raw)

	rs.Filter(raw)
}

func (rs *RuleSpider) Filter(in chan string) {
	go func(in chan string) {
		for match := range in {
			if rs.filter(match) {
				continue
			}
			rs.dbLock.Lock()
			if _, ok := rs.db[match]; ok {
				rs.dbLock.Unlock()
				continue
			}

			fmt.Println("Register New Link: ", match, " into db, total count: ", len(rs.db))
			rs.db[match] = UNPROCESSED
			rs.dbLock.Unlock()
			rs.result <- match
		}
	}(in)
}

func (rs *RuleSpider) Start() chan string {
	go func(pages chan string) {
		for p := range pages {
			rs.dbLock.Lock()
			state, ok := rs.db[p]
			rs.dbLock.Unlock()
			if ok {
				if state != PROCESSED {
					go rs.Spide(p)
				}
			} else {
				fmt.Println("Received link that is not in the db: ", p)
			}
		}
	}(rs.pages)

	go func(newlink []string) {
		for _, l := range newlink {
			rs.dbLock.Lock()
			_, ok := rs.db[l]
			if !ok {
				fmt.Println("Register New Link: ", l, " into db, total count: ", len(rs.db))
				rs.db[l] = UNPROCESSED
				rs.dbLock.Unlock()
				rs.pages <- l
				continue
			}
			rs.dbLock.Unlock()
		}
	}([]string{rs.url})

	go func() {
		for l := range rs.Done {
			rs.dbLock.Lock()
			_, ok := rs.db[l]
			if ok {
				rs.db[l] = PROCESSED
				fmt.Println("Process done for link: ", l)
			} else {
				fmt.Println("Received notification for unknown link: ", l)
			}
			rs.dbLock.Unlock()
		}
	}()
	return rs.result
}

func (rs *RuleSpider) RegisterFilter(filter Filter) {
	rs.filter = filter
}

func (rs *RuleSpider) RegisterLinkGenerator(generator LinkGenerator) {
	rs.linkGenerator = generator
}
