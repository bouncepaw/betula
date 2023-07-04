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
	Status string
}

func (a *auther) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	// Auth is OK if it is set up or the user wants to set it up or they request static data.
	authOK := auth.Ready() ||
		strings.HasPrefix(rq.URL.Path, "/static/") ||
		strings.HasPrefix(rq.URL.Path, "/register")

	// We don't support anything else.
	// A thought for a future Bouncepaw: maybe we should support HEAD?
	allowedMethod := rq.Method == http.MethodGet || rq.Method == http.MethodPost

	if !allowedMethod {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(
			fmt.Sprintf("Method %s is not supported by this server. Use POST and GET.", rq.Method)))
		return
	}

	if !authOK {
		templateExec(w, templateRegisterForm, dataAuthorized{
			dataCommon: emptyCommon(),
		}, rq)
		return
	}

	a.Handler.ServeHTTP(w, rq)
}
