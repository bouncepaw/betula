// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
package web

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/settings"
	"log"
	"net/http"
	"strings"
)

var serverRestartChannel = make(chan struct{})

func StartServer() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.NetworkPort()),
		Handler: &auther{mux},
	}
	go func() {
		// слушать и служить
		log.Printf("Starting HTTP server at port %d\n", settings.NetworkPort())
		if err := srv.ListenAndServe(); err.Error() != "http: Server closed" {
			log.Fatalln(err)
		}
	}()
	for {
		select {
		case <-serverRestartChannel:
			if err := srv.Close(); err != nil {
				// Is it important? Does it matter?
				log.Println(err)
			}
			srv = &http.Server{
				Addr:    fmt.Sprintf(":%d", settings.NetworkPort()),
				Handler: &auther{mux},
			}
			log.Printf("Restarting HTTP server at port %d\n", settings.NetworkPort())
			go func() {
				if err := srv.ListenAndServe(); err.Error() != "http: Server closed" {
					log.Fatalln(err)
				}
			}()
		}
	}
}

func restartServer() {
	serverRestartChannel <- struct{}{}
}

type auther struct {
	http.Handler
}

type dataAuthorized struct {
	*dataCommon
}

func (a *auther) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	if auth.Ready() ||
		strings.HasPrefix(rq.URL.Path, "/static/") ||
		strings.HasPrefix(rq.URL.Path, "/register") {
		a.Handler.ServeHTTP(w, rq)
		return
	}
	templateExec(w, templateRegisterForm, dataAuthorized{
		dataCommon: emptyCommon(),
	}, rq)
}
