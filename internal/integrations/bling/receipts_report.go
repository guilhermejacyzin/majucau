package bling

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

const ReceiptsReportSourceEntity = "bling_contas_recebidas_report_v1"

var (
	ErrMissingReportColumn = errors.New("bling report column is missing")
	ErrInvalidReportHeader = errors.New("invalid bling receipts report header")
)

type ReceiptCandidate struct {
	SourceSystem     domain.Origin
	SourceEntity     string
	SourceID         string
	CustomerName     string
	History          string
	PaymentMethod    string
	ExternalDocument string
	DueDate          time.Time
	ReceiptDate      time.Time
	Fee              domain.Money
	Amount           domain.Money
	Status           domain.DataStatus
}

type RowError struct {
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ReceiptsReport struct {
	Receipts []ReceiptCandidate
	Errors   []RowError
}

// ParseReceiptsCSV parses the semicolon-delimited export of Bling's
// "Relatório de Contas Recebidas". It intentionally treats Bling as the
// source of truth even when the payment method mentions NuvemPago/Nuvemshop.
// The parser never turns an unpaid row into a receipt.
func ParseReceiptsCSV(input io.Reader) (ReceiptsReport, error) {
	reader := csv.NewReader(input)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ReceiptsReport{}, ErrInvalidReportHeader
		}
		return ReceiptsReport{}, err
	}
	columns, err := requiredColumns(header)
	if err != nil {
		return ReceiptsReport{}, err
	}

	result := ReceiptsReport{}
	seen := make(map[string]int)
	for line := 2; ; line++ {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			result.Errors = append(result.Errors, RowError{Line: line, Code: "CSV_READ", Message: "não foi possível ler a linha do relatório"})
			continue
		}
		candidate, rowErr := parseReceiptRow(row, columns)
		if rowErr != nil {
			result.Errors = append(result.Errors, RowError{Line: line, Code: rowErr.code, Message: rowErr.message})
			continue
		}
		if previousLine, exists := seen[candidate.SourceID]; exists {
			result.Errors = append(result.Errors, RowError{Line: line, Code: "DUPLICATE_SOURCE_ID", Message: fmt.Sprintf("documento repetido; primeira ocorrência na linha %d", previousLine)})
			continue
		}
		seen[candidate.SourceID] = line
		result.Receipts = append(result.Receipts, candidate)
	}
	return result, nil
}

type rowParseError struct{ code, message string }

func parseReceiptRow(row []string, columns map[string]int) (ReceiptCandidate, *rowParseError) {
	value := func(name string) string {
		index := columns[name]
		if index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}
	status := strings.ToLower(value("status"))
	if status != "pago" {
		return ReceiptCandidate{}, &rowParseError{code: "NOT_PAID", message: "linha não importada como recebimento porque a situação não é pago"}
	}
	sourceID := value("document")
	if sourceID == "" {
		return ReceiptCandidate{}, &rowParseError{code: "MISSING_SOURCE_ID", message: "número do documento não informado"}
	}
	dueDate, err := parseBlingDate(value("due_date"))
	if err != nil {
		return ReceiptCandidate{}, &rowParseError{code: "INVALID_DUE_DATE", message: "vencimento inválido"}
	}
	receiptDate, err := parseBlingDate(value("receipt_date"))
	if err != nil {
		return ReceiptCandidate{}, &rowParseError{code: "INVALID_RECEIPT_DATE", message: "data de liquidação inválida"}
	}
	fee, err := parseBrazilianMoney(value("fee"))
	if err != nil || fee.IsNegative() {
		return ReceiptCandidate{}, &rowParseError{code: "INVALID_FEE", message: "valor de taxa inválido"}
	}
	amount, err := parseBrazilianMoney(value("amount"))
	if err != nil || amount.IsNegative() {
		return ReceiptCandidate{}, &rowParseError{code: "INVALID_AMOUNT", message: "valor recebido inválido"}
	}
	return ReceiptCandidate{
		SourceSystem: domain.OriginBling, SourceEntity: ReceiptsReportSourceEntity, SourceID: sourceID,
		CustomerName: value("customer"), History: value("history"), PaymentMethod: value("payment_method"),
		ExternalDocument: sourceID, DueDate: dueDate, ReceiptDate: receiptDate, Fee: fee, Amount: amount,
		Status: domain.StatusConfirmed,
	}, nil
}

func requiredColumns(header []string) (map[string]int, error) {
	aliases := map[string][]string{
		"customer":       {"cliente"},
		"history":        {"historico"},
		"payment_method": {"forma de pagamento", "forma pagamento"},
		"document":       {"no documento", "nro documento", "numero documento", "documento"},
		"due_date":       {"vencimento"},
		"receipt_date":   {"liquidacao", "data liquidacao"},
		"status":         {"situacao", "status"},
		"fee":            {"valor taxa", "taxa"},
		"amount":         {"recebido", "valor recebido"},
	}
	normalized := make(map[string]int, len(header))
	for index, name := range header {
		normalized[canonicalHeader(name)] = index
	}
	columns := make(map[string]int, len(aliases))
	for name, options := range aliases {
		for _, option := range options {
			if index, ok := normalized[canonicalHeader(option)]; ok {
				columns[name] = index
				break
			}
		}
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingReportColumn, name)
		}
	}
	return columns, nil
}

func canonicalHeader(value string) string {
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
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			builder.WriteRune(r)
		}
	}
	value = strings.Join(strings.Fields(builder.String()), " ")
	value = strings.Replace(value, "nro ", "no ", 1)
	return value
}

func parseBlingDate(value string) (time.Time, error) {
	for _, layout := range []string{"02/01/2006", "02/01/2006 15:04", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, strings.TrimSpace(value), time.UTC); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid date")
}

func parseBrazilianMoney(value string) (domain.Money, error) {
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
