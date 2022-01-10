package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/8thgencore/bookings/internal/config"
	"github.com/8thgencore/bookings/internal/handlers"
	"github.com/8thgencore/bookings/internal/models"
	"github.com/8thgencore/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8000"

var app config.AppConfig
var session *scs.SessionManager

// main is the main application function
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server is running on port: %s\n", portNumber)
	fmt.Printf("URL is accessible at: http://localhost%s", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal("Error in ListenAndServe", err)
	}
}

func run() error {
	// what am I going to put in the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// create template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err)
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	// pass template cache to handlers
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	// pass template cache to renderer
	render.NewTemplates(&app)

	return nil
}
