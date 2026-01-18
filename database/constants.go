package database

// Stati commessa
const (
	StatoCommessaAperta = "Aperta"
	StatoCommessaChiusa = "Chiusa"
)

// Tipi movimento prima nota
const (
	TipoMovimentoEntrata = "Entrata"
	TipoMovimentoUscita  = "Uscita"
)

// Metodi di pagamento
const (
	MetodoPagamentoCassa    = "CASSA"
	MetodoPagamentoBanca    = "BANCA"
	MetodoPagamentoPOS      = "POS"
	MetodoPagamentoAssegno  = "ASSEGNO"
	MetodoPagamentoBonifico = "BONIFICO"
)

// ValidMetodiPagamento restituisce la lista dei metodi di pagamento validi
func ValidMetodiPagamento() []string {
	return []string{
		MetodoPagamentoCassa,
		MetodoPagamentoBanca,
		MetodoPagamentoPOS,
		MetodoPagamentoAssegno,
		MetodoPagamentoBonifico,
	}
}

// IsValidMetodoPagamento verifica se un metodo di pagamento è valido
func IsValidMetodoPagamento(metodo string) bool {
	for _, m := range ValidMetodiPagamento() {
		if m == metodo {
			return true
		}
	}
	return false
}

// ValidStatiCommessa restituisce la lista degli stati commessa validi
func ValidStatiCommessa() []string {
	return []string{
		StatoCommessaAperta,
		StatoCommessaChiusa,
	}
}

// IsValidStatoCommessa verifica se uno stato commessa è valido
func IsValidStatoCommessa(stato string) bool {
	for _, s := range ValidStatiCommessa() {
		if s == stato {
			return true
		}
	}
	return false
}

// ValidTipiMovimento restituisce la lista dei tipi movimento validi
func ValidTipiMovimento() []string {
	return []string{
		TipoMovimentoEntrata,
		TipoMovimentoUscita,
	}
}

// IsValidTipoMovimento verifica se un tipo movimento è valido
func IsValidTipoMovimento(tipo string) bool {
	for _, t := range ValidTipiMovimento() {
		if t == tipo {
			return true
		}
	}
	return false
}

// Ruoli operatore comuni
var RuoliOperatoreComuni = []string{
	"Meccanico",
	"Carrozziere",
	"Elettrauto",
	"Gommista",
	"Responsabile Officina",
	"Addetto Accettazione",
}
