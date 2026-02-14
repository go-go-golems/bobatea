package asyncprovider

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSuccess(t *testing.T) {
	out, err := Run(
		context.Background(),
		7,
		100*time.Millisecond,
		"provider",
		"provider",
		func(context.Context) (string, error) {
			return "ok", nil
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}

func TestRunPanicRecovery(t *testing.T) {
	out, err := Run(
		context.Background(),
		9,
		100*time.Millisecond,
		"provider",
		"provider",
		func(context.Context) (string, error) {
			panic("boom")
		},
	)
	assert.Empty(t, out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider panic: boom")
}

func TestRunTimeout(t *testing.T) {
	start := time.Now()
	_, err := Run(
		context.Background(),
		11,
		10*time.Millisecond,
		"provider",
		"provider",
		func(ctx context.Context) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
	)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.Less(t, time.Since(start), time.Second)
}

func TestRunHonorsCanceledBaseContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Run(
		ctx,
		13,
		200*time.Millisecond,
		"provider",
		"provider",
		func(ctx context.Context) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
	)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), context.Canceled.Error()))
}
