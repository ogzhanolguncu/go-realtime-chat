package chat_ratelimit

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenBucket(t *testing.T) {
	t.Run("should consume tokens when check called", func(t *testing.T) {
		ratelimit := NewRatelimit(TokenBucket{RefillInterval: 300 * time.Millisecond, RefillRate: 1, BucketLimit: 10})
		conn := &net.TCPConn{}

		ratelimit.Add(conn)
		ratelimit.Check(conn)
		ratelimit.Check(conn)
		ratelimit.Check(conn)

		availableToken := ratelimit.userRatelimitMap[conn]
		require.GreaterOrEqual(t, availableToken, AvailableToken(7))

		time.Sleep(500 * time.Millisecond)
		availableToken = ratelimit.userRatelimitMap[conn]
		require.GreaterOrEqual(t, availableToken, AvailableToken(8))
	})

	t.Run("should remove connection from map", func(t *testing.T) {
		ratelimit := NewRatelimit(TokenBucket{RefillInterval: 300 * time.Millisecond, RefillRate: 1, BucketLimit: 10})
		conn := &net.TCPConn{}

		ratelimit.Add(conn)
		ratelimit.Remove(conn)
		result := ratelimit.Check(conn)

		require.False(t, result)
	})

}
