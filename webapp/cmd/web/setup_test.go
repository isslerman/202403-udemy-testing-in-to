package main

import (
	"os"
	"testing"
	"webapp/pkg/repository/dbrepo"
)

var app application

func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates/"

	app.Session = getSession()
	// START DB TEST CONN
	app.DB = &dbrepo.TestDBRepo{}
	// END DB CONN

	os.Exit(m.Run())
}
