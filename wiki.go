package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var allTemplates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}

func renderTemplate(
	responseWriter http.ResponseWriter,
	templateFileName string,
	page *Page,
) {
	templatesError := allTemplates.ExecuteTemplate(
		responseWriter,
		templateFileName+".html",
		page,
	)
	if templatesError != nil {
		http.Error(
			responseWriter,
			templatesError.Error(),
			http.StatusInternalServerError,
		)
	}
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, readError := ioutil.ReadFile(filename)
	if readError != nil {
		return nil, readError
	}
	return &Page{Title: title, Body: body}, nil
}

func (page *Page) savePage() error {
	filename := page.Title + ".txt"
	return ioutil.WriteFile(filename, page.Body, 0600)
}

func makeHandler(
	handlerFunction func(http.ResponseWriter, *http.Request, string),
) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		textMatchsList := validPath.FindStringSubmatch(request.URL.Path)
		if textMatchsList == nil {
			http.NotFound(responseWriter, request)
			return
		}
		handlerFunction(responseWriter, request, textMatchsList[2])
	}
}

func viewHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	title string,
) {
	page, pageError := loadPage(title)
	if pageError != nil {
		http.Redirect(
			responseWriter,
			request,
			"/edit/"+title,
			http.StatusFound,
		)
		return
	}
	renderTemplate(responseWriter, "view", page)
}

func editHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	title string,
) {
	page, pageError := loadPage(title)
	if pageError != nil {
		page = &Page{Title: title}
	}
	renderTemplate(responseWriter, "edit", page)
}

func saveHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	title string,
) {
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	pageError := page.savePage()
	if pageError != nil {
		http.Error(responseWriter, pageError.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(
		responseWriter,
		request,
		"/view/"+title,
		http.StatusFound,
	)
}
