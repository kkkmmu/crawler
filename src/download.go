package main

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	//"net/http/httputil"
	"os"
	"strings"
	"time"
)

var ROOT = "/mnt/hgfs/GOOD/pics/"

func main() {
	data, err := ioutil.ReadFile("pics.txt")
	if err != nil {
		log.Println("Error happened when read file: ", err.Error())
		return
	}
	d := NewDownloader(ROOT)
	sr := bytes.NewBufferString(string(data))
	d.Start()
	go func() {
		for {
			line, err := sr.ReadString('\n')
			if err != nil {
				log.Println("Error when readline: ", err.Error())
				return
			}
			//Read string will return the "\n", we must strip this when do http reqeuest"
			d.Download(line[:len(line)-1])
		}
	}()

	for {
		select {
		case <-d.done:
			log.Println("Download Finished")
		case link := <-d.success:
			log.Println("Download success: ", link)
		}
	}
}

type Downloader struct {
	root      string
	ratelimit <-chan time.Time
	client    *http.Client
	queue     chan string
	done      chan bool
	failed    chan string
	success   chan string
	userAgent string
}

func NewDownloader(root string) *Downloader {
	return &Downloader{
		root:      root,
		ratelimit: time.Tick(time.Millisecond * 100),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		queue:   make(chan string, 100),
		done:    make(chan bool),
		failed:  make(chan string),
		success: make(chan string),
		//	userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36",
		userAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36",
	}
}

func (d *Downloader) Download(link string) {
	d.queue <- d.verifyLink(link)
}

func (d *Downloader) verifyLink(link string) string {
	/*
		if !strings.HasPrefix(link, "http") {
			return "http://" + link
		}
	*/
	return link
}

func (d *Downloader) Start() {
	if _, err := os.Stat(d.root); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(d.root, 0666)
		}
	}
	go d.download()
}

func (d *Downloader) download() {
	for link := range d.queue {
		<-d.ratelimit
		go func(link string) {
			log.Println("Downloading: ", link)
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				log.Println("Create new reqeust error ", err.Error())
				return
			}
			req.Header.Add("User-Agent", d.userAgent)
			resp, err := d.client.Do(req)
			if err != nil {
				log.Println("Download ", link, " error: ", err.Error())
				return
			}

			//dump, _ := httputil.DumpRequestOut(req, true)

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Cannot get the response body")
				return
			}
			defer resp.Body.Close()

			name := d.root + strings.Replace(link, "/", "", -1)
			name = strings.Replace(name, ":", "", -1)
			name = strings.Replace(name, ".", "", -1)
			name = strings.Replace(name, "\n", "", -1)
			name = name + ".jpg"

			/*
				if err := ioutil.WriteFile(name, body, 0666); err != nil {
					log.Println("Cannot create new file for link: ", link, " error: ", err.Error())
					return
				}
			*/

			file, err := os.Create(name)
			if err != nil {
				log.Println("Cannot create new file for link: ", link, " error: ", err.Error())
				return
			}
			defer file.Close()
			file.Write(body)
			d.success <- link
		}(link)
	}
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}
