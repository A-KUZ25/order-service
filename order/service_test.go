package order

import "context"

type mockRepo struct {
	FetchUnpaidFn        func(ctx context.Context, f UnpaidFilter) ([]int64, error)
	FetchBadReviewFn     func(ctx context.Context, f BadReviewFilter) ([]int64, error)
	FetchExceededPriceFn func(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
}
