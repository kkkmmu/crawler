package spider

import (
	"crypto/tls"
	"errors"
	"gopkg.in/redis.v5"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type RuleSpider struct {
	root          string
	rule          string
	hc            *http.Client
	re            *regexp.Regexp
	ltp           *TaskProducer
	ctp           *TaskProducer
	ltc           *TaskConsumer
	ctc           *TaskConsumer
	filter        Filter //@liwei: We only need one filter.
	linkGenerator LinkGenerator
	ratelimit     <-chan time.Time
}

type Task struct {
	task string
}

type ProduceFunction func(rs *RuleSpider, task string)
type TaskProducer struct {
	name    string
	client  *redis.Client
	produce ProduceFunction
}

type ConsumeFunction func(rs *RuleSpider)
type TaskConsumer struct {
	name    string
	client  *redis.Client
	consume ConsumeFunction
}

func NewRuleSpider(root, rule string) (*RuleSpider, error) {
	if root == "" || rule == "" {
		return nil, errors.New("Invalid url and rule")
	}

	re, err := regexp.Compile(rule)
	if err != nil {
		return nil, errors.New("Invalid rule!")
	}

	return &RuleSpider{
		root:      root,
		rule:      rule,
		re:        re,
		ratelimit: time.Tick(time.Millisecond * 100),
		hc: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		ltp: &TaskProducer{
			name: root + ":[LProducer]",
			client: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			produce: linkTaskProduce,
		},
		ltc: &TaskConsumer{
			name: root + ":[LConsumer]",
			client: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			consume: linkTaskConsume,
		},

		ctp: &TaskProducer{
			name: root + ":[CProducer]",
			client: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			produce: contentTaskProduce,
		},

		ctc: &TaskConsumer{
			name: root + ":[CConsumer]",
			client: redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}),
			consume: contentTaskConsume,
		},
		filter:        defaultFilter,
		linkGenerator: defaultLinkGenerator,
	}, nil
}

func linkTaskProduce(rs *RuleSpider, task string) {
	res, _ := rs.ltc.client.HExists(rs.linkCacheName(), task).Result()
	if res {
		log.Println("[link]: ", task, " + ", res, " Already in db")
		return
	}
	<-rs.ratelimit
	log.Println("Add : ", task, " into queue: ", rs.linkQueueName())
	rs.ltp.client.RPush(rs.linkQueueName(), task)
	/* This should be done after get from working queue and all is success */
	if true {
		log.Println("Process done for link: ", task)
		rs.ctp.client.HMSet(rs.linkCacheName(), map[string]string{task: "1"})
	} else {
		//If failed, re-insert into link queue.
		rs.ltp.client.RPush(rs.linkQueueName(), task)
	}
}

func linkTaskConsume(rs *RuleSpider) {
	go func(rs *RuleSpider) {
		for {
			c, _ := rs.ltc.client.BLPop(time.Second*1000, rs.linkWorkQueueName()).Result()
			log.Println("Get task: ", c[1], " from link working queue")
			/* Should be put in Spider */
			resp, err := rs.hc.Get(c[1])
			if err != nil {
				log.Println("Error happened when get url: ", err.Error())
				continue
			}
			defer resp.Body.Close()

			document, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error happend when get reponse body: ", err.Error())
				continue
			}

			news, err := rs.linkGenerator(rs.root, string(document))
			if err != nil {
				log.Println("Error happened when fetch new link in page: ", rs.root)
				continue
			}

			log.Println("=+============================================+=")
			log.Println("News: ", news)
			log.Println("=+============================================+=")
			matches := rs.re.FindAllStringSubmatch(string(document), -1)
			log.Println("[content]: Find result: ", matches)
			for _, v := range matches {
				if rs.filter(v[1]) {
					continue
				}
				rs.ctp.produce(rs, v[1])
			}

			//	go func(newlink []string) {
			for _, l := range news {
				log.Println("[link]: produce new link: ", l)
				rs.ltp.produce(rs, l)
			}
			//			}(news)

		}
	}(rs)
	go func(rs *RuleSpider) {

		for {
			content, _ := rs.ltc.client.BLPop(time.Second*1000, rs.linkQueueName()).Result()
			log.Println("Move task: ", content[1], " into link working queue: ", rs.linkWorkQueueName())
			rs.ltc.client.RPush(rs.linkWorkQueueName(), content[1])
		}
	}(rs)

}

