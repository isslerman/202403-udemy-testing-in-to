package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_app_enableCORS(t *testing.T) {
	// dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	var tests = []struct {
		name         string
		method       string
		expectHeader bool
	}{
		{"preflight", "OPTIONS", true},
		{"get", "GET", false},
	}

	for _, e := range tests {
		handlerToTest := app.enableCORS(nextHandler)

		req := httptest.NewRequest(e.method, "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if e.expectHeader && rr.Result().Header.Get("Access-Control-Allow-Credentials") == "" {
			t.Errorf("%s: Access-Control-Allow-Origin header not set", e.name)
		}

		if !e.expectHeader && rr.Result().Header.Get("Access-Control-Allow-Credentials") != "" {
			t.Errorf("%s: Access-Control-Allow-Origin header not set", e.name)
		}
	}
}

func Test_app_authRequired(t *testing.T) {
	// dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	var tests = []struct {
		name             string
		token            string
		expectAuthorized bool
		setHeader        bool
	}{
		{name: "valid token", token: fmt.Sprintf("Bearer %s", tokens.Token), expectAuthorized: true, setHeader: true},
		{name: "no token", token: "", expectAuthorized: false, setHeader: true},
		{name: "invalid token", token: fmt.Sprintf("Bearer %s", expiredToken), expectAuthorized: false, setHeader: true},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		if e.setHeader {
			req.Header.Set("Authorization", e.token)
		}

		rr := httptest.NewRecorder()

		handlerToTest := app.authRequired(nextHandler)
		handlerToTest.ServeHTTP(rr, req)

		if e.expectAuthorized && rr.Code == http.StatusUnauthorized {
			t.Errorf("%s: got code 401, and should not have", e.name)
		}

		if !e.expectAuthorized && rr.Code != http.StatusUnauthorized {
			t.Errorf("%s: did not get code 401, and should have", e.name)
		}
	}
}
