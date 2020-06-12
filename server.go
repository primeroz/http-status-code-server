/*
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var acceptedMediaTypes = []string{
	"image/webp",
	"image/svg+xml",
	"image/jpeg",
	"image/png",
	"image/",
}

const jsonContentType = "application/json; encoding=utf-8"

func main() {
	// command line arguments
	port := flag.Int("port", 8080, "Port number to serve on.")
	timeout := flag.Duration("timeout", 5*time.Second, "Time in seconds to wait before forcefully terminating the server.")

	flag.Parse()

	statusServer := newHTTPServer(fmt.Sprintf(":%d", *port), statusHandler())

	// start the main http server
	go func() {
		err := statusServer.ListenAndServe()
		if err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "could not start http server: %s\n", err)
			os.Exit(1)
		}
	}()

	waitShutdown(statusServer, *timeout)
}

type server struct {
	mux *http.ServeMux
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}
}

// Mostly a copy of https://github.com/mccutchen/go-httpbin/blob/dfdd248a96298dec09ec3fcd90b6254ceed6dafd/httpbin/handlers.go#L148
func status(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	code, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	type statusCase struct {
		headers map[string]string
		body    []byte
	}

	redirectHeaders := &statusCase{
		headers: map[string]string{
			"Location": "/redirect/1",
		},
	}

	notAcceptableBody, _ := json.Marshal(map[string]interface{}{
		"message": "Client did not request a supported media type",
		"accept":  acceptedMediaTypes,
	})

	specialCases := map[int]*statusCase{
		301: redirectHeaders,
		302: redirectHeaders,
		303: redirectHeaders,
		305: redirectHeaders,
		307: redirectHeaders,
		401: {
			headers: map[string]string{
				"WWW-Authenticate": `Basic realm="Fake Realm"`,
			},
		},
		402: {
			body: []byte("pay me!"),
			headers: map[string]string{
				"X-More-Info": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/402",
			},
		},
		403: {
			body: []byte("Access Forbidden"),
		},
		406: {
			body: notAcceptableBody,
			headers: map[string]string{
				"Content-Type": jsonContentType,
			},
		},
		407: {
			headers: map[string]string{
				"Proxy-Authenticate": `Basic realm="Fake Realm"`,
			},
		},
		418: {
			body: []byte("I'm a teapot!"),
			headers: map[string]string{
				"X-More-Info": "http://tools.ietf.org/html/rfc2324",
			},
		},
	}

	if specialCase, ok := specialCases[code]; ok {
		if specialCase.headers != nil {
			for key, val := range specialCase.headers {
				w.Header().Set(key, val)
			}
		}
		w.WriteHeader(code)
		if specialCase.body != nil {
			w.Write(specialCase.body)
		}
	} else {
		w.WriteHeader(code)
	}
}

func statusHandler(options ...func(*server)) *server {
	s := &server{mux: http.NewServeMux()}

	// Handle Healthz and Readiness Endpoint
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	s.mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Status Handling
	s.mux.HandleFunc("/status/", status)

	s.mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found - 404")
	})

	// Default handling
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found - 404")
	})

	return s
}

func waitShutdown(s *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Fprintf(os.Stdout, "stopping http server...\n")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "could not gracefully shutdown http server: %s\n", err)
	}
}
