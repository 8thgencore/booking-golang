package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/8thgencore/bookings/internal/config"
	"github.com/8thgencore/bookings/internal/driver"
	"github.com/8thgencore/bookings/internal/handlers"
	"github.com/8thgencore/bookings/internal/helpers"
	"github.com/8thgencore/bookings/internal/models"
	"github.com/8thgencore/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
)

const portNumber = ":8000"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	defer close(app.MailChan)

	// fmt.Println("Starting mail listener...")
	// listenForMail()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Error loading PORT from .env file")
	}

	addr := fmt.Sprintf(":%s", port)

	fmt.Printf("Server is running on port: %s\n", addr)
	fmt.Printf("URL is accessible at: http://localhost%s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// mailChan := make(chan models.MailData)
	// app.MailChan = mailChan

	// change this to true when in production
	app.InProduction = false
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("Error loading DSN from .env file")
	}

	// connect to database
	log.Println("Connecting to databese...")
	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database!")

	// create template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache", err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)

	helpers.NewHelpers(&app)

	return db, nil
}
