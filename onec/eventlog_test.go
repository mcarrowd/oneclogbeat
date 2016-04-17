package onec

import (
	"golang.org/x/text/encoding/charmap"
	"testing"
)

func TestWindows1251Inverse(t *testing.T) {
	testCases := []struct {
		testString string
	}{
		{"Ира Рома Энгельс"},
		{"АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"},
		{"абвгдеёжзийклмнопрстуфхцчшщъыьэюя"},
	}
	for _, testCase := range testCases {
		want := testCase.testString
		got := decodeWindows1251(encodeWindows1251(want))
		if want != got {
			t.Errorf("%q, %q", want, got)
		}
	}
}

func decodeWindows1251(s string) string {
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes([]uint8(s))
	return string(out)
}
