package tgx

import "time"

func retry(callback func() error, maxRetries int, interval time.Duration) error {
	var err error
	for i := 1; i <= maxRetries; i++ {
		if err = callback(); err != nil {
			time.Sleep(interval)
			continue
		}
		return nil
	}
	return err
}
