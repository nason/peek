package spinner

import (
	"fmt"
	"time"

	"github.com/tj/go-spin"
)

// StartSpinning displays an async animated spinner with a message
func StartSpinning(msg string) chan bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)

	go func() {
		s := spin.New()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("\r%s... %s", msg, s.Next())
			case <-done:
				fmt.Printf("\r%s... done\n\n", msg)
				ticker.Stop()
				return
			}
		}
	}()

	return done
}
