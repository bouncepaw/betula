// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
package web

import (
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"net/http"
	"strconv"
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

// returns id, found
func extractPostID(w http.ResponseWriter, rq *http.Request) (int, bool) {
	parts := strings.Split(rq.URL.Path, "/")
	if len(parts) != 3 {
		handlerNotFound(w, rq)
		log.Printf("Extracting post no. from %s: wrong format\n", rq.URL.Path)
		return 0, false
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("Extracting post no. from %s: wrong format\n", rq.URL.Path)
		handlerNotFound(w, rq)
		return 0, false
	}

	return id, true
}

// Wrap handlers that only make sense for the admin with this thingy in init().
func adminOnly(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		authed := auth.AuthorizedFromRequest(rq)
		if !authed {
			log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
			handlerUnauthorized(w, rq)
			return
		}
		next(w, rq)
	}
}

func federatedOnly(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		federated := settings.FederationEnabled()
		if !federated {
			log.Printf("Attempt to access %s failed because Betula is not federated. %d.\n", rq.URL.Path, http.StatusUnauthorized)
			handlerNotFederated(w, rq)
			return
		}
		next(w, rq)
	}
}

func postOnly(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		if rq.Method != http.MethodPost {
			log.Printf("Accessing %s with method %s, which is not POST. 404.\n", rq.URL.Path, rq.Method)
			handlerNotFound(w, rq)
			return
		}
		next(w, rq)
	}
}

func fediverseWebFork(
	nextFedi func(http.ResponseWriter, *http.Request),
	nextWeb func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		if rq.Header.Get("Accept") == types.ActivityType && nextFedi != nil {
			federatedOnly(nextFedi)(w, rq)
		} else if nextWeb != nil {
			nextWeb(w, rq)
		} else {
			handlerNotFound(w, rq)
		}
	}
}
