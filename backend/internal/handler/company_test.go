package handler

import "testing"

func TestCompanyAppealObjectPrefix(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: "unknown"},
		{name: "trimmed short", input: " 1234567 ", want: "1234567"},
		{name: "exact eight", input: "12345678", want: "12345678"},
		{name: "long", input: "1234567890", want: "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := companyAppealObjectPrefix(tt.input); got != tt.want {
				t.Fatalf("companyAppealObjectPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
