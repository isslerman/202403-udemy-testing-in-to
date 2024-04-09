package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Has(t *testing.T) {
	form := NewForm(nil)

	has := form.Has("booboo")
	if has {
		t.Error("form shows has field when it shoud not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = NewForm(postedData)

	has = form.Has("a")
	if !has {
		t.Error("shows form does not have field when it should")
	}
}

func TestForm_Required(t *testing.T) {
	// create a request
	r := httptest.NewRequest("POST", "/whatever", nil)
	// get the parsed post data from request
	form := NewForm(r.PostForm)

	// set the required data
	form.Required("a", "b", "c")

	if form.Valid() {
		t.Error("form should not be valid when it has required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r, _ = http.NewRequest("POST", "/whatever", nil)
	r.PostForm = postedData

	form = NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form should be valid when it has all required fields")
	}
}

func TestForm_Check(t *testing.T) {
	form := NewForm(nil)

	form.Check(false, "password", "password is required")
	if form.Valid() {
		t.Error("Valid() returns false, and it should be true when calling Check()")
	}
}

func TestForm_ErrorGet(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")
	s := form.Errors.Get("password")
	if len(s) == 0 {
		t.Error("should have an error message and has not")
	}

	s = form.Errors.Get("whatever")
	if len(s) != 0 {
		t.Error("should not have and error, but do")
	}

}
