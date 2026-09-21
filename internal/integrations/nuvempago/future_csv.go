package nuvempago

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"majucau.local/financial-intelligence/internal/domain"
)

// FutureSourceEntity is the stable contract name for the controlled export
// placed in 02_nuvem_pago/recebimentos_futuros. It is deliberately different
// from the Bling receipts entity: this file represents a projected right to
// receive, never money already received.
const FutureSourceEntity = "nuvem_pago_recebimentos_futuros_csv_v1"

var (
	ErrMissingFutureColumn = errors.New("nuvem pago future column is missing")
	ErrInvalidFutureHeader = errors.New("invalid nuvem pago future export header")
)

// FutureReceivableCandidate is the normalized, non-persistent representation
// of one Nuvem Pago future-export row. Monetary deductions are normalized to
// non-negative amounts for the receivables schema; the original signed values
// remain available in the source file that will be stored as RAW by the next
// persistence increment.
type FutureReceivableCandidate struct {
	SourceSystem          domain.Origin
	SourceEntity          string
	SourceID              string
	CustomerName          string
	PaymentMethod         string
	Brand                 string
	PaymentDate           time.Time
	ExpectedReceiptDate   time.Time
	InstallmentNumber     *int
	InstallmentCount      *int
	GrossAmount           domain.Money
	FeeAmount             domain.Money
	FeeAmountSigned       domain.Money
	InterestAmount        domain.Money
	InterestAmountSigned  domain.Money
	TotalCostAmount       domain.Money
	TotalCostAmountSigned domain.Money
	NetAmount             domain.Money
	Status                domain.DataStatus
}

type RowError struct {
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FutureReport struct {
	Receivables []FutureReceivableCandidate
	Errors      []RowError
}

// ParseFutureCSV parses the semicolon-delimited Nuvem Pago future-receivables
// export supplied by the operator. It never emits a confirmed receipt and it
// does not calculate a tariff from the pricing table: exported tax, costs and
// net value are preserved as facts from the file.
func ParseFutureCSV(input io.Reader) (FutureReport, error) {
	reader := csv.NewReader(input)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return FutureReport{}, ErrInvalidFutureHeader
		}
		return FutureReport{}, err
	}
	columns, err := requiredFutureColumns(header)
	if err != nil {
		return FutureReport{}, err
	}

	result := FutureReport{}
	seen := make(map[string]int)
	for line := 2; ; line++ {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			result.Errors = append(result.Errors, RowError{Line: line, Code: "CSV_READ", Message: "não foi possível ler a linha do extrato"})
			continue
		}
		candidate, rowErr := parseFutureRow(row, columns)
		if rowErr != nil {
			result.Errors = append(result.Errors, RowError{Line: line, Code: rowErr.code, Message: rowErr.message})
			continue
		}
		if previousLine, exists := seen[candidate.SourceID]; exists {
			result.Errors = append(result.Errors, RowError{Line: line, Code: "DUPLICATE_SOURCE_ID", Message: fmt.Sprintf("transação repetida; primeira ocorrência na linha %d", previousLine)})
			continue
		}
		seen[candidate.SourceID] = line
		result.Receivables = append(result.Receivables, candidate)
	}
	return result, nil
}

type rowParseError struct{ code, message string }

