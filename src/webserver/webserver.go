package webserver

import (
	//"encoding/json"
	"gopkg.in/redis.v5"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Web struct {
	redisClient *redis.Client
	ImageSlice  []string
}

var web Web

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
	cookies := r.Cookies()
	log.Printf("%v", cookies)

	ss, err := r.Cookie("Seq")
	if err != nil {
		log.Println(err.Error())
		//This is stupid, should in the login page
		ss = &http.Cookie{Name: "Seq", Value: "-1", Path: "/getoneimage"}
	}

	s, err := strconv.Atoi(ss.Value)
	if err != nil {
		log.Println(err.Error())
	}

	cookie := &http.Cookie{
		Name:  "Seq",
		Value: strconv.Itoa(s + 1),
		Path:  "/getoneimage",
	}

	http.SetCookie(w, cookie)

	io.WriteString(w, web.ImageSlice[rand.Intn(len(web.ImageSlice))])
}

func VUE(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("vue.html").Delims("||", "||").ParseFiles("asset/web/template/vue.html", "asset/web/template/vuefooter.html", "asset/web/template/vueheader.html")
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

func VUEForm(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("vueform.html").Delims("||", "||").ParseFiles("asset/web/template/vueform.html", "asset/web/template/vuefooter.html", "asset/web/template/vueheader.html")
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

func Script(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("script.html").Delims("||", "||").ParseFiles("asset/web/template/script.html", "asset/web/template/vuefooter.html", "asset/web/template/vueheader.html")
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

func RunScript(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println(r.Method)
	for k, v := range r.Form {
		log.Println(k, v)
	}

	io.WriteString(w, "Success")
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
	http.HandleFunc("/vue", VUE)
	http.HandleFunc("/script", Script)
	http.HandleFunc("/runscript", RunScript)
	http.HandleFunc("/vueform", VUEForm)
	http.HandleFunc("/", MainPage)
	http.Handle("/asset/web/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8080", nil)
}

func init() {
	web.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	web.ImageSlice = make([]string, 0, 100000)

	rand.Seed(time.Now().UnixNano())

	res, err := web.redisClient.HGetAll("SPIDER:RESULT:CACHE").Result()
	if err == nil {
		log.Println("Get Images from DB")
		for k, _ := range res {
			web.ImageSlice = append(web.ImageSlice, k)
		}

		return
	}
	log.Println("Get Images from File")
	data, err := ioutil.ReadFile("asset/web/txt/pics.txt")
	if err != nil {
		panic(err)
	}
	web.ImageSlice = strings.Split(string(data), "\n")

	log.Println("Totally: ", len(web.ImageSlice), " images!")
}