func contentTaskProduce(rs *RuleSpider, task string) {
	res, _ := rs.ltc.client.HExists(rs.contentCacheName(), task).Result()
	if res {
		log.Println("[content]: ", res, " Already in content db")
		return
	}

	log.Println("Add content: ", task, " into content queue: ", rs.contentQueueName())

	rs.ltp.client.RPush(rs.contentQueueName(), task)

	//This part should be set when all the operation is successful on a content
	if true {
		rs.ctp.client.HMSet(rs.contentCacheName(), map[string]string{task: "1"})
		log.Println("Process done for link: ", task)
	} else {
		//If failed, re-insert into link queue.
		rs.ltp.client.RPush(rs.contentQueueName(), task)
		log.Println("Process failed for link: ", task)
	}
}

func contentTaskConsume(rs *RuleSpider) {
	go func(rs *RuleSpider) {
		for {
			content, _ := rs.ctp.client.BLPop(time.Second*1000, rs.contentQueueName()).Result()
			log.Println("Add content: ", content[1], " into content work queue")
			rs.ctp.client.RPush(rs.contentWorkQueueName(), content[1])
		}
	}(rs)

	go func(rs *RuleSpider) {
		for {
			content, _ := rs.ctp.client.BLPop(time.Second*1000, rs.contentWorkQueueName()).Result()
			log.Println("Get content: ", content[1], " from content working Queue: ", rs.contentWorkQueueName())
			if true {
				rs.ctp.client.HMSet(rs.contentCacheName(), map[string]string{content[1]: "1"})
				log.Println("Process done for content: ", content[1])
			} else {
				//If failed, re-insert into link queue.
				rs.ltp.client.RPush(rs.contentQueueName(), content[1])
				log.Println("Process failed for content: ", content[1])
			}
		}
	}(rs)
}

func (rs *RuleSpider) Start() {
	rs.ltp.client.Del(rs.linkQueueName())
	rs.ltp.client.Del(rs.linkWorkQueueName())
	rs.ltp.client.Del(rs.linkCacheName())
	rs.ctp.client.Del(rs.contentQueueName())
	rs.ctp.client.Del(rs.contentWorkQueueName())
	rs.ctp.client.Del(rs.contentCacheName())

	rs.ltc.consume(rs)
	rs.ctc.consume(rs)
	rs.ltp.produce(rs, rs.root)
}

func (rs *RuleSpider) setLinkProduceHandler(tp ProduceFunction) {
	rs.ltp.produce = tp
}

func (rs *RuleSpider) setLinkConsumeHandler(tc ConsumeFunction) {
	rs.ltc.consume = tc
}

func (rs *RuleSpider) setContentProduceHandler(tp ProduceFunction) {
	rs.ltp.produce = tp
}

func (rs *RuleSpider) setContentConsumeHandler(tc ConsumeFunction) {
	rs.ltc.consume = tc
}

func (rs *RuleSpider) linkQueueName() string {
	return "LINK:" + rs.root + ":QUEUE"
}

func (rs *RuleSpider) linkWorkQueueName() string {
	return "LINK:" + rs.root + ":WORKQUEUE"
}

func (rs *RuleSpider) linkCacheName() string {
	return "LINK:" + rs.root + ":CACHE"
}

func (rs *RuleSpider) contentQueueName() string {
	return "CONTENT:" + rs.root + ":QUEUE"
}

func (rs *RuleSpider) contentWorkQueueName() string {
	return "CONTENT:" + rs.root + ":WORKQUEUE"
}

func (rs *RuleSpider) contentCacheName() string {
	return "CONTENT:" + rs.root + ":CACHE"
}

func (rs *RuleSpider) RegisterFilter(filter Filter) {
	rs.filter = filter
}

func (rs *RuleSpider) RegisterLinkGenerator(generator LinkGenerator) {
	rs.linkGenerator = generator
}

func defaultFilter(in string) bool {
	if strings.HasSuffix(in, "css") || strings.HasSuffix(in, "js") || strings.HasSuffix(in, "asp") || strings.HasSuffix(in, "jsp") || strings.HasSuffix(in, "xml") {
		log.Println(" ", in, " is filtered by defaultFilter")
		return true
	}
	log.Println(in, " passed the default filter!")
	return false
}

func defaultLinkGenerator(page string, document string) ([]string, error) {
	re, err := regexp.Compile(`href=\"(?P<link>[[:word:]\-_#\$\^&=:\~/\.]+)\"`)
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
		if strings.HasPrefix(v[1], "http://") || strings.HasPrefix(v[1], "https://") {
			if !strings.HasSuffix(v[1], "js") && !strings.HasSuffix(v[1], "css") && !strings.HasSuffix(v[1], "jpg") && !strings.HasSuffix(v[1], "png") && !strings.HasSuffix(v[1], "gif") && !strings.HasSuffix(v[1], "jpeg") && !strings.HasSuffix(v[1], "xml") {
				/* We do not go out of this site */
				if strings.Contains(v[1], u.Scheme+"://"+u.Host) {
					links = append(links, v[1])
				}
			}
		}
	}

	return links, nil
}
