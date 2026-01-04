package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateNotEmpty verifica che un campo non sia vuoto
func ValidateNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s non può essere vuoto", fieldName)
	}
	return nil
}

// ValidateTarga valida una targa italiana (formato flessibile)
// Accetta formati: AA000AA, AA 000 AA, AA000AA, ecc.
func ValidateTarga(targa string) error {
	t := strings.ToUpper(strings.ReplaceAll(targa, " ", ""))

	if len(t) < 6 {
		return fmt.Errorf("targa troppo corta (minimo 6 caratteri)")
	}

	if len(t) > 8 {
		return fmt.Errorf("targa troppo lunga (massimo 8 caratteri)")
	}

	// Controllo caratteri validi (solo lettere e numeri)
	match, _ := regexp.MatchString(`^[A-Z0-9]+$`, t)
	if !match {
		return fmt.Errorf("targa contiene caratteri non validi")
	}

	return nil
}

// ValidateEmail valida un indirizzo email (opzionale)
// Restituisce nil se l'email è vuota o valida
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return nil // Email opzionale
	}

	// Regex base per email
	match, _ := regexp.MatchString(`^[\w\-\.]+@([\w\-]+\.)+[\w\-]{2,4}$`, email)
	if !match {
		return fmt.Errorf("formato email non valido")
	}

	return nil
}

// ValidateCodiceFiscale valida un codice fiscale italiano
// Controllo base sulla lunghezza e caratteri
func ValidateCodiceFiscale(cf string) error {
	cf = strings.ToUpper(strings.ReplaceAll(cf, " ", ""))

	if cf == "" {
		return nil // CF opzionale
	}

	if len(cf) != 16 {
		return fmt.Errorf("codice fiscale deve essere di 16 caratteri")
	}

	match, _ := regexp.MatchString(`^[A-Z]{6}[0-9]{2}[A-Z][0-9]{2}[A-Z][0-9]{3}[A-Z]$`, cf)
	if !match {
		return fmt.Errorf("formato codice fiscale non valido")
	}

	return nil
}

// ValidatePartitaIVA valida una partita IVA italiana
func ValidatePartitaIVA(piva string) error {
	piva = strings.ReplaceAll(piva, " ", "")

	if piva == "" {
		return nil // P.IVA opzionale
	}

	if len(piva) != 11 {
		return fmt.Errorf("partita IVA deve essere di 11 cifre")
	}

	match, _ := regexp.MatchString(`^[0-9]{11}$`, piva)
	if !match {
		return fmt.Errorf("partita IVA deve contenere solo numeri")
	}

	return nil
}

// ValidateCAP valida un CAP italiano
func ValidateCAP(cap string) error {
	cap = strings.ReplaceAll(cap, " ", "")

	if cap == "" {
		return nil // CAP opzionale
	}

	if len(cap) != 5 {
		return fmt.Errorf("CAP deve essere di 5 cifre")
	}

	match, _ := regexp.MatchString(`^[0-9]{5}$`, cap)
	if !match {
		return fmt.Errorf("CAP deve contenere solo numeri")
	}

	return nil
}

// ValidateTelefono valida un numero di telefono (formato flessibile)
func ValidateTelefono(tel string) error {
	tel = strings.ReplaceAll(tel, " ", "")
	tel = strings.ReplaceAll(tel, "-", "")
	tel = strings.ReplaceAll(tel, "/", "")

	if tel == "" {
		return nil // Telefono opzionale
	}

	// Rimuovi prefisso internazionale se presente
	tel = strings.TrimPrefix(tel, "+39")
	tel = strings.TrimPrefix(tel, "0039")

	if len(tel) < 8 || len(tel) > 12 {
		return fmt.Errorf("numero di telefono non valido")
	}

	match, _ := regexp.MatchString(`^[0-9]+$`, tel)
	if !match {
		return fmt.Errorf("telefono deve contenere solo numeri")
	}

	return nil
}

// ValidateImporto valida che un importo sia positivo
func ValidateImporto(importo float64) error {
	if importo < 0 {
		return fmt.Errorf("importo non può essere negativo")
	}
	return nil
}

// ValidateImportoPositivo valida che un importo sia maggiore di zero
func ValidateImportoPositivo(importo float64) error {
	if importo <= 0 {
		return fmt.Errorf("importo deve essere maggiore di zero")
	}
	return nil
}
