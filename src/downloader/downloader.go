package downloader

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Job struct {
	link string
	name string
	err  error
}

type ResultHandler func(success, failed <-chan Job)
type LinkVerifier func(link string) string

type Downloader struct {
	root          string
	client        *http.Client
	ratelimit     <-chan time.Time
	userAgent     string
	done          chan bool
	queue         chan Job
	failed        chan Job
	success       chan Job
	linkVerifier  LinkVerifier
	resultHandler ResultHandler
}

func NewDownloader(root string) (*Downloader, error) {
	fi, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(root, 0644)
			return nil, err
		}
		return nil, err
	}

	if !fi.IsDir() {
		return nil, errors.New(root + " exist but is not a directory!")
	}

	if err := ioutil.WriteFile(root+"/test.txt", []byte("Hello world"), 644); err != nil {
		return nil, err
	}

	os.Remove(root + "/test.txt")

	d := &Downloader{
		root: root,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		ratelimit:     time.Tick(time.Second * 1),
		done:          make(chan bool),
		queue:         make(chan Job, 100),
		failed:        make(chan Job),
		success:       make(chan Job),
		resultHandler: defaultResultHandler,
		linkVerifier:  defaultLinkVerifier,
		userAgent:     "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36",
	}

	go d.handleResult()
	go d.start()

	return d, nil
}

func (d *Downloader) Done() <-chan bool {
	return d.done
}

func (d *Downloader) start() {
	for _ = range d.ratelimit {
		for job := range d.queue {
			go d.download(job.link, job.name)
		}
	}
}

func (d *Downloader) handleResult() {
	d.resultHandler(d.success, d.failed)
}

func (d *Downloader) Download(link, name string) {
	d.queue <- Job{link: d.linkVerifier(link), name: name}
}

func (d *Downloader) download(link, name string) {
	if name == "" {
		name = generateNameFromLink(link)
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		d.failed <- Job{link: link, name: name, err: errors.New("Cannot open new request: " + err.Error())}
		return
	}

	req.Header.Add("User-Agent", d.userAgent)
	resp, err := d.client.Do(req)
	if err != nil {
		d.failed <- Job{link: link, name: name, err: errors.New("Cannot open the link: " + err.Error())}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		d.failed <- Job{link: link, name: name, err: errors.New("Response status code error: " + resp.Status)}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		d.failed <- Job{link: link, name: name, err: errors.New("Cannot get response body: " + err.Error())}
		return
	}

	if err := ioutil.WriteFile(d.root+"/"+name, body, 0644); err != nil {
		d.failed <- Job{link: link, name: name, err: errors.New("Cannot create new file: " + err.Error())}
	}

	d.success <- Job{link: link, name: name, err: nil}
}

func (d *Downloader) SetReusltHandler(handler ResultHandler) {
	d.resultHandler = handler
}

func generateNameFromLink(link string) string {
	name := strings.Replace(link, "/", "", -1)
	name = strings.Replace(name, "\\", "", -1)
	name = strings.Replace(name, "\n", "", -1)
	return name
}

func defaultResultHandler(success, failed <-chan Job) {
	for {
		select {
		case job := <-success:
			log.Println("[SUCCESS]: ", job.link)
		case job := <-failed:
			log.Println("[FAILED ]: ", job.link, " reason: ", job.err.Error())
		}
	}
}

func defaultLinkVerifier(link string) string {
	if !strings.HasPrefix(link, "http") {
		return "http://" + link
	}
	return link
}
