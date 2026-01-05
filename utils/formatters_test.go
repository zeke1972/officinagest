package utils

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "data valida",
			time: time.Date(2026, 1, 8, 12, 0, 0, 0, time.UTC),
			want: "08/01/2026",
		},
		{
			name: "data zero",
			time: time.Time{},
			want: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDate(tt.time); got != tt.want {
				t.Errorf("FormatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "datetime valida",
			time: time.Date(2026, 1, 8, 14, 30, 0, 0, time.UTC),
			want: "08/01/2026 14:30",
		},
		{
			name: "datetime zero",
			time: time.Time{},
			want: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDateTime(tt.time); got != tt.want {
				t.Errorf("FormatDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatEuro(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{"importo positivo", 100.50, "€ 100.50"},
		{"importo zero", 0, "€ 0.00"},
		{"importo con decimali", 123.456, "€ 123.46"},
		{"importo grande", 10000.00, "€ 10000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatEuro(tt.amount); got != tt.want {
				t.Errorf("FormatEuro() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"numero con punto", "123.45", 123.45, false},
		{"numero con virgola", "123,45", 123.45, false},
		{"numero intero", "100", 100, false},
		{"con simbolo euro", "€ 100.50", 100.50, false},
		{"con spazi", "  123.45  ", 123.45, false},
		{"stringa vuota", "", 0, false},
		{"formato invalido", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"stringa corta", "test", 10, "test"},
		{"stringa esatta", "test", 4, "test"},
		{"stringa lunga", "questo è un test", 10, "questo..."},
		{"max troppo piccolo", "test", 2, "t..."},
		{"stringa vuota", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.input, tt.max); got != tt.want {
				t.Errorf("Truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"numero piccolo", 123, "123"},
		{"numero migliaia", 1234, "1.234"},
		{"numero milioni", 1234567, "1.234.567"},
		{"zero", 0, "0"},
		{"numero esatto migliaia", 1000, "1.000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatInt(tt.input); got != tt.want {
				t.Errorf("FormatInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
