package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from APO"))
	})
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handle %v %v", status, http.StatusOK)
	}
	expected := "Hello from API\n"
	if rr.Body.String() != expected {
		t.Errorf("handler %v %v", rr.Body.String(), expected)
	}
}
