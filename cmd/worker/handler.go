package main

import (
	"context"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/ipc"
)

type workerHandler struct{ health application.StaticHealth }

func newWorkerHandler() workerHandler {
	return workerHandler{health: application.StaticHealth{Service: "majucau-worker", Version: "dev"}}
}
func (h workerHandler) Handle(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	switch req.Method {
	case ipc.MethodHealth:
		return ipc.NewResponse(req.RequestID, h.health.CheckHealth(ctx))
	case ipc.MethodIntegrationStatus:
		return ipc.NewResponse(req.RequestID, []application.IntegrationStatus{
			{Provider: domain.OriginBling, Status: domain.IntegrationNotConfigured},
			{Provider: domain.OriginNuvemshop, Status: domain.IntegrationNotConfigured},
			{Provider: domain.OriginNuvemPago, Status: domain.IntegrationUnavailable},
		})
	default:
		return ipc.Response{}, ipc.ErrUnsupportedMethod
	}
}
