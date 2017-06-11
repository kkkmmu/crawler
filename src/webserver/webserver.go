package webserver

import (
	//"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Web struct {
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/index.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func SliderPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/slider.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func ProjectPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/project.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func GalleryPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/gallery.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	data, err := ioutil.ReadFile("asset/web/txt/pics.txt")
	if err != nil {
		log.Println("Error happened when read file: ", err.Error())
		return
	}

	links := strings.Split(string(data), "\n")

	err = t.Execute(w, links[:1000])
	if err != nil {
		log.Println(err.Error())
	}
}

func WaterfallPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/waterfall.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	data, err := ioutil.ReadFile("asset/web/txt/pics.txt")
	if err != nil {
		log.Println("Error happened when read file: ", err.Error())
		return
	}

	links := strings.Split(string(data), "\n")

	err = t.Execute(w, links[:100])
	if err != nil {
		log.Println(err.Error())
	}
}

func LoadMorePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/loadmore.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func RWaterFallPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("asset/web/template/rwaterfall.html", "asset/web/template/footer.html", "asset/web/template/header.html")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func GetOneImagePage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Printf("%v", r.Form)
	cookie := r.Cookies()
	log.Printf("%v", cookie)

	log.Println(ImageSlice[rand.Intn(len(ImageSlice))])
	io.WriteString(w, ImageSlice[rand.Intn(len(ImageSlice))])
}

func Start() {
	http.HandleFunc("/index", MainPage)
	http.HandleFunc("/slider", SliderPage)
	http.HandleFunc("/project", ProjectPage)
	http.HandleFunc("/gallery", GalleryPage)
	http.HandleFunc("/waterfall", WaterfallPage)
	http.HandleFunc("/loadmore", LoadMorePage)
	http.HandleFunc("/getoneimage", GetOneImagePage)
	http.HandleFunc("/rwaterfall", RWaterFallPage)
	http.HandleFunc("/", MainPage)
	http.Handle("/asset/web/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8080", nil)
}

var ImageSlice []string

func init() {
	rand.Seed(time.Now().UnixNano())

	data, err := ioutil.ReadFile("asset/web/txt/pics.txt")
	if err != nil {
		log.Println("Error happened when read file: ", err.Error())
		return
	}

	ImageSlice = strings.Split(string(data), "\n")
}
