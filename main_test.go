package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var url = "https://domain.tld"

func cleanDb() {
	db = make(map[string]entry)
	rmap = make(map[string]string)
}

func TestGenShortcode(t *testing.T) {
	cleanDb()
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

func TestSimpleGet(t *testing.T) {
	cleanDb()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if got != "sirus ^)" {
		t.Fatal("Did not get response back from server")
	}
}

func TestCreateShort(t *testing.T) {
	cleanDb()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if !strings.Contains(got, BASE_URL) {
		t.Fatalf("%v not found in response", BASE_URL)
	}
}

func TestCreateShortDuplicate(t *testing.T) {
	cleanDb()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if !strings.Contains(got, BASE_URL) {
		t.Fatalf("%v not found in response", BASE_URL)
	}

	request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	response = httptest.NewRecorder()
	handler(response, request)
	got2 := response.Body.String()

	if got != got2 {
		t.Fatalf("short url not same; expected '%v', got '%v'", got, got2)
	}
}

func TestRedirect(t *testing.T) {
	cleanDb()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	splits := strings.Split(got, "/")
	code := splits[len(splits)-1]

	request, _ = http.NewRequest(http.MethodGet, "/"+code, nil)
	response = httptest.NewRecorder()
	handler(response, request)

	if response.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected redirect, got '%v'", response.Code)
	}

	loc, err := response.Result().Location()
	if err != nil {
		t.Fatal("no redirect location provided")
	}
	if loc.String() != url {
		t.Errorf("incorrect redirect location; expected '%v', got '%v'", url, loc.String())
	}
}
