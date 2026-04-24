package address

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAddress_ReturnsNilOnEmptyInput(t *testing.T) {
	parser := NewParser()

	got, err := parser.ParseAddress("")

	require.NoError(t, err)
	require.Nil(t, got)
}

func TestParseAddress_PreservesOrderAndPHPNullability(t *testing.T) {
	parser := NewParser()
	raw := `a:2:{` +
		`i:1;a:8:{` +
		`s:7:"city_id";s:5:"26068";` +
		`s:4:"city";s:7:"Izhevsk";` +
		`s:6:"street";s:22:"proezd imeni Deryabina";` +
		`s:5:"label";N;` +
		`s:5:"house";s:1:"3";` +
		`s:3:"apt";s:0:"";` +
		`s:7:"parking";s:9:"Deryabina";` +
		`s:8:"place_id";s:0:"";` +
		`}` +
		`i:0;a:8:{` +
		`s:7:"city_id";i:26068;` +
		`s:4:"city";s:7:"Izhevsk";` +
		`s:6:"street";s:22:"Supermarket Magnit #97";` +
		`s:5:"label";s:0:"";` +
		`s:5:"house";s:0:"";` +
		`s:3:"apt";N;` +
		`s:7:"parking";s:6:"123456";` +
		`s:8:"place_id";i:0;` +
		`}` +
		`}`

	got, err := parser.ParseAddress(raw)

	require.NoError(t, err)
	require.Len(t, got, 2)

	require.Equal(t, "26068", deref(got[0].ID))
	require.Equal(t, "Supermarket Magnit #97", deref(got[0].Street))
	require.Nil(t, got[0].Label)
	require.Nil(t, got[0].House)
	require.Nil(t, got[0].Apt)
	require.Equal(t, "123456", deref(got[0].Parking))
	require.Equal(t, "house", got[0].Type)

	require.Equal(t, "proezd imeni Deryabina", deref(got[1].Street))
	require.Nil(t, got[1].Label)
	require.Equal(t, "3", deref(got[1].House))
	require.Nil(t, got[1].Apt)
	require.Equal(t, "Deryabina", deref(got[1].Parking))
	require.Equal(t, "house", got[1].Type)
}

func TestParseAddress_SetsPlaceTypeForNonEmptyPlaceID(t *testing.T) {
	parser := NewParser()
	raw := `a:1:{i:0;a:8:{` +
		`s:7:"city_id";s:5:"26068";` +
		`s:4:"city";s:7:"Izhevsk";` +
		`s:6:"street";s:6:"Street";` +
		`s:5:"label";s:5:"Label";` +
		`s:5:"house";s:1:"1";` +
		`s:3:"apt";s:1:"2";` +
		`s:7:"parking";s:1:"P";` +
		`s:8:"place_id";s:2:"55";` +
		`}}`

	got, err := parser.ParseAddress(raw)

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "place", got[0].Type)
}

func TestParseAddress_ReturnsErrorOnBrokenPayload(t *testing.T) {
	parser := NewParser()

	got, err := parser.ParseAddress(`a:1:{`)

	require.Error(t, err)
	require.Nil(t, got)
}

func TestParseAddress_SkipsUnexpectedItems(t *testing.T) {
	parser := NewParser()
	raw := `a:2:{i:0;s:5:"wrong";i:1;a:8:{` +
		`s:7:"city_id";s:5:"26068";` +
		`s:4:"city";s:7:"Izhevsk";` +
		`s:6:"street";s:6:"Street";` +
		`s:5:"label";N;` +
		`s:5:"house";N;` +
		`s:3:"apt";N;` +
		`s:7:"parking";N;` +
		`s:8:"place_id";N;` +
		`}}`

	got, err := parser.ParseAddress(raw)

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "Street", deref(got[0].Street))
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
