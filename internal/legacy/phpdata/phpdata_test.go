package phpdata

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshal_Primitives(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    any
		wantErr bool
	}{
		{name: "empty", raw: "", want: nil},
		{name: "string", raw: `s:5:"hello";`, want: "hello"},
		{name: "int", raw: `i:42;`, want: int64(42)},
		{name: "float", raw: `d:12.5;`, want: 12.5},
		{name: "bool true", raw: `b:1;`, want: true},
		{name: "bool false", raw: `b:0;`, want: false},
		{name: "null", raw: `N;`, want: nil},
		{name: "broken", raw: `a:1:{`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal([]byte(tt.raw))
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUnmarshal_NormalizesAssociativeAndNestedArrays(t *testing.T) {
	raw := `a:3:{` +
		`s:4:"city";s:7:"Izhevsk";` +
		`s:6:"nested";a:2:{i:0;s:3:"one";i:1;s:3:"two";}` +
		`i:7;s:5:"seven";` +
		`}`

	got, err := Unmarshal([]byte(raw))

	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"city":   "Izhevsk",
		"nested": []any{"one", "two"},
		"7":      "seven",
	}, got)
}

func TestCoerceString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "string", in: "abc", want: "abc"},
		{name: "bytes", in: []byte("abc"), want: "abc"},
		{name: "int64", in: int64(42), want: "42"},
		{name: "float64", in: 12.5, want: "12.5"},
		{name: "bool", in: true, want: "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, CoerceString(tt.in))
		})
	}
}

func TestCoerceInt64(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int64
		ok   bool
	}{
		{name: "int", in: 12, want: 12, ok: true},
		{name: "int64", in: int64(34), want: 34, ok: true},
		{name: "float64", in: 56.9, want: 56, ok: true},
		{name: "string", in: "78", want: 78, ok: true},
		{name: "bytes", in: []byte("90"), want: 90, ok: true},
		{name: "invalid string", in: "abc", want: 0, ok: false},
		{name: "bool", in: true, want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CoerceInt64(tt.in)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
