package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/lithammer/shortuuid"
)

// Example for sub mode
// /g is set to redirect to https://github.com/meain
// Now we can do :host/g/blog -> https://github.com/meain/blog

// sqlite with datasette can probably provide visualisations
type entry struct {
	url    string
	code   string
	mode   string // exact/sub
	count  int    // number of times we redirected
	scount int    // no of times we shortened
}
type createRequest struct {
	Url  string `json:"url"`
	Code string `json:"code"`
	Mode string `json:"mode"`
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

func genCode(url string, code string, mode string) (string, bool) {
	if len(mode) == 0 {
		mode = "exact"
	}

	e, ok := getExisting(url)
	if ok {
		if e.mode != mode {
			return "", false
		}
		return e.code, true
	}

	if len(code) == 0 {
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
		saveEntry(entry{url: url, code: u, mode: mode, count: 0, scount: 1})
		return u, true
	}

	saveEntry(entry{url: url, code: code, mode: mode, count: 0, scount: 1})
	return code, true
}

func create(w http.ResponseWriter, r *http.Request) {
	p, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	var d createRequest
	err = json.Unmarshal(p, &d)
	if err != nil {
		http.Error(w, "Unable to get url", http.StatusBadRequest)
		return
	}

	code, ok := genCode(d.Url, d.Code, d.Mode)
	if !ok {
		http.Error(w, "Short url not available", http.StatusBadRequest)
		return
	}
	fmt.Println("000:", d.Url, "->", code)
	fmt.Fprint(w, BASE_URL+"/"+code)
}

func getRedirectUrl(path string) (string, bool) {
	splits := strings.SplitN(path, "/", 2)
	code := splits[0]
	entry, ok := db[code]
	if !ok {
		return "", false
	}
	switch len(splits) {
	case 1:
		return entry.url, true
	case 2:
		url := strings.Join([]string{entry.url, splits[1]}, "/")
		return url, true
	}
	return "", false
}

func redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if len(path) == 0 {
		fmt.Fprint(w, "sirus ^)")
		return
	}
	url, ok := getRedirectUrl(path)
	if ok {
		fmt.Println("307:", path, "->", url)
		bumpCount(path)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else {
		fmt.Println("404:", path)
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
