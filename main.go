package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/lithammer/shortuuid"
)

// sqlite with datasette can probably provide visualisations
type entry struct {
	url    string
	code   string
	mode   string // exact/sub
	count  int    // number of times we redirected
	scount int    // no of times we shortened
}
type createRequest struct {
	Url string `json:"url"`
}

var BASE_URL = "http://localhost:8088"
var db = make(map[string]entry)    // shortcode to full info
var rmap = make(map[string]string) // reverse map from url to shortcode

// separate func so that we can abstract to sqlite
func saveEntry(e entry) {
	db[e.code] = e
	rmap[e.url] = e.code
}
func getExisting(url string) (entry, bool) {
	code, ok := rmap[url]
	if ok {
		e, ok := db[code]
		if ok {
			bumpScount(code)
			return e, true
		}
		fmt.Printf("rmap available for '%v' but no db entry, ignoring rmap\n", url)
	}
	return entry{}, false
}
func bumpScount(code string) {
	e, ok := db[code]
	if ok {
		e.scount += 1
		db[code] = e
	}
}
func bumpCount(code string) {
	e, ok := db[code]
	if ok {
		e.count += 1
		db[code] = e
	}
}

func genCode(url string) string {
	e, ok := getExisting(url)
	if ok {
		return e.code
	}

	// we don't already have it, create new
	u := shortuuid.New()[:7]
	for {
		_, ok := db[u]
		if ok {
			// shorturl already used, create new
			u = shortuuid.New()[:5]
		} else {
			break
		}
	}
	saveEntry(entry{url: url, code: u, mode: "exact", count: 0, scount: 1})
	return u
}

func create(w http.ResponseWriter, r *http.Request) {
	p, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
	}

	var d createRequest
	err = json.Unmarshal(p, &d)
	if err != nil {
		http.Error(w, "Unable to get url", http.StatusBadRequest)
	}

	code := genCode(d.Url)
	fmt.Println("000:", d.Url, "->", code)
	fmt.Fprint(w, BASE_URL+"/"+code)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	entry, ok := db[code]
	if ok {
		fmt.Println("307:", code, "->", entry.url)
		bumpCount(code)
		http.Redirect(w, r, entry.url, http.StatusTemporaryRedirect)
	} else {
		fmt.Println("404:", code)
		http.NotFound(w, r)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		create(w, r)
	case "GET":
		redirect(w, r)
	default:
		http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8088"
	} else {
		port = ":" + port
	}
	fmt.Println("Starting server on", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
