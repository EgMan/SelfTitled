package main

import (
	//"io"
	"log"
	"fmt"
	"time"
	"strconv"
	"net/http"
	"io/ioutil"
	"html/template"
)

//config vars
var port				= "8000"
var favicon 			= "./config/static/trash.ico"
var rootHtml			= "./config/root_page.html"
var mainHtml			= "./files/main_page.html"
var bioHtml				= "./files/bio_page.html"
var otherHtml			= "./files/other_page.html"
var resumeHtml			= "./files/resume_page.html"
var complainInterval	= 30*time.Second

type Page struct{
	domain			string 										//Subdomain registered to Page
	displayHTML		string 										//Html (if any) to display in root template wrapper
	handle			func(w http.ResponseWriter, r *http.Request)//Handler (if any) for extra functionality needed (before displayHTML is pushed)
}
type Templ_info struct {
	Location		string
	Complaint		string
}

var rootTemplate *template.Template
var comp_count int
var handlers []Page

func main() {
	tmpl, err := template.New("rootTemplate").ParseFiles(rootHtml)
	if catch(err, "Error locating rootHtml"){
		panic("Aborting startup")
	}
	rootTemplate=tmpl

	var dat []byte
	dat, err = ioutil.ReadFile("./config/count.dat")
	if catch(err, "Error finding count.dat"){
		panic("Aborting startup")
	}
	var temp int
	temp, err = strconv.Atoi(string(dat))
	if catch(err, "Error parsing count.dat"){
		panic("Aborting startup")
	}
	comp_count=temp
	log.Printf("Complaint count recovered: %d\n", comp_count)

	log.Printf("Attaching handlers...\n")
	handlers = []Page{
		Page{"/",				mainHtml,	nil},
		Page{"/home",			mainHtml,	nil},
		Page{"/bio",			bioHtml,	nil},
		Page{"/other",			otherHtml,	nil},
		Page{"/resume",			resumeHtml,	nil},
		Page{"/search",			"",			HandleSearch},
		Page{"/refresh",		mainHtml,	HandleRefresh},
		Page{"/complain",		"",			HandleComplain},
		Page{"/favicon.ico",	"",			HandleFavicon}}
	for _,tmp := range handlers {
		h:=tmp//so closure isn't global
		http.HandleFunc(h.domain, func(w http.ResponseWriter, r *http.Request){
			logAccess(r)
			if h.handle != nil {
				h.handle(w,r)
			}
			if h.displayHTML != "" {
				ExecTempl(w, h.displayHTML)
			}
		})
	}
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./config/static/"))))
	log.Printf("Server up and listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func HandleRefresh(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("rootTemplate").ParseFiles(rootHtml)
	if catch(err, "Error locating rootHtml"){
		return
	}
	rootTemplate=tmpl
}

func ExecTempl(w http.ResponseWriter, html string){
	err := rootTemplate.ExecuteTemplate(w, "rootTemplate", Templ_info{html, strconv.Itoa(comp_count)})
	if catch(err, "Error templating"){
		return
	}
}

func HandleComplain(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("wusscookie")
	if err != nil || c.Expires.After(time.Now()) {
		comp_count++
		saveComplaint(comp_count)
		log.Printf("Wow, peple are bitching.  Complain count: %d\n", comp_count)
		cooks := http.Cookie{Name: "wusscookie", Value: "Wow, don't look at my cookies. These are private...", Expires: time.Now().Add(complainInterval)}
		http.SetCookie(w, &cooks)
		fmt.Fprintf(w, "%d", comp_count)
	} else {
		fmt.Fprintf(w, "#%s", strconv.FormatFloat(c.Expires.Sub(time.Now()).Minutes(), 'E', -1, 64))
		log.Printf("Complaint Spam")
	}
}

func saveComplaint(count int){
	err := ioutil.WriteFile("./config/count.dat", []byte(strconv.Itoa(count)), 0644)
	if catch(err, "Error updating complaint number"){
		return
	}
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./config/search.json")
}

func HandleFavicon(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "./config/static/favicon.ico")
}

func logAccess(r *http.Request){
	log.Printf(": %s : %s\n", r.Method, r.URL)
}

func catch(err error, message string) bool{
	if err != nil {
		log.Printf("%s\n%s",message,err)
		return true
	}
	return false
}