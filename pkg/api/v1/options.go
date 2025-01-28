package v1

import (
	"time"

	"golang.org/x/time/rate"
)

type ApiOption func(api *Api)

func WithTimeout(timeout time.Duration) ApiOption {
	return func(api *Api) {
		api.timeout = timeout
	}
}
func WithRateLimit(rps int) ApiOption {
	return func(api *Api) {
		api.rateLimiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), rps)
	}
}
