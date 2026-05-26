package scan

import "testing"

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{size: 0, want: "0 B"},
		{size: 1, want: "1 B"},
		{size: 1023, want: "1023 B"},
		{size: 1024, want: "1 KB"},
		{size: 1024*1024 + 777, want: "1 MB"},
		{size: 1024*1024*1024 + 999999, want: "1 GB"},
	}

	for _, tc := range tests {
		if got := FormatSize(tc.size); got != tc.want {
			t.Fatalf("FormatSize(%d): got %q want %q", tc.size, got, tc.want)
		}
	}
}
