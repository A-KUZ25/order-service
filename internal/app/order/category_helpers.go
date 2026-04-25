package order

func toSet(values []int64) map[int64]struct{} {
	set := make(map[int64]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return set
}

var categories = []struct {
	Name     string
	Statuses map[int64]struct{}
}{
	{
		Name: "new",
		Statuses: toSet([]int64{
			1, 2, 3, 4, 5, 52, 108, 109, 115, 127, 128, 130, 131, 136,
		}),
	},
	{
		Name: "works",
		Statuses: toSet([]int64{
			17, 26, 27, 29, 30, 36, 54, 55, 106, 110, 113, 114, 132, 133, 134, 135,
		}),
	},
	{
		Name: "warning",
		Statuses: toSet([]int64{
			5, 16, 27, 30, 38, 45, 46, 47, 48, 52, 54, 129,
		}),
	},
	{
		Name: "pre_order",
		Statuses: toSet([]int64{
			6, 7, 16, 111, 112, 116, 117, 118, 119,
		}),
	},
	{
		Name: "completed",
		Statuses: toSet([]int64{
			37, 38,
		}),
	},
	{
		Name: "rejected",
		Statuses: toSet([]int64{
			39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
			49, 50, 51, 107, 120, 121, 122, 123, 124, 125, 126,
		}),
	},
}

func GetCategory(statusID int64) string {
	for _, c := range categories {
		if _, ok := c.Statuses[statusID]; ok {
			return c.Name
		}
	}
	return ""
}

const (
	DeviceDispatcher = "DISPATCHER"
	DeviceIOS        = "IOS"
	DeviceAndroid    = "ANDROID"
	DeviceWorker     = "WORKER"
	DeviceCabinet    = "CABINET"
	DeviceWeb        = "WEB"
)

func GetDeviceName(device string) string {
	switch device {
	case DeviceDispatcher:
		return "Диспетчер"
	case DeviceIOS:
		return "IOS"
	case DeviceAndroid:
		return "Android"
	case DeviceWorker:
		return "Борт"
	case DeviceCabinet:
		return "Кабинет"
	case DeviceWeb:
		return "Web site"
	default:
		return ""
	}
}
