// Based on the example by the Authrs of go
// This is a fun project in order to learn some go and htmx

package main

import (
	"html/template"
	"log"
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
	"os"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func (p *Page) delete() error {
	filename := p.Title + ".txt"
	return os.Remove(filename)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "root", nil)
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	var menuitems []string
	files, err := ioutil.ReadDir(".")
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			menuitems = append(menuitems, file.Name()[:len(file.Name())-4])
		}
	}
	err2 := templates.ExecuteTemplate(w, "menu.html", menuitems)
	if err2 != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: title, Body: nil}
	err := os.Remove(title + ".txt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "ok", p)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("name")
	if len(title) == 0 {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("ok.html", "edit.html", "view.html", "root.html", "menu.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|delete)/([a-zA-Z0-9 ]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	log.Println("main() startet")
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/new", newHandler)
	
	log.Println("handler initialized")
	log.Println(http.ListenAndServe(":8080", nil))
}
