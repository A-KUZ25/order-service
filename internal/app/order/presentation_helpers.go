package order

var redStatuses = map[int64]struct{}{
	10: {}, 16: {}, 27: {}, 30: {}, 38: {},
	39: {}, 52: {}, 54: {}, 117: {}, 118: {},
	120: {}, 135: {},
}

func GetColor(statusID int64) string {
	if _, ok := redStatuses[statusID]; ok {
		return "#cc1919"
	}
	return "#088142"
}

func BuildDispatcher(o FormattedOrder) any {
	if o.Device == DeviceDispatcher {
		return map[string]any{
			"device": "Диспетчер",
			"user": map[string]any{
				"userId":     o.UserCreated.UserID,
				"name":       o.UserCreated.Name,
				"lastName":   o.UserCreated.LastName,
				"secondName": o.UserCreated.SecondName,
			},
		}
	}

	return map[string]any{
		"device": GetDeviceName(o.Device),
	}
}

func BuildWorker(o FormattedOrder) *WorkerView {
	if o.WorkerID == nil {
		return nil
	}

	name := ""
	if o.Worker.LastName != nil && o.Worker.Name != nil {
		name = *o.Worker.LastName + " " + string([]rune(*o.Worker.Name)[0]) + "."
	}

	return &WorkerView{
		WorkerID: o.Worker.WorkerID,
		Callsign: o.Worker.Callsign,
		Name:     name,
		Phone:    o.Worker.Phone,
	}
}

func BuildCar(o FormattedOrder) *CarView {
	if o.CarID == nil {
		return nil
	}

	return &CarView{
		CarID:  o.Car.CarID,
		Name:   o.Car.Name,
		Color:  o.Car.Color,
		Number: o.Car.GosNumber,
	}
}

func BuildTariff(o FormattedOrder) TariffView {
	t := TariffView{
		TariffID: o.Tariff.TariffID,
		Name:     o.Tariff.Name,
	}

	if o.Tariff.TariffType == "QUANTITATIVE" {
		t.QuantitativeTitle = &o.Tariff.QuantitativeTitle
		t.PriceForUnit = &o.Tariff.PriceForUnit
		t.UnitName = &o.Tariff.UnitName
	}

	return t
}
