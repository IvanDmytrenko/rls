package rls

import "time"

const (
	defaultPrefix   = "user"
	defaultKey      = "default_action"
	defaultDuration = time.Second
)

//Options configures a Rate Limiter
type Options struct {
	Limit    int64
	Prefix   string
	Key      string
	Duration time.Duration
}
