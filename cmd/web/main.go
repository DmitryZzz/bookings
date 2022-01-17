package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"github.com/DmitryZzz/bookings/internal/config"
	"github.com/DmitryZzz/bookings/internal/handlers"
	"github.com/DmitryZzz/bookings/internal/models"
	"github.com/DmitryZzz/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"
var app config.AppConfig
var session *scs.SessionManager

func main() {

	gob.Register(models.Reservation{})
	
	// change this to true when in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session
	
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}
	app.TemplateCache = tc
	app.UseCache = false
	render.NewTemplates(&app)

	repo := handlers.NewRepo(&app)

	handlers.NewHandlers(repo)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
