package database

import (
	"testing"
	"time"
)

func TestClienteValidate(t *testing.T) {
	tests := []struct {
		name    string
		cliente Cliente
		wantErr bool
	}{
		{
			name: "cliente valido",
			cliente: Cliente{
				Nome:    "Mario",
				Cognome: "Rossi",
			},
			wantErr: false,
		},
		{
			name: "nome vuoto",
			cliente: Cliente{
				Nome:    "",
				Cognome: "Rossi",
			},
			wantErr: true,
		},
		{
			name: "cognome vuoto",
			cliente: Cliente{
				Nome:    "Mario",
				Cognome: "",
			},
			wantErr: true,
		},
		{
			name: "nome con spazi",
			cliente: Cliente{
				Nome:    "   ",
				Cognome: "Rossi",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cliente.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cliente.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClienteFullName(t *testing.T) {
	c := Cliente{
		Nome:    "Mario",
		Cognome: "Rossi",
	}

	expected := "Mario Rossi"
	if got := c.FullName(); got != expected {
		t.Errorf("Cliente.FullName() = %v, want %v", got, expected)
	}
}

func TestVeicoloValidate(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name    string
		veicolo Veicolo
		wantErr bool
	}{
		{
			name: "veicolo valido",
			veicolo: Veicolo{
				Targa:     "AB123CD",
				Marca:     "Fiat",
				Modello:   "Panda",
				Anno:      2020,
				ClienteID: 1,
			},
			wantErr: false,
		},
		{
			name: "targa vuota",
			veicolo: Veicolo{
				Targa:     "",
				Marca:     "Fiat",
				ClienteID: 1,
				Anno:      2020,
			},
			wantErr: true,
		},
		{
			name: "marca vuota",
			veicolo: Veicolo{
				Targa:     "AB123CD",
				Marca:     "",
				ClienteID: 1,
				Anno:      2020,
			},
			wantErr: true,
		},
		{
			name: "cliente_id invalido",
			veicolo: Veicolo{
				Targa: "AB123CD",
				Marca: "Fiat",
				Anno:  2020,
			},
			wantErr: true,
		},
		{
			name: "anno troppo vecchio",
			veicolo: Veicolo{
				Targa:     "AB123CD",
				Marca:     "Fiat",
				ClienteID: 1,
				Anno:      1899,
			},
			wantErr: true,
		},
		{
			name: "anno futuro valido",
			veicolo: Veicolo{
				Targa:     "AB123CD",
				Marca:     "Fiat",
				ClienteID: 1,
				Anno:      currentYear + 1,
			},
			wantErr: false,
		},
		{
			name: "anno troppo nel futuro",
			veicolo: Veicolo{
				Targa:     "AB123CD",
				Marca:     "Fiat",
				ClienteID: 1,
				Anno:      currentYear + 2,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.veicolo.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Veicolo.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommessaValidate(t *testing.T) {
	tests := []struct {
		name     string
		commessa Commessa
		wantErr  bool
	}{
		{
			name: "commessa valida aperta",
			commessa: Commessa{
				VeicoloID:       1,
				Stato:           StatoCommessaAperta,
				CostoManodopera: 100.0,
				CostoRicambi:    50.0,
			},
			wantErr: false,
		},
		{
			name: "commessa valida chiusa",
			commessa: Commessa{
				VeicoloID:       1,
				Stato:           StatoCommessaChiusa,
				CostoManodopera: 100.0,
				CostoRicambi:    50.0,
			},
			wantErr: false,
		},
		{
			name: "veicolo_id invalido",
			commessa: Commessa{
				VeicoloID:       0,
				Stato:           StatoCommessaAperta,
				CostoManodopera: 100.0,
			},
			wantErr: true,
		},
		{
			name: "stato invalido",
			commessa: Commessa{
				VeicoloID: 1,
				Stato:     "InCorso",
			},
			wantErr: true,
		},
		{
			name: "costo manodopera negativo",
			commessa: Commessa{
				VeicoloID:       1,
				Stato:           StatoCommessaAperta,
				CostoManodopera: -10.0,
			},
			wantErr: true,
		},
		{
			name: "costo ricambi negativo",
			commessa: Commessa{
				VeicoloID:    1,
				Stato:        StatoCommessaAperta,
				CostoRicambi: -5.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.commessa.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Commessa.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommessaCalculateTotal(t *testing.T) {
	c := &Commessa{
		CostoManodopera: 100.0,
		CostoRicambi:    50.0,
	}

	c.CalculateTotal()

	expected := 150.0
	if c.Totale != expected {
		t.Errorf("Commessa.CalculateTotal() Totale = %v, want %v", c.Totale, expected)
	}
}

func TestCommessaIsOpen(t *testing.T) {
	tests := []struct {
		name     string
		commessa Commessa
		want     bool
	}{
		{
			name:     "commessa aperta",
			commessa: Commessa{Stato: StatoCommessaAperta},
			want:     true,
		},
		{
			name:     "commessa chiusa",
			commessa: Commessa{Stato: StatoCommessaChiusa},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.commessa.IsOpen(); got != tt.want {
				t.Errorf("Commessa.IsOpen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMovimentoPrimaNotaValidate(t *testing.T) {
	tests := []struct {
		name      string
		movimento MovimentoPrimaNota
		wantErr   bool
	}{
		{
			name: "movimento valido entrata",
			movimento: MovimentoPrimaNota{
				Tipo:    TipoMovimentoEntrata,
				Importo: 100.0,
				Metodo:  MetodoPagamentoCassa,
			},
			wantErr: false,
		},
		{
			name: "movimento valido uscita",
			movimento: MovimentoPrimaNota{
				Tipo:    TipoMovimentoUscita,
				Importo: 50.0,
				Metodo:  MetodoPagamentoBonifico,
			},
			wantErr: false,
		},
		{
			name: "tipo invalido",
			movimento: MovimentoPrimaNota{
				Tipo:    "Altro",
				Importo: 100.0,
				Metodo:  MetodoPagamentoCassa,
			},
			wantErr: true,
		},
		{
			name: "importo zero",
			movimento: MovimentoPrimaNota{
				Tipo:    TipoMovimentoEntrata,
				Importo: 0,
				Metodo:  MetodoPagamentoCassa,
			},
			wantErr: true,
		},
		{
			name: "importo negativo",
			movimento: MovimentoPrimaNota{
				Tipo:    TipoMovimentoEntrata,
				Importo: -10.0,
				Metodo:  MetodoPagamentoCassa,
			},
			wantErr: true,
		},
		{
			name: "metodo invalido",
			movimento: MovimentoPrimaNota{
				Tipo:    TipoMovimentoEntrata,
				Importo: 100.0,
				Metodo:  "CARTA_CREDITO",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.movimento.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MovimentoPrimaNota.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
