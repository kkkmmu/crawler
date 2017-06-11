package spider

import (
	"bytes"
	"crypto/tls"
	//"errors"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/redis.v5"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

/*
The basic understanding for a spider is as this：
	First we use links to represent the graph(edge) of a website.
	Second we use content to information(leaf) that we want to get.

	For Each LINK/CONTENT we maintain three queue and a CACHE.
		1. XXXQUEUE: Save all the unprocessed XXX.
		2. XXXWORKQUEEU: Save all the under-processing XXX. We use this queue to implement rate limit.
		3. XXXFAILEDQUEUE: Save all the XXX that is failed during the process.
		4. XXXCACHE: Save all the processed XXX. For FAILED/UNPROCESSED entry, the value is "0", For SUCCESSED entry, the value is "1";
*/

type Task struct {
	task    string
	success bool
}

type DomSpider struct {
	root                   string //root
	domain                 string
	selector               string //Please refer to the CSS selector docuemnt to get the right selector
	attribute              string
	result                 chan string /* Channel which will give user client the result information. */
	done                   chan string
	filter                 Filter //@liwei: We only need one filter.
	cleaner                ResultCleaner
	linkGenerator          LinkGenerator
	linkPublisher          *Publisher
	linkFailedPublisher    *Publisher
	contentPublisher       *Publisher
	contentFailedPublisher *Publisher
	linkConsumer           *Consumer
	linkFailedConsumer     *Consumer
	contentConsumer        *Consumer
	contentFailedConsumer  *Consumer
	resultGenerator        *ResultGenerator
	linkConfirmChannel     chan *Task /* Channel for confirmation of successful link process. */
	resultConfirmChannel   chan *Task /* Channel which is used to get the user confirmation for a particular result. */
	ratelimit              <-chan time.Time
}

type Queue struct {
	Name string
}

type Cache struct {
	Name string
}

type Publisher struct {
	queue       *Queue
	cache       *Cache
	Name        string
	redisClient *redis.Client
}

func (p *Publisher) Publish(msg string) {
	if p.IsPublished(msg) {
		//log.Println("Msg ", msg, " already exist in processed db: ", p.cache.Name)
		return
	}
	//log.Println("Publisher ", p.Name, " published message: ", msg)
	p.redisClient.RPush(p.queue.Name, msg)
	p.redisClient.HMSet(p.cache.Name, map[string]string{msg: "0"})
}

func (p *Publisher) IsPublished(msg string) bool {
	res, _ := p.redisClient.HExists(p.cache.Name, msg).Result()
	return res
}

type Consumer struct {
	workQueue   *Queue
	cache       *Cache
	Name        string
	httpClient  *http.Client
	redisClient *redis.Client
}

func (c *Consumer) Consume(msg string) {
	if c.IsConsumed(msg) {
		//	log.Println("Msg ", msg, " already processed! db: ", c.cache.Name)
		return
	}
	//log.Println("Consumer ", c.Name, " consumed message: ", msg)
	c.redisClient.RPush(c.workQueue.Name, msg)
}

func (c *Consumer) IsConsumed(msg string) bool {
	res, _ := c.redisClient.HGet(c.cache.Name, msg).Result()
	if res == "1" {
		return true
	}
	return false
}

type ResultGenerator struct {
	cache       *Cache
	redisClient *redis.Client
}

func (rg *ResultGenerator) Generate(task string) string {
	_, err := rg.redisClient.HMSet(rg.cache.Name, map[string]string{task: "1"}).Result()
	if err != nil {
		log.Println("Cannot save: ", task, " with: ", err.Error())
	}

	return task
}

func CreateNewDomSpider(root, selector, attribute string) (*DomSpider, error) {
	u, err := url.Parse(root)
	if err != nil {
		log.Println(" Error happened when paresing: ", root)
		return nil, errors.New("Invalid page url")
	}

	domain := u.Scheme + "://" + u.Host

	return &DomSpider{
		root:                 root,
		domain:               domain,
		selector:             selector,
		attribute:            attribute,
		done:                 make(chan string),
		result:               make(chan string),
		linkConfirmChannel:   make(chan *Task, 10),
		resultConfirmChannel: make(chan *Task, 10),
		ratelimit:            time.Tick(time.Second * 1),
		filter:               defaultFilter,
		cleaner:              defaultCleaner,
		linkGenerator:        defaultLinkGenerator,
		resultGenerator: &ResultGenerator{
			cache: &Cache{
				Name: "SPIDER:RESULT:CACHE",
			},

			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
		},
		linkPublisher: &Publisher{
			Name: "LinkPublisher",
			queue: &Queue{
				Name: "LINK:" + root + ":QUEUE",
			},
			cache: &Cache{
				Name: "LINK:" + root + ":CACHE",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
		},
		linkFailedPublisher: &Publisher{
			Name: "LinkFailedPublisher",
			queue: &Queue{
				Name: "LINK:" + root + ":FAILEDQUEUE",
			},
			cache: &Cache{
				Name: "LINK:" + root + ":FAILEDCACHE",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
		},

		linkConsumer: &Consumer{
			Name: "LinkConsumer",
			cache: &Cache{
				Name: "LINK:" + root + ":CACHE",
			},
			workQueue: &Queue{
				Name: "LINK:" + root + ":WORKQUEUEU",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			httpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},

		linkFailedConsumer: &Consumer{
			Name: "LinkFailedConsumer",
			cache: &Cache{
				Name: "LINK:" + root + ":FAILEDCACHE",
			},
			workQueue: &Queue{
				Name: "LINK:" + root + ":FAILEDWORKQUEUEU",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			httpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},

		contentPublisher: &Publisher{
			Name: "contentPublisher",
			queue: &Queue{
				Name: "CONTENT:" + root + ":QUEUE",
			},
			cache: &Cache{
				Name: "CONTENT:" + root + ":CACHE",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
		},

		contentFailedPublisher: &Publisher{
			Name: "ContentFailedPublisher",
			queue: &Queue{
				Name: "CONTENT:" + root + ":FailedQUEUE",
			},
			cache: &Cache{
				Name: "CONTENT:" + root + ":FAILEDCACHE",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
		},

		contentConsumer: &Consumer{
			Name: "ContentConsumer",
			cache: &Cache{
				Name: "CONTENT:" + root + ":CACHE",
			},
			workQueue: &Queue{
				Name: "CONTENT:" + root + ":WORKQUEUEU",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			httpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},

		contentFailedConsumer: &Consumer{
			Name: "ContentFailedConsumer",
			cache: &Cache{
				Name: "CONTENT:" + root + ":FAILEDCACHE",
			},
			workQueue: &Queue{
				Name: "CONTENT:" + root + ":FAILEDWORKQUEUEU",
			},
			redisClient: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			httpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},
	}, nil
}

func (s *DomSpider) SetLinkGenerator(generator LinkGenerator) {
	s.linkGenerator = generator
}

func (s *DomSpider) SetFilter(filter Filter) {
	s.filter = filter
}

func (s *DomSpider) SetResultCleaner(cleaner ResultCleaner) {
	s.cleaner = cleaner
}

func (s *DomSpider) ResultAccepted(task string) {
	s.resultConfirmChannel <- &Task{task: task, success: true}
}

func (s *DomSpider) ResultRejected(task string) {
	s.resultConfirmChannel <- &Task{task: task, success: false}
}

func (s *DomSpider) linkAccepted(task string) {
	s.linkConfirmChannel <- &Task{task: task, success: true}
}

func (s *DomSpider) linkRejected(task string) {
	s.linkConfirmChannel <- &Task{task: task, success: false}
}

/*
func (s *DomSpider) GetHtml(page, rule string) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	resp, err := s.client.Get(page)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Cannot Open page: " + s.root)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("Cannot create Docuement by Response")
	}

	doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
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

func (s *DomSpider) GetText(page, rule string) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	resp, err := s.client.Get(page)
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

	doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
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

func (s *DomSpider) spide(doc *goquery.Document) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
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

//With this function we get all the content that we need from a particular page.
func (s *DomSpider) GetAttr(page, rule, attr string) ([]string, error) {
	var (
		res = make([]string, 0) //for leaf
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	resp, err := s.client.Get(page)
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

	doc.Find(rule).Each(func(ix int, sl *goquery.Selection) {
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
*/

func (s *DomSpider) Spide() <-chan string {
	//Thread 1 Link Consume and ratelimit thread
	go func(s *DomSpider) {
		for {
			link, err := s.linkPublisher.redisClient.BLPop(time.Second*10000, s.linkPublisher.queue.Name).Result()
			if err != nil {
				log.Println("Error happed when get  link from queue: ", err.Error())
				continue
			}
			//log.Println("[LINK]Get task: ", link[1], " from link queue")
			s.linkConsumer.Consume(link[1])
		}
	}(s)

	//Thread 2 Thread Content generation ratelimit thread
	go func(s *DomSpider) {
		for {
			content, err := s.contentPublisher.redisClient.BLPop(time.Second*10000, s.contentPublisher.queue.Name).Result()
			if err != nil {
				log.Println("Error happed when get  content from queue: ", err.Error())
				continue
			}
			//log.Println("[CONTENT]Get task: ", content[1], " from content queue")
			s.contentConsumer.Consume(content[1])
		}
	}(s)

	//Thread 3 Link Consume thread.
	go func(s *DomSpider) {
		for {
			link, err := s.linkConsumer.redisClient.BLPop(time.Second*10000, s.linkConsumer.workQueue.Name).Result()
			if err != nil {
				log.Println("Error happed when get link from working queue: ", err.Error())
				continue
			}
			//log.Println("[LINK]Get task: ", link[1], " from link working queue")
			// Process the link
			//May be we should create http client for both producer an consumer
			req, err := http.NewRequest("GET", link[1], nil)
			if err != nil {
				log.Println("Cannot create new request for: ", link[1])
				continue
			}
			req.Header.Add("UserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36")
			resp, err := s.linkConsumer.httpClient.Do(req)
			if err != nil {
				log.Println("Error happened when get url: ", link[1], " error: ", err.Error())
				continue
			}
			defer resp.Body.Close()

			//Get all new links
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error happend when get reponse body: ", err.Error())
				continue
			}
			//log.Println(string(body))

			news, err := s.linkGenerator(s.root, string(body))
			if err != nil {
				log.Println("Error happened when fetch new link in page: ", s.root)
				continue
			}

			//log.Println(news)
			go func(news []string) {
				for _, l := range news {
					//log.Println("[Link]: Produce new link: ", l)
					//<-s.ratelimit
					s.linkPublisher.Publish(l)
				}
			}(news)

			//After last step, body is currupt. So the following three function cannot work
			//doc, err := goquery.NewDocumentFromResponse(resp)
			//doc, err := goquery.NewDocumentFromReader(resp.Body)
			//doc, err := goquery.NewDocument(link[1])
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if err != nil {
				log.Println("Error happend when create docuement")
				continue
			}

			//This may make a good performance? Make a local copy of each doc.
			go func(doc *goquery.Document) {
				var wg sync.WaitGroup
				doc.Find(s.selector).Each(func(ix int, selection *goquery.Selection) {
					html, _ := selection.Html()
					//log.Println("Matched: ", sh, " sl.Attr: ", a)
					wg.Add(1)
					go func() {
						defer wg.Done()
						if s.attribute == "" {
							s.contentPublisher.Publish(html)
						} else {
							attr, ok := selection.Attr(s.attribute)
							if ok {
								s.contentPublisher.Publish(attr)
							}
						}
					}()
				})
				wg.Wait()
			}(doc)
			s.linkAccepted(link[1])
		}
	}(s)

	//Thread 4 User result generation thread
	go func(s *DomSpider) {
		for {
			//If this operation is long blocked, how about another operation which use the same client in different thread?
			content, err := s.contentConsumer.redisClient.BLPop(time.Second*10000, s.contentConsumer.workQueue.Name).Result()
			if err != nil {
				log.Println("Error happend when get content from workqueue: ", err.Error())
				continue
			}
			//log.Println("[Content]Get task: ", content[1], " from content working queue")
			res, err := s.resultGenerator.redisClient.HGet(s.resultGenerator.cache.Name, content[1]).Result()
			if res == "1" {
				continue
			}

			s.result <- s.resultGenerator.Generate(s.cleaner(s.domain, content[1]))
			// Procuess the content
		}
	}(s)

	//Thread 5: Confirmation check  thread
	go func(s *DomSpider) {
		for {
			select {
			case l := <-s.linkConfirmChannel:
				log.Println("Accepted link: ", l)
				if l.success {
					s.linkPublisher.redisClient.HMSet(s.linkPublisher.cache.Name, map[string]string{l.task: "1"})
				} else {
					s.linkFailedPublisher.Publish(l.task)
				}
			case c := <-s.resultConfirmChannel:
				if c.success {
					s.contentPublisher.redisClient.HMSet(s.contentPublisher.cache.Name, map[string]string{c.task: "1"})
				} else {
					s.contentFailedPublisher.Publish(c.task)
				}
			}
		}
	}(s)

	//Thread 6: Error entry handler and Finish check
	go func(s *DomSpider) {
		//Need more check for this logic, since that there always be case that all the queue is empty but the spid is not finished.
		//So How to confirm that all is done ?
		for _ = range time.Tick(time.Minute * 60) {
			if l, err := s.contentFailedConsumer.redisClient.LLen(s.contentFailedPublisher.queue.Name).Result(); err == nil && l > 50 {
				go func() {
					content, err := s.contentFailedConsumer.redisClient.BLPop(time.Second*60, s.contentFailedPublisher.queue.Name).Result()
					if err != nil {
						log.Println("Get content from Failed queue with error: ", err.Error())
						return
					}
					s.contentConsumer.Consume(content[1])
				}()

			}

			if l, err := s.linkFailedConsumer.redisClient.LLen(s.linkFailedPublisher.queue.Name).Result(); err == nil && l > 50 {
				go func() {
					link, err := s.linkFailedConsumer.redisClient.BLPop(time.Second*60, s.linkFailedPublisher.queue.Name).Result()
					if err != nil {
						log.Println("Get link from Failed queue with error: ", err.Error())
						return
					}
					s.linkConsumer.Consume(link[1])
				}()
			}

			//If All queue are empty, we think that we are finished.
			if len, _ := s.linkFailedConsumer.redisClient.LLen(s.linkPublisher.queue.Name).Result(); len == 0 {
				if len, _ := s.linkFailedConsumer.redisClient.LLen(s.linkFailedPublisher.queue.Name).Result(); len == 0 {
					if len, _ := s.linkFailedConsumer.redisClient.LLen(s.linkConsumer.workQueue.Name).Result(); len == 0 {
						if len, _ := s.contentFailedConsumer.redisClient.LLen(s.contentPublisher.queue.Name).Result(); len == 0 {
							if len, _ := s.contentFailedConsumer.redisClient.LLen(s.contentFailedPublisher.queue.Name).Result(); len == 0 {
								if len, _ := s.contentFailedConsumer.redisClient.LLen(s.contentConsumer.workQueue.Name).Result(); len == 0 {
									s.done <- "Done"
								}
							}
						}
					}
				}
			}
		}
	}(s)

	s.linkPublisher.Publish(s.root)
	return s.result
}

func (s *DomSpider) Start() <-chan string {
	return s.Spide()
}

func (s *DomSpider) Done() <-chan string {
	return s.done
}

func (s *DomSpider) Reset() {
	s.linkPublisher.redisClient.Del(s.linkPublisher.queue.Name)
	s.linkPublisher.redisClient.Del(s.linkPublisher.cache.Name)
	s.linkFailedPublisher.redisClient.Del(s.linkFailedPublisher.queue.Name)
	s.linkFailedPublisher.redisClient.Del(s.linkFailedPublisher.cache.Name)

	s.linkConsumer.redisClient.Del(s.linkConsumer.workQueue.Name)
	s.linkConsumer.redisClient.Del(s.linkConsumer.cache.Name)
	s.linkFailedConsumer.redisClient.Del(s.linkFailedConsumer.workQueue.Name)
	s.linkFailedConsumer.redisClient.Del(s.linkFailedConsumer.cache.Name)
	s.contentConsumer.redisClient.Del(s.contentConsumer.workQueue.Name)
	s.contentConsumer.redisClient.Del(s.contentConsumer.cache.Name)
	s.contentFailedConsumer.redisClient.Del(s.contentFailedConsumer.workQueue.Name)
	s.contentFailedConsumer.redisClient.Del(s.contentFailedConsumer.cache.Name)
}
