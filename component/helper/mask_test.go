package helper

import "testing"

func TestMaskPhone(t *testing.T) {
	cases := map[string]string{
		"13812345678": "138****5678",
		"1381234":     "1*****4",
		"138":         "1*8",
		"12":          "12",
		"":            "",
	}
	for in, want := range cases {
		if got := MaskPhone(in); got != want {
			t.Errorf("MaskPhone(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMaskIdCard(t *testing.T) {
	cases := map[string]string{
		"110101199001011234": "110***********1234",
		"110101":             "1****1",
		"110":                "1*0",
	}
	for in, want := range cases {
		if got := MaskIdCard(in); got != want {
			t.Errorf("MaskIdCard(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMaskString(t *testing.T) {
	if got := MaskString("abcdef", 1, 1, '*'); got != "a****f" {
		t.Errorf("MaskString = %q, want %q", got, "a****f")
	}
	// head+tail >= length 时保留首尾各一位
	if got := MaskString("abc", 2, 2, '*'); got != "a*c" {
		t.Errorf("MaskString = %q, want %q", got, "a*c")
	}
	// 长度 <= 2 原样返回
	if got := MaskString("ab", 3, 4, '*'); got != "ab" {
		t.Errorf("MaskString = %q, want %q", got, "ab")
	}
}

