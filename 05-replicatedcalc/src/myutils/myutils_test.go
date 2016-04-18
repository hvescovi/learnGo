package myutils

import "testing"

func TestGetMyIP(t *testing.T) {
	
	cases := []struct {
		in, want string
	}{
		{"150.162.", "150.162.64.177"},
		{"172.17", "172.17.42.1"},
	}
	
	for _, c := range cases {
		got := GetMyIP(c.in)
		if got != c.want {
			t.Errorf("GetMyIP(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}