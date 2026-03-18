package redisactive

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"orders-service/internal/legacy/phpdata"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type ActiveOrdersRepository struct {
	client *redis.Client
}

func NewActiveOrdersRepository(client *redis.Client) *ActiveOrdersRepository {
	return &ActiveOrdersRepository{
		client: client,
	}
}

func (r *ActiveOrdersRepository) GetWorkerWaitingTime(
	ctx context.Context,
	tenantID, orderID int64,
) (int64, error) {
	raw, err := r.client.HGet(
		ctx,
		strconv.FormatInt(tenantID, 10),
		strconv.FormatInt(orderID, 10),
	).Bytes()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	payload, err := maybeGunzip(raw)
	if err != nil {
		return 0, err
	}

	value, err := phpdata.Unmarshal(payload)
	if err != nil {
		return 0, err
	}

	orderData, ok := value.(map[string]any)
	if !ok {
		return 0, nil
	}

	waitTime, ok := phpdata.CoerceInt64(orderData["wait_time"])
	if !ok {
		return 0, nil
	}

	return waitTime * 60, nil
}

func maybeGunzip(raw []byte) ([]byte, error) {
	if len(raw) < 2 || raw[0] != 0x1f || raw[1] != 0x8b {
		return raw, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("gunzip active order payload: %w", err)
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read active order payload: %w", err)
	}

	return payload, nil
}
