package task

import (
	"math/big"
	"time"

	"github.com/bsm/redislock"
)

const (
	lockPrefix     = "lock:"
	lockRetryDelay = 25 * time.Millisecond
	lockTimeout    = 1 * time.Second
)

var (
	// 20 gwei = max gas price we are willing to pay
	// 250k = max gas limit
	// minGasBalanceRequired is optimistic that the immidiate next transfer request will be successful
	// but the subsequent one could fail (though low probability), therefore we can trigger a gas lock.
	// Therefore our system wide threshold is 0.01 CELO or 10000000000000000 gas units
	// UPDATE: Feb 2025
	// 0.04455 CELO
	minGasBalanceRequired = big.NewInt(27000000000 * 550000 * 3)
)

// lockRetry will at most try to obtain the lock 20 times within ~0.5s.
// it is expected to prevent immidiate requeue of the task at the expense of more redis calls.
func lockRetry() redislock.RetryStrategy {
	return redislock.LimitRetry(
		redislock.LinearBackoff(lockRetryDelay),
		20,
	)
}

// balanceCheck compares the network balance with the system set min as threshold to execute a transfer.
func balanceCheck(networkBalance big.Int) bool {
	return minGasBalanceRequired.Cmp(&networkBalance) < 0
}
