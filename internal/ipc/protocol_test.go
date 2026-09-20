package ipc

import "testing"

func TestProtocolRoundTripAndValidation(t *testing.T) {
	r := Request{Version: ProtocolVersion, RequestID: "abc", Method: MethodHealth}
	b, err := EncodeRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRequest(b)
	if err != nil || got.Method != MethodHealth {
		t.Fatalf("roundtrip: %#v %v", got, err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"2","request_id":"x","method":"health.get"}`)); err == nil {
		t.Fatal("version mismatch must fail")
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"unknown"}`)); err == nil {
		t.Fatal("unknown method must fail")
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.receipts.preview","payload":{"folder":"C:\\imports"}}`)); err != nil {
		t.Fatalf("Bling preview method must be accepted: %v", err)
	}
}
