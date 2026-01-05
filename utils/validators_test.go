package utils

import "testing"

func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{"valore valido", "test", "campo", false},
		{"valore vuoto", "", "campo", true},
		{"solo spazi", "   ", "campo", true},
		{"con spazi ma valido", "  test  ", "campo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotEmpty(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTarga(t *testing.T) {
	tests := []struct {
		name    string
		targa   string
		wantErr bool
	}{
		{"targa valida standard", "AB123CD", false},
		{"targa con spazi", "AB 123 CD", false},
		{"targa vecchio formato", "MI123456", false},
		{"targa troppo corta", "AB12", true},
		{"targa troppo lunga", "AB1234567", true},
		{"caratteri non validi", "AB-123-CD", true},
		{"solo lettere", "ABCDEFG", false},
		{"solo numeri", "1234567", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTarga(tt.targa)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTarga(%v) error = %v, wantErr %v", tt.targa, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"email valida", "test@example.com", false},
		{"email con subdomain", "test@mail.example.com", false},
		{"email vuota (opzionale)", "", false},
		{"email con numeri", "test123@example.com", false},
		{"email invalida senza @", "testexample.com", true},
		{"email invalida senza dominio", "test@", true},
		{"email invalida formato", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%v) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCodiceFiscale(t *testing.T) {
	tests := []struct {
		name    string
		cf      string
		wantErr bool
	}{
		{"CF valido", "RSSMRA80A01H501U", false},
		{"CF vuoto (opzionale)", "", false},
		{"CF con spazi", "RSSMRA80A01H501U ", false},
		{"CF troppo corto", "RSSMRA80A01H501", true},
		{"CF troppo lungo", "RSSMRA80A01H501UX", true},
		{"CF formato invalido", "1234567890123456", true},
		{"CF lowercase", "rssmra80a01h501u", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodiceFiscale(tt.cf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCodiceFiscale(%v) error = %v, wantErr %v", tt.cf, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePartitaIVA(t *testing.T) {
	tests := []struct {
		name    string
		piva    string
		wantErr bool
	}{
		{"P.IVA valida", "12345678901", false},
		{"P.IVA vuota (opzionale)", "", false},
		{"P.IVA con spazi", "123 456 789 01", false},
		{"P.IVA troppo corta", "1234567890", true},
		{"P.IVA troppo lunga", "123456789012", true},
		{"P.IVA con lettere", "1234567890A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePartitaIVA(tt.piva)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePartitaIVA(%v) error = %v, wantErr %v", tt.piva, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCAP(t *testing.T) {
	tests := []struct {
		name    string
		cap     string
		wantErr bool
	}{
		{"CAP valido", "00100", false},
		{"CAP valido Milano", "20100", false},
		{"CAP vuoto (opzionale)", "", false},
		{"CAP con spazi", "00 100", false},
		{"CAP troppo corto", "0010", true},
		{"CAP troppo lungo", "001000", true},
		{"CAP con lettere", "0010A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCAP(tt.cap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCAP(%v) error = %v, wantErr %v", tt.cap, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTelefono(t *testing.T) {
	tests := []struct {
		name    string
		tel     string
		wantErr bool
	}{
		{"telefono valido mobile", "3331234567", false},
		{"telefono valido fisso", "0612345678", false},
		{"telefono con prefisso +39", "+393331234567", false},
		{"telefono con prefisso 0039", "00393331234567", false},
		{"telefono vuoto (opzionale)", "", false},
		{"telefono con spazi", "333 123 4567", false},
		{"telefono con trattini", "333-123-4567", false},
		{"telefono troppo corto", "123456", true},
		{"telefono troppo lungo", "12345678901234", true},
		{"telefono con lettere", "333123456A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTelefono(tt.tel)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTelefono(%v) error = %v, wantErr %v", tt.tel, err, tt.wantErr)
			}
		})
	}
}

func TestValidateImporto(t *testing.T) {
	tests := []struct {
		name    string
		importo float64
		wantErr bool
	}{
		{"importo positivo", 100.50, false},
		{"importo zero", 0, false},
		{"importo negativo", -10.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImporto(tt.importo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImporto(%v) error = %v, wantErr %v", tt.importo, err, tt.wantErr)
			}
		})
	}
}

func TestValidateImportoPositivo(t *testing.T) {
	tests := []struct {
		name    string
		importo float64
		wantErr bool
	}{
		{"importo positivo", 100.50, false},
		{"importo zero", 0, true},
		{"importo negativo", -10.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImportoPositivo(tt.importo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImportoPositivo(%v) error = %v, wantErr %v", tt.importo, err, tt.wantErr)
			}
		})
	}
}
