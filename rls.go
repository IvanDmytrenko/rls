// Package rls provides rate limit service
package rls

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

//RateLimiter is a struct
type RateLimiter struct {
	RClient *redis.Client
	Options Options
}

// NewRateLimiter is a constructor for RateLimiter.
func NewRateLimiter(rClient *redis.Client, opts Options) *RateLimiter {
	if opts.Prefix == "" {
		opts.Prefix = defaultPrefix
	}
	if opts.Key == "" {
		opts.Key = defaultKey
	}
	if opts.Duration == 0 {
		opts.Duration = defaultDuration
	}
	rls := &RateLimiter{
		RClient: rClient,
		Options: opts,
	}

	return rls
}

//Allow reports whether events may happen now
func (rls *RateLimiter) Allow(id string) (bool, error) {
	key := fmt.Sprintf("%s_%s_%s", rls.Options.Prefix, id, rls.Options.Key)
	val, err := rls.getSet(key)
	if err != nil {
		err = errors.Wrapf(err, "limiter: failed to getSet %s", key)
		return false, err
	}

	if val <= 0 {
		return false, nil
	}

	err = rls.RClient.Decr(key).Err()
	if err != nil {
		err = errors.Wrapf(err, "limiter: failed to decr %s", key)
		return false, err
	}

	return true, nil
}

func (rls *RateLimiter) getSet(key string) (int, error) {
	exists, err := rls.RClient.Exists(key).Result()
	if err != nil {
		err = errors.Wrapf(err, "limiter: failed to check existence %s", key)
		return 0, err
	}

	if exists == 0 {
		err := rls.RClient.Set(key, rls.Options.Limit, rls.Options.Duration).Err()
		if err != nil {
			err = errors.Wrapf(err, "limiter: failed to set %s", key)
			return 0, err
		}
	}

	val, err := rls.RClient.Get(key).Result()
	if err != nil {
		err = errors.Wrapf(err, "limiter: failed to get %s", key)
		return 0, err
	}

	valConv, err := strconv.Atoi(val)
	if err != nil {
		err = errors.Wrapf(err, "limiter: failed to convert value %s", key)
		return 0, err
	}

	return valConv, nil
}
