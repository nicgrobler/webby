package webby

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HTTPServer is the basic abstraction used for handling http requests
type HTTPServer struct {
	Slug     string
	Address  string
	Done     chan error
	router   *http.ServeMux
	listener *http.Server
}

// NewHTTPServer takes a valid address - can be of form IP:Port, or :Port - and returns a server
func NewHTTPServer(description, address string, idleConnectionTimeout int) *HTTPServer {
	s := &HTTPServer{Slug: description, Address: address, Done: make(chan error), router: http.NewServeMux()}
	s.setListener(&http.Server{Addr: address, Handler: s.router, IdleTimeout: time.Duration(idleConnectionTimeout)})
	return s
}

func (s *HTTPServer) setListener(l *http.Server) {
	s.listener = l
}

// RegisterHandler allows caller to set routing and handler functions as needed
func (s *HTTPServer) RegisterHandler(path string, handlerfn func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(path, handlerfn)
}

// StartListener starts the server's listener with a context, allowing for later graceful shutdown.
// the supplied timeout is the amount of time that is allowed before the server forcefully
// closes any remaining conections. Once done close Done channel
// note: this is a blocking call
func (s *HTTPServer) StartListener(ctx context.Context, timeout time.Duration) {

	go func() {
		if err := s.listener.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http listen:%+s", err)
		}
	}()

	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
	}()

	if err := s.listener.Shutdown(ctxShutDown); err != nil {
		if e := s.listener.Close(); e != nil {
			log.Fatalf(s.Slug+" forced shutdown failed:%+s", err)
		}
	}

	// let parent know that we are done
	close(s.Done)
}

//
// HELPER functions that follow are here for convenience and are optional
//

// SetCors is a helper that can be called from within any writeHandlers to enable CORS
func SetCors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Authorization")
	}
}

// SignalContext listens for os signals, and when received, calls cancel on returned context.
func SignalContext() context.Context {

	// listen for any and all signals
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	// set context so we can cancel the listner(s)
	ctx, cancel := context.WithCancel(context.Background())

	// prepare to cancel context on receipt of os signal
	go func() {
		<-c
		cancel()
	}()

	return ctx

}
