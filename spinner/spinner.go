package spinner

import (
	"fmt"
	"time"

	"github.com/tj/go-spin"
)

// StartSpinning displays an async animated spinner with a message
func StartSpinning(msg string) func() {
	ticker := time.NewTicker(100 * time.Millisecond)
	stop := make(chan bool)
	done := make(chan bool)

	go func() {
		s := spin.New()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("\r%s... %s", msg, s.Next())
			case <-stop:
				fmt.Printf("\r%s... done\n\n", msg)
				ticker.Stop()
				done <- true
				return
			}
		}
	}()

	return func() {
		stop <- true
		<-done
	}
}
