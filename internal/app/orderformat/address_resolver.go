package orderformat

import "orders-service/order"

type RawAddressParser interface {
	ParseAddress(raw string) ([]order.AddressView, error)
}

type AddressResolver struct {
	parser RawAddressParser
}

func NewAddressResolver(parser RawAddressParser) *AddressResolver {
	return &AddressResolver{parser: parser}
}

func (r *AddressResolver) ResolveAddresses(orders []order.FullOrder) (map[int64][]order.AddressView, error) {
	result := make(map[int64][]order.AddressView, len(orders))

	for _, o := range orders {
		if o.Address == "" || r.parser == nil {
			result[o.OrderID] = nil
			continue
		}

		addresses, err := r.parser.ParseAddress(o.Address)
		if err != nil {
			return nil, err
		}
		result[o.OrderID] = addresses
	}

	return result, nil
}
