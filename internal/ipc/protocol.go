package ipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const ProtocolVersion = "1"
const MaxMessageSize = 1 << 20

var (
	ErrInvalidRequest     = errors.New("invalid ipc request")
	ErrUnsupportedVersion = errors.New("unsupported ipc protocol version")
	ErrUnsupportedMethod  = errors.New("unsupported ipc method")
)

type Method string

const (
	MethodHealth                 Method = "health.get"
	MethodIntegrationStatus      Method = "integration.status"
	MethodBlingConfigSave        Method = "bling.config.save"
	MethodNuvemshopConfigSave    Method = "nuvemshop.config.save"
	MethodBlingOAuthStart        Method = "bling.oauth.start"
	MethodBlingOAuthStatus       Method = "bling.oauth.status"
	MethodBlingOAuthTest         Method = "bling.oauth.test"
	MethodBlingSync              Method = "bling.sync"
	MethodBlingReceiptsPreview   Method = "bling.receipts.preview"
	MethodBlingReceiptsImport    Method = "bling.receipts.import"
	MethodNuvemPagoFuturePreview Method = "nuvem_pago.future.preview"
	MethodNuvemPagoFutureImport  Method = "nuvem_pago.future.import"
	MethodDashboardSnapshot      Method = "dashboard.snapshot"
)

type Request struct {
	Version   string          `json:"version"`
	RequestID string          `json:"request_id"`
	Method    Method          `json:"method"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}
type Response struct {
	Version   string          `json:"version"`
	RequestID string          `json:"request_id"`
	OK        bool            `json:"ok"`
	Error     *ErrorBody      `json:"error,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r Request) Validate() error {
	if r.Version != ProtocolVersion {
		return fmt.Errorf("%w: %s", ErrUnsupportedVersion, r.Version)
	}
	if strings.TrimSpace(r.RequestID) == "" || len(r.RequestID) > 128 {
		return ErrInvalidRequest
	}
	if r.Method != MethodHealth && r.Method != MethodIntegrationStatus && r.Method != MethodBlingConfigSave && r.Method != MethodNuvemshopConfigSave && r.Method != MethodBlingOAuthStart && r.Method != MethodBlingOAuthStatus && r.Method != MethodBlingOAuthTest && r.Method != MethodBlingSync && r.Method != MethodBlingReceiptsPreview && r.Method != MethodBlingReceiptsImport && r.Method != MethodNuvemPagoFuturePreview && r.Method != MethodNuvemPagoFutureImport && r.Method != MethodDashboardSnapshot {
		return fmt.Errorf("%w: %s", ErrUnsupportedMethod, r.Method)
	}
	if len(r.Payload) > MaxMessageSize {
		return ErrInvalidRequest
	}
	if len(r.Payload) > 0 && !json.Valid(r.Payload) {
		return ErrInvalidRequest
	}
	return nil
}

func EncodeRequest(r Request) ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxMessageSize {
		return nil, ErrInvalidRequest
	}
	return b, nil
}
func DecodeRequest(data []byte) (Request, error) {
	if len(data) == 0 || len(data) > MaxMessageSize || bytes.IndexByte(data, 0) >= 0 {
		return Request{}, ErrInvalidRequest
	}
	var r Request
	if err := json.Unmarshal(data, &r); err != nil {
		return Request{}, ErrInvalidRequest
	}
	if err := r.Validate(); err != nil {
		return Request{}, err
	}
	return r, nil
}
func NewResponse(id string, payload any) (Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}
	return Response{Version: ProtocolVersion, RequestID: id, OK: true, Payload: b}, nil
}
func NewErrorResponse(id, code, message string) Response {
	return Response{Version: ProtocolVersion, RequestID: id, OK: false, Error: &ErrorBody{Code: code, Message: message}}
}

