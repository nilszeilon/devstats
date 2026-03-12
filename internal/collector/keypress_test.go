package collector

import "testing"

func TestKeyCodeToString(t *testing.T) {
	tests := []struct {
		keycode  int64
		expected string
	}{
		{0, "a"},
		{1, "s"},
		{2, "d"},
		{3, "f"},
		{11, "b"},
		{12, "q"},
		{13, "w"},
		{14, "e"},
		{15, "r"},
		{16, "y"},
		{17, "t"},
		{18, "1"},
		{19, "2"},
		{36, "return"},
		{48, "tab"},
		{49, "space"},
		{51, "delete"},
		{53, "escape"},
		{55, "command"},
		{56, "shift"},
		{57, "capslock"},
		{58, "option"},
		{59, "control"},
		{96, "f5"},
		{97, "f6"},
		{98, "f7"},
		{99, "f3"},
		{100, "f8"},
		{101, "f9"},
		{123, "left_arrow"},
		{124, "right_arrow"},
		{125, "down_arrow"},
		{126, "up_arrow"},
		{999, "key_999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := keyCodeToString(tt.keycode)
			if result != tt.expected {
				t.Errorf("keyCodeToString(%d) = %q; want %q", tt.keycode, result, tt.expected)
			}
		})
	}
}
