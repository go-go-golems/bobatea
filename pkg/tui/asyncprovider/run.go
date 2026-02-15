package asyncprovider

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog/log"
)

// Run executes fn with timeout and panic recovery semantics suitable for UI
// provider calls.
func Run[T any](
	baseCtx context.Context,
	requestID uint64,
	timeout time.Duration,
	providerName string,
	panicPrefix string,
	fn func(context.Context) (T, error),
) (T, error) {
	ctx, cancel := context.WithTimeout(baseCtx, timeout)
	defer cancel()

	var (
		out T
		err error
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Uint64("request_id", requestID).
					Str("provider", providerName).
					Msg("provider panicked")
				err = fmt.Errorf("%s panic: %v", panicPrefix, r)
			}
		}()
		out, err = fn(ctx)
	}()

	return out, err
}
