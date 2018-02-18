// package main demonstrates a context based service start-stop pattern.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var clean = flag.Bool("clean", false, "services won't fail, requiring signal to exit.")

// Service represents a long running service in our application.
type Service struct {
	name    string
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
}

// New creates a new Service that will fail after timeout.
func New(name string, timeout time.Duration) *Service {
	return &Service{name: name, timeout: timeout}
}

// Start calls run() in a new goroutine, returning an error channel which will be closed once
// this service has exited.  Start is not thread safe, do not call from multiple goroutines.
func (s *Service) Start() <-chan error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	errc := make(chan error)
	go func() {
		defer close(errc)
		if err := s.run(); err != nil {
			errc <- err
		}
	}()
	return errc
}

// Stop requests our service to shutdown.
func (s *Service) Stop() {
	s.cancel()
}

// run would be where our service performs its work, starts its listener, etc.
func (s *Service) run() error {
	log.Printf("service %s started", s.name)
	// s.ctx should be used as a parent for request contexts, and sync.WaitGroup leveraged to
	// prevent this function from returning until all workers are finished.
	failc := time.After(time.Hour * 1000)
	if !*clean {
		failc = time.After(s.timeout)
	}
	select {
	case <-failc:
		// Pretend there was an error requiring this service to stop.
		return fmt.Errorf("service %s timed out after %v", s.name, s.timeout)
	case <-s.ctx.Done():
		// Stop requested.
		log.Printf("service %s stopped", s.name)
	}
	return nil
}

// main starts our services, restarts them after failures.
func main() {
	flag.Parse()

	// Create services, ignoring configuration errors.
	a := New("a", time.Second*3)
	b := New("b", time.Second*2)
	c := New("c", time.Second*5)
	// Start services.
	ac := a.Start()
	bc := b.Start()
	cc := c.Start()
	// Setup signal handler
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
retryLoop:
	for retries := 2; retries >= 0; retries-- {
		// Wait for any service to fail, restart them a couple times.
		select {
		case err := <-ac:
			log.Printf("error: %v", err)
			if retries > 0 {
				ac = a.Start()
			}
		case err := <-bc:
			log.Printf("error: %v", err)
			if retries > 0 {
				bc = b.Start()
			}
		case err := <-cc:
			log.Printf("error: %v", err)
			if retries > 0 {
				cc = c.Start()
			}
		case sig := <-sigc:
			log.Printf("got signal %v", sig)
			break retryLoop
		}
		log.Printf("(%v retries remaining)", retries)
	}
	log.Printf("shutting down")
	// Stop all services.
	a.Stop()
	b.Stop()
	c.Stop()
	// Wait for all services to finish.
	if err := <-ac; err != nil {
		log.Printf("a error: %v", err)
	}
	if err := <-bc; err != nil {
		log.Printf("b error: %v", err)
	}
	if err := <-cc; err != nil {
		log.Printf("c error: %v", err)
	}
}
