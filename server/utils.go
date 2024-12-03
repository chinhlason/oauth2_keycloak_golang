package server

import (
	"fmt"
	"time"
)

func Retry(attemps int, delay time.Duration, callback func() error) error {
	var err error
	for i := 0; i < attemps; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("failed after %d attemps, last error: %s", attemps, err)
}
