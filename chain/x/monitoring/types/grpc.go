package types

// Query service request/response types

type QueryNetworkHealthRequest struct{}

type QueryNetworkHealthResponse struct {
	Health *NetworkHealth `json:"health"`
}

type QueryValidatorUptimeRequest struct {
	ValidatorAddress string `json:"validator_address"`
}

type QueryValidatorUptimeResponse struct {
	Uptime *ValidatorUptime `json:"uptime"`
}

type QueryAlertsRequest struct {
	Severity string `json:"severity,omitempty"`
	Type     string `json:"type,omitempty"`
}

type QueryAlertsResponse struct {
	Alerts []*Alert `json:"alerts"`
}

type QueryGasPriceRequest struct{}

type QueryGasPriceResponse struct {
	Tracking *GasPriceTracking `json:"tracking"`
}

type QueryTVLRequest struct{}

type QueryTVLResponse struct {
	Tvl *TVLMonitoring `json:"tvl"`
}

// Message service request/response types

type MsgAcknowledgeAlert struct {
	AlertId        string `json:"alert_id"`
	AcknowledgedBy string `json:"acknowledged_by"`
}

type MsgAcknowledgeAlertResponse struct {
	Success bool `json:"success"`
}

type MsgResolveAlert struct {
	AlertId string `json:"alert_id"`
}

type MsgResolveAlertResponse struct {
	Success bool `json:"success"`
}

// RegisterQueryServer registers the query server (stub for compilation)
func RegisterQueryServer(server interface{}, impl interface{}) {
	// In a real implementation, this would register with the gRPC server
}

// RegisterMsgServer registers the message server (stub for compilation)
func RegisterMsgServer(server interface{}, impl interface{}) {
	// In a real implementation, this would register with the gRPC server
}
