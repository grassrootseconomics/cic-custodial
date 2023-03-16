package task

import (
	"time"

	"github.com/bsm/redislock"
)

const (
	lockPrefix     = "lock:"
	lockRetryDelay = 25 * time.Millisecond
)

// lockRetry will at most try to obtain the lock 20 times within ~0.5s.
// it is expected to prevent immidiate requeue of the task at the expense of more redis calls.
func lockRetry() redislock.RetryStrategy {
	return redislock.LimitRetry(
		redislock.LinearBackoff(lockRetryDelay),
		20,
	)
}
