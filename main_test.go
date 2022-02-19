package main

import (
	"testing"
)

func cleanDb() {
	db = make(map[string]entry)
	rmap = make(map[string]string)
}

func TestGenShortcode(t *testing.T) {
	cleanDb()
	url := "https://domain.tld"
	code := genCode(url)
	if len(code) == 0 {
		t.Error("got empty shortcode")
	}

	if db[code].scount != 1 {
		t.Errorf("invalid scount; expected '%v', got '%v'", 1, db[code].scount)
	}
	if db[code].count != 0 {
		t.Errorf("invalid count; expected '%v', got '%v'", 0, db[code].count)
	}
	if db[code].url != url {
		t.Errorf("invalid url; expected '%v', got '%v'", url, db[code].url)
	}
	if db[code].code != code {
		t.Errorf("invalid code; expected '%v', got '%v'", code, db[code].code)
	}
}

func TestGenShortcodeDuplicate(t *testing.T) {
	cleanDb()

	url := "https://domain.tld"
	code := genCode(url)
	if len(code) == 0 {
		t.Error("got empty shortcode")
	}

	ncode := genCode(url)
	if code != ncode {
		t.Error("got different codes on subsequent calls")
	}

	if db[code].scount != 2 {
		t.Errorf("invalid scount; expected '%v', got '%v'", 2, db[code].scount)
	}
}
