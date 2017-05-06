package spider

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
)

const (
	UNPROCESSED = iota
	PROCESSED
	PROCESSFAILED
)

type RuleSpiderChan struct {
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

func NewRuleSpiderChan(url string, rule string) (*RuleSpiderChan, error) {
	if url == "" || rule == "" {
		return nil, errors.New("Invalid url and rule")
	}

	re, err := regexp.Compile(rule)
	if err != nil {
		return nil, errors.New("Invalid rule!")
	}

	return &RuleSpiderChan{
		url:           url,
		rule:          rule,
		re:            re,
		filter:        defaultFilter,
		linkGenerator: defaultLinkGenerator,
		pages:         make(chan string, 2),
		db:            make(map[string]int, 1000000), //@liwei: How to make this more flaxiable.
		dbLock:        &sync.RWMutex{},
		Done:          make(chan string, 2),
		result:        make(chan string),
	}, nil
}

func (rs *RuleSpiderChan) Spide(page string) {
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

func (rs *RuleSpiderChan) Filter(in chan string) {
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

func (rs *RuleSpiderChan) Start() chan string {
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

func (rs *RuleSpiderChan) RegisterFilter(filter Filter) {
	rs.filter = filter
}

func (rs *RuleSpiderChan) RegisterLinkGenerator(generator LinkGenerator) {
	rs.linkGenerator = generator
}
