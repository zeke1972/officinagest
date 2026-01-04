package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatDate formatta un time.Time in formato italiano (GG/MM/AAAA)
// Restituisce "-" se la data è zero
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("02/01/2006")
}

// FormatDateTime formatta un time.Time in formato italiano con ora (GG/MM/AAAA HH:MM)
// Restituisce "-" se la data è zero
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("02/01/2006 15:04")
}

// FormatEuro formatta un importo in valuta Euro con 2 decimali
func FormatEuro(amount float64) string {
	return fmt.Sprintf("€ %.2f", amount)
}

// ParseFloat converte una stringa in float64, gestendo virgole come separatori decimali
// Accetta sia "," che "." come separatore decimale
func ParseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, "€", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, nil
	}

	return strconv.ParseFloat(s, 64)
}

// Truncate tronca una stringa alla lunghezza massima specificata
// Aggiunge "..." se la stringa viene troncata
func Truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}

	if len(s) <= max {
		return s
	}

	return s[:max-3] + "..."
}

// FormatInt formatta un intero con separatori delle migliaia
func FormatInt(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune('.')
		}
		result.WriteRune(digit)
	}

	return result.String()
}
