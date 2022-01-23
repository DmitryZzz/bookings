package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("must be ok, but got invalid")
	}

	m := make(map[string][]string)
	m["one"] = []string{"error"}
	form.Errors = m

	isValid = form.Valid()
	if isValid == true {
		t.Error("must have been error")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form is invalid, but fields are filled")
	}
}

func TestForm_Has(t *testing.T) {
	postedValues := url.Values{}
	form := New(postedValues)

	if form.Has("a") {
		t.Error("field wasn`t added, but no error")
	}

	postedValues = url.Values{}
	postedValues.Add("a", "a")

	form.Values = postedValues

	if !form.Has("a") {
		t.Error("not empty field provided, but got error")
	}
}

func TestForm_MinLength(t *testing.T) {
	postedValues := url.Values{}
	form := New(postedValues)

	form.MinLength("a", 3)
	if form.Valid(){
		t.Error("form shows min length for non-existent field")
	}

	isError := form.Errors.Get("a")
	if isError == "" {
		t.Error("should have an error, but didn`t get one")
	}

	postedValues = url.Values{}
	postedValues.Add("a", "123")
	form = New(postedValues)

	form.MinLength("a", 3)
	if !form.Valid() {
		t.Error("shows min length of 3 is not met  when it is")
	}

	isError = form.Errors.Get("a")
	if isError != "" {
		t.Error("should not have an error, but got one")
	}
}

func TestForm_IsEmail(t *testing.T) {
	postedValues := url.Values{}
	form := New(postedValues)

	form.IsEmail("x")
	if form.Valid() {
		t.Error("form shows valid email for non-existent field")
	}

	postedValues = url.Values{}
	postedValues.Add("wrong_email", "e.mail")
	form = New(postedValues)
	form.IsEmail("wrong_email")
	if form.Valid() {
		t.Error("")
	}

	postedValues = url.Values{}
	postedValues.Add("email", "me@company.com")
	form = New(postedValues)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("got valid for invalid email address")
	}
}