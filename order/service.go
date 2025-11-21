package order

import (
	"context"
)

// SortOrder описывает направление сортировки.
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// UnpaidOrdersFilter — то, что нужно сервису, чтобы найти неоплаченные заказы.
type UnpaidOrdersFilter struct {
	TenantID               int64
	CityIDs                []int64 // может быть null
	Date                   *string // пока строка, потом можем сделать time.Time
	StatusTimeFrom         *int64  // unix timestamp (seconds)
	StatusTimeTo           *int64
	Status                 []string
	Tariffs                []int64
	UserPositions          []int64
	SortField              string
	SortOrder              SortOrder
	StatusCompletedNotPaid int64
}

// Repository — интерфейс для работы с хранилищем заказов.
type Repository interface {
	FetchUnpaidOrderIDs(ctx context.Context, filter UnpaidOrdersFilter) ([]int64, error)
}

// Service — интерфейс доменного сервиса заказов.
type Service interface {
	GetUnpaidOrderIDs(ctx context.Context, filter UnpaidOrdersFilter) ([]int64, error)
}

// service — конкретная реализация Service.
type service struct {
	repo Repository
}

// NewService — конструктор сервиса заказов.
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetUnpaidOrderIDs(ctx context.Context, filter UnpaidOrdersFilter) ([]int64, error) {
	return s.repo.FetchUnpaidOrderIDs(ctx, filter)
}
