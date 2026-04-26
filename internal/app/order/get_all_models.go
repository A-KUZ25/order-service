package order

type SearchAttribute struct {
	Attribute    string
	SearchString string
}

type GetAllOrdersFilter struct {
	BaseFilter

	Page         int
	PageSize     int
	SearchStatus string
	Attributes   []SearchAttribute
	SearchString map[string]string
	ShopIDs      []int64
}

type GetAllOrdersResult struct {
	OrderTotalCount int64
	CountPerPage    int
	Orders          []OrderView
}
