package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lithammer/shortuuid"
)

// sqlite with datasette can probably provide visualisations
type entry struct {
	url       string
	shortcode string
	mode      string // exact/sub
	count     int
}

var db []entry

func addEntry(url, shortcode, mode string) {}
func updateCounter(shortcode string)       {}

func getShortcode() string {
	u := shortuuid.New()[:5]
	// TODO: check if we already have that id
	return u
}

func findEntry(shortcode string) (string, bool) {
	for _, val := range db {
		if val.shortcode == shortcode {
			return val.url, true
		}
	}
	return "", false
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported", http.StatusNotFound)
		return
	}
	url := r.URL.Path
	fullurl, found := findEntry(url[1:])
	if found {
		fmt.Println("307:", url[1:], "->", fullurl)
		http.Redirect(w, r, fullurl, 307)
	} else {
		fmt.Println("404:", url[1:])
		http.NotFound(w, r)
	}
}

func main() {
	fmt.Println("howdy!")
	fmt.Println(db)
	fmt.Println(getShortcode())
	db = append(db, entry{url: "https://meain.io", shortcode: "meain"})

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
