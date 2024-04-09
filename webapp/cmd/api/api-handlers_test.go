package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"webapp/pkg/data"

	"github.com/go-chi/chi/v5"
)

func Test_app_authenticate(t *testing.T) {
	var theTests = []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid user", `{"email": "admin@example.com","password":"secret"}`, http.StatusOK},
		{"not json", `oh my good`, http.StatusUnauthorized},
		{"empty json", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email": ""}`, http.StatusUnauthorized},
		{"empty password", `{"email": "admin@example.com","password":""}`, http.StatusUnauthorized},
		{"invalid user", `{"email": "admin@otherblabla.com","password":"secret"}`, http.StatusUnauthorized},
	}

	for _, e := range theTests {
		reader := strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)

		if e.expectedStatusCode != rr.Code {
			t.Errorf("%s: returned wrong status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_app_refresh(t *testing.T) {
	// Table test
	var tests = []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid", "", http.StatusOK, true},                                    // status 200
		{"expired token", expiredToken, http.StatusBadRequest, false},         // status 400
		{"valid but not yet ready to expire", "", http.StatusTooEarly, false}, // status 425

	}

	// valid user for test
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	// start the loop for tests
	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 1
			}
			token, _ := app.generateTokenPair(&testUser) // generate a tokenpair
			tkn = token.RefreshToken                     // get the refresh token
		} else {
			tkn = e.token
		}

		// create the posted data
		postedData := url.Values{
			"refresh_token": {tkn},
		}

		// create the post request using the postedData
		req, _ := http.NewRequest("POST", "/refresh-token", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // setting the header
		rr := httptest.NewRecorder()

		// setting the handler
		handler := http.HandlerFunc(app.refresh)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status of %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		refreshTokenExpiry = oldRefreshTime
	}
}

func Test_app_userHandlers(t *testing.T) {
	var tests = []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"allUsers", "GET", "", "", app.allUsers, http.StatusOK},
		{"deleteUser", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"deleteUser bad URL param", "DELETE", "", "Y", app.deleteUser, http.StatusBadRequest},
		{"getUser valid", "GET", "", "1", app.getUser, http.StatusOK},
		{"getUser invalid", "GET", "", "100", app.getUser, http.StatusBadRequest},
		{"getUser bad URL param", "GET", "", "Y", app.getUser, http.StatusBadRequest},
		{
			"updateUser valid",
			"PATCH",
			`{"id":1,"first_name":"Administrator","last_name":"User","email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusNoContent,
		},
		{
			"updateUser invalid",
			"PATCH",
			`{"id":100,"first_name":"Administrator","last_name":"User","email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"updateUser invalid json",
			"PATCH",
			`{"id":1,first_name:"Administrator","last_name":"User","email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"insertUser valid",
			"PUT",
			`{"first_name":"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusNoContent,
		},
		{
			"insertUser invalid",
			"PUT",
			`{"foo":"bar","first_name":"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
		{
			"insertUser invalid json",
			"PUT",
			`{first_name:"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
	}

	// start the loop for tests
	for _, e := range tests {
		var req *http.Request
		if e.json == "" {
			req, _ = http.NewRequest(e.method, "/", nil)
		} else {
			req, _ = http.NewRequest(e.method, "/", strings.NewReader(e.json))
		}

		if e.paramID != "" {
			fmt.Println("adding userID param,", e.paramID)
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(e.handler)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf("%s: wrong status returned; expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}

func Test_app_refreshUsingCookie(t *testing.T) {
	// valid user for test
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	testCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		Secure:   true,
		HttpOnly: true,
	}

	badCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    "somebadstring",
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		Secure:   true,
		HttpOnly: true,
	}

	var tests = []struct {
		name           string
		addCoodie      bool
		cookie         *http.Cookie
		expectedStatus int
	}{
		{"valid cookie", true, testCookie, http.StatusOK},
		{"invalid cookie", true, badCookie, http.StatusBadRequest},
		{"no cookie", false, nil, http.StatusUnauthorized},
	}

	for _, e := range tests {
		rr := httptest.NewRecorder()

		req, _ := http.NewRequest("GET", "/", nil)
		if e.addCoodie {
			req.AddCookie(e.cookie)
		}

		handler := http.HandlerFunc(app.refreshUsingCookie)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf("%s: wrong status code return; expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}

func Test_app_deleteRefreshCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.deleteRefreshCookie)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("wrong status; expected %d but got %d", http.StatusAccepted, rr.Code)
	}

	foundCookie := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "__Host-refresh_token" {
			foundCookie = true
			if c.Expires.After(time.Now()) {
				t.Errorf("cookie expiration in future, and shout not be: %v", c.Expires.UTC())
			}
		}
	}

	if !foundCookie {
		t.Errorf("__Host-refresh_token cookie not found")
	}
}