func parseFutureRow(row []string, columns map[string]int) (FutureReceivableCandidate, *rowParseError) {
	value := func(name string) string {
		index := columns[name]
		if index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}
	if !strings.EqualFold(value("movement_type"), "entrada") {
		return FutureReceivableCandidate{}, &rowParseError{code: "NOT_INFLOW", message: "linha não importada porque o tipo de movimentação não é entrada"}
	}
	if !strings.EqualFold(value("transaction_type"), "venda") {
		return FutureReceivableCandidate{}, &rowParseError{code: "NOT_SALE", message: "linha não importada porque o tipo de transação não é venda"}
	}
	sourceID := value("source_id")
	if sourceID == "" {
		return FutureReceivableCandidate{}, &rowParseError{code: "MISSING_SOURCE_ID", message: "número da transação não informado"}
	}
	paymentDate, err := parseFutureDate(value("payment_date"))
	if err != nil {
		return FutureReceivableCandidate{}, &rowParseError{code: "INVALID_PAYMENT_DATE", message: "data do pagamento inválida"}
	}
	expectedDate, err := parseFutureDate(value("expected_receipt_date"))
	if err != nil {
		return FutureReceivableCandidate{}, &rowParseError{code: "INVALID_EXPECTED_RECEIPT_DATE", message: "data de recebimento inválida"}
	}
	if expectedDate.Before(paymentDate) {
		return FutureReceivableCandidate{}, &rowParseError{code: "RECEIPT_DATE_BEFORE_PAYMENT", message: "data prevista de recebimento anterior ao pagamento"}
	}
	amounts := make(map[string]domain.Money, 5)
	signedAmounts := make(map[string]domain.Money, 5)
	for _, name := range []string{"gross", "fee", "interest", "costs", "net"} {
		parsed, parseErr := parseFutureMoney(value(name))
		if parseErr != nil {
			return FutureReceivableCandidate{}, &rowParseError{code: "INVALID_" + strings.ToUpper(name), message: "valor monetário inválido"}
		}
		signedAmounts[name] = parsed
		if parsed.IsNegative() {
			parsed, parseErr = parsed.AbsExact()
			if parseErr != nil {
				return FutureReceivableCandidate{}, &rowParseError{code: "INVALID_" + strings.ToUpper(name), message: "valor monetário fora do limite"}
			}
		}
		amounts[name] = parsed
	}
	if amounts["gross"].IsNegative() || amounts["net"].IsNegative() {
		return FutureReceivableCandidate{}, &rowParseError{code: "NEGATIVE_AMOUNT", message: "bruto e líquido devem ser não negativos"}
	}
	installmentCount, err := parseOptionalPositiveInt(value("installment_count"))
	if err != nil {
		return FutureReceivableCandidate{}, &rowParseError{code: "INVALID_INSTALLMENTS", message: "número de parcelas inválido"}
	}
	return FutureReceivableCandidate{
		SourceSystem: domain.OriginNuvemPago, SourceEntity: FutureSourceEntity, SourceID: sourceID,
		CustomerName: value("customer"), PaymentMethod: value("payment_method"), Brand: value("brand"),
		PaymentDate: paymentDate, ExpectedReceiptDate: expectedDate,
		InstallmentCount: installmentCount, GrossAmount: amounts["gross"], FeeAmount: amounts["fee"],
		FeeAmountSigned: signedAmounts["fee"], InterestAmount: amounts["interest"], InterestAmountSigned: signedAmounts["interest"],
		TotalCostAmount: amounts["costs"], TotalCostAmountSigned: signedAmounts["costs"], NetAmount: amounts["net"],
		Status: domain.StatusProjected,
	}, nil
}

func requiredFutureColumns(header []string) (map[string]int, error) {
	aliases := map[string][]string{
		"payment_date":          {"data do pagamento"},
		"expected_receipt_date": {"data de recebimento"},
		"movement_type":         {"tipo de movimentação", "tipo de movimentacao"},
		"transaction_type":      {"tipo de transação", "tipo de transacao"},
		"source_id":             {"nº transação", "no transação", "nº transacao", "no transacao"},
		"customer":              {"nome"},
		"payment_method":        {"forma de pagamento"},
		"brand":                 {"bandeira"},
		"installment_count":     {"nº parcelas", "no parcelas"},
		"gross":                 {"valor bruto (r$)", "valor bruto"},
		"fee":                   {"taxas (r$)", "taxas", "taxa"},
		"interest":              {"juros (r$)", "juros"},
		"costs":                 {"custos totais (r$)", "custos totais"},
		"net":                   {"valor liquido (r$)", "valor líquido (r$)", "valor liquido", "valor líquido"},
	}
	normalized := make(map[string]int, len(header))
	for index, name := range header {
		normalized[canonicalFutureHeader(name)] = index
	}
	columns := make(map[string]int, len(aliases))
	for name, options := range aliases {
		for _, option := range options {
			if index, ok := normalized[canonicalFutureHeader(option)]; ok {
				columns[name] = index
				break
			}
		}
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingFutureColumn, name)
		}
	}
	return columns, nil
}

func canonicalFutureHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")))
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä':
			r = 'a'
		case 'é', 'è', 'ê', 'ë':
			r = 'e'
		case 'í', 'ì', 'î', 'ï':
			r = 'i'
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			r = 'o'
		case 'ú', 'ù', 'û', 'ü':
			r = 'u'
		case 'ç':
			r = 'c'
		case 'º', '°':
			r = 'o'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '(' || r == ')' || r == '$' {
			builder.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func parseFutureDate(value string) (time.Time, error) {
	for _, layout := range []string{"02/01/2006", "02/01/2006 15:04", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, strings.TrimSpace(value), time.UTC); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid date")
}

func parseFutureMoney(value string) (domain.Money, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "R$"))
	value = strings.ReplaceAll(value, " ", "")
	if value == "" {
		return domain.Money{}, domain.ErrInvalidMoney
	}
	negative := strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")")
	if negative {
		value = "-" + strings.TrimSuffix(strings.TrimPrefix(value, "("), ")")
	}
	if strings.Contains(value, ",") {
		value = strings.ReplaceAll(value, ".", "")
		value = strings.Replace(value, ",", ".", 1)
	}
	return domain.ParseMoney(value)
}

func parseOptionalPositiveInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed < 1 {
		return nil, errors.New("invalid installment count")
	}
	return &parsed, nil
}
