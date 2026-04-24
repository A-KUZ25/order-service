package redisactive

import (
	"bytes"
	"compress/gzip"
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestMaybeGunzip(t *testing.T) {
	t.Run("plain payload", func(t *testing.T) {
		raw := []byte(`a:1:{s:9:"wait_time";i:5;}`)

		got, err := maybeGunzip(raw)

		require.NoError(t, err)
		require.Equal(t, raw, got)
	})

	t.Run("gzipped payload", func(t *testing.T) {
		raw := []byte(`a:1:{s:9:"wait_time";i:5;}`)
		gzipped := gzipBytes(t, raw)

		got, err := maybeGunzip(gzipped)

		require.NoError(t, err)
		require.Equal(t, raw, got)
	})

	t.Run("broken gzip", func(t *testing.T) {
		_, err := maybeGunzip([]byte{0x1f, 0x8b, 0x08, 0x00})
		require.Error(t, err)
	})
}

func TestGetWorkerWaitingTime(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewActiveOrdersRepository(client)

	t.Run("missing order", func(t *testing.T) {
		got, err := repo.GetWorkerWaitingTime(ctx, 68, 100)
		require.NoError(t, err)
		require.Zero(t, got)
	})

	t.Run("plain payload with int wait time", func(t *testing.T) {
		mr.HSet("68", "101", `a:1:{s:9:"wait_time";i:5;}`)

		got, err := repo.GetWorkerWaitingTime(ctx, 68, 101)

		require.NoError(t, err)
		require.Equal(t, int64(300), got)
	})

	t.Run("gzipped payload with string wait time", func(t *testing.T) {
		payload := []byte(`a:1:{s:9:"wait_time";s:2:"12";}`)
		mr.HSet("68", "102", string(gzipBytes(t, payload)))

		got, err := repo.GetWorkerWaitingTime(ctx, 68, 102)

		require.NoError(t, err)
		require.Equal(t, int64(720), got)
	})

	t.Run("payload without wait time", func(t *testing.T) {
		mr.HSet("68", "103", `a:1:{s:4:"test";s:2:"ok";}`)

		got, err := repo.GetWorkerWaitingTime(ctx, 68, 103)

		require.NoError(t, err)
		require.Zero(t, got)
	})

	t.Run("broken payload", func(t *testing.T) {
		mr.HSet("68", "104", `a:1:{`)

		got, err := repo.GetWorkerWaitingTime(ctx, 68, 104)

		require.Error(t, err)
		require.Zero(t, got)
	})
}

func gzipBytes(t *testing.T, payload []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(payload)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return buf.Bytes()
}
