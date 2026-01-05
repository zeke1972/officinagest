package database

import "testing"

func TestIsValidMetodoPagamento(t *testing.T) {
	tests := []struct {
		name   string
		metodo string
		want   bool
	}{
		{"cassa valido", MetodoPagamentoCassa, true},
		{"banca valido", MetodoPagamentoBanca, true},
		{"pos valido", MetodoPagamentoPOS, true},
		{"assegno valido", MetodoPagamentoAssegno, true},
		{"bonifico valido", MetodoPagamentoBonifico, true},
		{"metodo invalido", "CARTA", false},
		{"stringa vuota", "", false},
		{"lowercase", "cassa", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidMetodoPagamento(tt.metodo); got != tt.want {
				t.Errorf("IsValidMetodoPagamento(%v) = %v, want %v", tt.metodo, got, tt.want)
			}
		})
	}
}

func TestIsValidStatoCommessa(t *testing.T) {
	tests := []struct {
		name  string
		stato string
		want  bool
	}{
		{"aperta valido", StatoCommessaAperta, true},
		{"chiusa valido", StatoCommessaChiusa, true},
		{"stato invalido", "InCorso", false},
		{"stringa vuota", "", false},
		{"lowercase", "aperta", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidStatoCommessa(tt.stato); got != tt.want {
				t.Errorf("IsValidStatoCommessa(%v) = %v, want %v", tt.stato, got, tt.want)
			}
		})
	}
}

func TestIsValidTipoMovimento(t *testing.T) {
	tests := []struct {
		name string
		tipo string
		want bool
	}{
		{"entrata valido", TipoMovimentoEntrata, true},
		{"uscita valido", TipoMovimentoUscita, true},
		{"tipo invalido", "Altro", false},
		{"stringa vuota", "", false},
		{"lowercase", "entrata", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTipoMovimento(tt.tipo); got != tt.want {
				t.Errorf("IsValidTipoMovimento(%v) = %v, want %v", tt.tipo, got, tt.want)
			}
		})
	}
}

func TestValidMetodiPagamento(t *testing.T) {
	metodi := ValidMetodiPagamento()

	if len(metodi) != 5 {
		t.Errorf("ValidMetodiPagamento() returned %d methods, want 5", len(metodi))
	}

	expectedMetodi := []string{
		MetodoPagamentoCassa,
		MetodoPagamentoBanca,
		MetodoPagamentoPOS,
		MetodoPagamentoAssegno,
		MetodoPagamentoBonifico,
	}

	for _, expected := range expectedMetodi {
		found := false
		for _, metodo := range metodi {
			if metodo == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidMetodiPagamento() missing expected method: %s", expected)
		}
	}
}

func TestValidStatiCommessa(t *testing.T) {
	stati := ValidStatiCommessa()

	if len(stati) != 2 {
		t.Errorf("ValidStatiCommessa() returned %d states, want 2", len(stati))
	}

	if stati[0] != StatoCommessaAperta || stati[1] != StatoCommessaChiusa {
		t.Errorf("ValidStatiCommessa() returned unexpected states")
	}
}

func TestValidTipiMovimento(t *testing.T) {
	tipi := ValidTipiMovimento()

	if len(tipi) != 2 {
		t.Errorf("ValidTipiMovimento() returned %d types, want 2", len(tipi))
	}

	if tipi[0] != TipoMovimentoEntrata || tipi[1] != TipoMovimentoUscita {
		t.Errorf("ValidTipiMovimento() returned unexpected types")
	}
}
