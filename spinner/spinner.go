package spinner

import (
	"fmt"
	"time"

	"github.com/tj/go-spin"
)

// Spinner controls an animated cli spinner with a message
type Spinner struct {
	message string
	ticker  *time.Ticker
	stop    chan bool
	done    chan bool
}

// New creates a new instantiation of a Spinner
func New(message string) Spinner {
	return Spinner{
		message: message,
		ticker:  time.NewTicker(100 * time.Millisecond),
		stop:    make(chan bool),
		done:    make(chan bool),
	}
}

// Start begins the spinner animation on a new goroutine
func (s *Spinner) Start() {
	go func() {
		spinFrames := spin.New()
		for {
			select {
			case <-s.ticker.C:
				fmt.Printf("\r%s... %s", s.message, spinFrames.Next())
			case <-s.stop:
				fmt.Printf("\r%s... done\n\n", s.message)
				s.ticker.Stop()
				s.done <- true
				return
			}
		}
	}()
}

// Stop halts the spinner animation synchronously
func (s *Spinner) Stop() {
	s.stop <- true
	<-s.done
}
