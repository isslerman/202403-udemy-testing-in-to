package main

import (
	"encoding/gob"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"webapp/pkg/data"
	"webapp/pkg/repository"
	"webapp/pkg/repository/dbrepo"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	DNS     string
	Session *scs.SessionManager
	// DB implements the interface repo
	DB repository.DatabaseRepo
}

func main() {
	// important to register the typo of user to be used with session
	gob.Register(data.User{})
	// setup application config
	app := application{}

	flag.StringVar(&app.DNS, "dns", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgre connection")
	// execute the flag config
	flag.Parse()

	// START DB CONN
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	// don't close the connection until the func main is closed.
	defer conn.Close()
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	// END DB CONN

	// get a session manager
	app.Session = getSession()

	// print out a message
	slog.Info("Server is running in port 8080")

	// start the server
	err = http.ListenAndServe(":8080", app.routes())
	if err != nil {
		slog.Error("Something went wrong! ", err)
	}

}
