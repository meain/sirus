package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var url = "https://domain.tld"

func setup() {
	user = ""
	pass = ""
	db = make(map[string]entry)
	rmap = make(map[string]string)
}

func TestGenShortcode(t *testing.T) {
	setup()
	code, _ := genCode(url, "", "")
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

func TestGenShortcodeSub(t *testing.T) {
	setup()
	code, _ := genCode(url, "", "sub")
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

func TestGenShortcodeCustom(t *testing.T) {
	setup()
	code, _ := genCode(url, "domain", "")
	if code != "domain" {
		t.Error("got incorrect shortcode for custom")
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

func TestGenShortcodeCustomSub(t *testing.T) {
	setup()
	code, _ := genCode(url, "domain", "sub")
	if code != "domain" {
		t.Error("got incorrect shortcode for custom")
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
	setup()

	url := "https://domain.tld"
	code, _ := genCode(url, "", "")
	if len(code) == 0 {
		t.Error("got empty shortcode")
	}

	ncode, _ := genCode(url, "", "")
	if code != ncode {
		t.Error("got different codes on subsequent calls")
	}

	if db[code].scount != 2 {
		t.Errorf("invalid scount; expected '%v', got '%v'", 2, db[code].scount)
	}
}

func TestSimpleGet(t *testing.T) {
	setup()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if got != "sirus ^)" {
		t.Fatal("Did not get response back from server")
	}
}

func TestCreateShort(t *testing.T) {
	setup()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if !strings.Contains(got, BASE_URL) {
		t.Fatalf("%v not found in response", BASE_URL)
	}
}

func TestCreateShortAuthenticated(t *testing.T) {
	setup()
	user = "user"
	pass = "pass"
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	if response.Result().StatusCode != http.StatusUnauthorized {
		t.Error("unautorized request allow to create shortened url")
	}

	request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`"}`)))
	request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
	response = httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if !strings.Contains(got, BASE_URL) {
		t.Fatalf("%v not found in response", BASE_URL)
	}
}

func TestCreateShortDuplicate(t *testing.T) {
	setup()
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
func TestCreateShortCustomDuplicate(t *testing.T) {
	setup()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`", "code": "domain"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`", "code": "domain"}`)))
	response = httptest.NewRecorder()
	handler(response, request)
	got2 := response.Body.String()

	if got != got2 {
		t.Fatalf("short url not same; expected '%v', got '%v'", got, got2)
	}
}

func TestCreateShortSubDuplicate(t *testing.T) {
	setup()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`", "code": "domain"}`)))
	response := httptest.NewRecorder()
	handler(response, request)

	request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`", "code": "domain", "mode": "sub"}`)))
	response = httptest.NewRecorder()
	handler(response, request)

	if response.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("url recreated with sub mode and same code")
	}
}

func TestRedirect(t *testing.T) {
	setup()
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

func TestRedirectSub(t *testing.T) {
	setup()
	request, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"url":"`+url+`", "code": "g", "mode": "sub"}`)))
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	splits := strings.Split(got, "/")
	code := splits[len(splits)-1]

	request, _ = http.NewRequest(http.MethodGet, "/"+code+"/out", nil)
	response = httptest.NewRecorder()
	handler(response, request)

	if response.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected redirect, got '%v'", response.Code)
	}

	loc, err := response.Result().Location()
	if err != nil {
		t.Fatal("no redirect location provided")
	}
	if loc.String() != url+"/out" {
		t.Errorf("incorrect redirect location; expected '%v', got '%v'", url+"/out", loc.String())
	}
}
