package forms

import (
	"net/http/httptest"
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
	//var f Form
}
