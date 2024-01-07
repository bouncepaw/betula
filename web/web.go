// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/settings"
)

var serverRestartChannel = make(chan struct{})

func StartServer() {
	go restartServer()
	var srv = &http.Server{}
	for range serverRestartChannel {
		if err := srv.Close(); err != nil {
			// Is it important? Does it matter?
			log.Println("Closing server:", err)
		}
		srv = &http.Server{
			Addr:    listenAddr(),
			Handler: &auther{mux},
		}
		log.Printf("Running HTTP server at %s\n", srv.Addr)
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln(err)
			}
		}()
	}
}

func listenAddr() string {
	return fmt.Sprintf("%s:%d", settings.NetworkHost(), settings.NetworkPort())
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
		templateExec(w, rq, templateRegisterForm, dataAuthorized{
			dataCommon: emptyCommon(),
		})
		return
	}

	a.Handler.ServeHTTP(w, rq)
}
