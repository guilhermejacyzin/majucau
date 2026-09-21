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
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.receipts.import","payload":{"folder":"C:\\imports"}}`)); err != nil {
		t.Fatalf("Bling import method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"nuvem_pago.future.preview","payload":{"folder":"C:\\imports"}}`)); err != nil {
		t.Fatalf("Nuvem Pago future preview method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"nuvem_pago.future.import","payload":{"folder":"C:\\imports"}}`)); err != nil {
		t.Fatalf("Nuvem Pago future import method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.config.save","payload":{"client_id":"client","redirect_uri":"https://app.example.test/callback","client_secret":"secret"}}`)); err != nil {
		t.Fatalf("Bling config method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.oauth.start"}`)); err != nil {
		t.Fatalf("Bling OAuth start method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.oauth.status","payload":{"session_id":"session"}}`)); err != nil {
		t.Fatalf("Bling OAuth status method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.oauth.test","payload":{}}`)); err != nil {
		t.Fatalf("Bling OAuth test method must be accepted: %v", err)
	}
	if _, err := DecodeRequest([]byte(`{"version":"1","request_id":"x","method":"bling.sync","payload":{"received_date_from":"2026-09-01","received_date_to":"2026-09-30"}}`)); err != nil {
		t.Fatalf("Bling sync method must be accepted: %v", err)
	}
}

