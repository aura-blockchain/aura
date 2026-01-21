package compliance

import (
	"context"
	"fmt"

	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"github.com/aura-chain/aura/sdk/go/pkg/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the compliance module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient compliancepb.QueryClient
	msgClient   compliancepb.MsgClient
}

// NewClient creates a new compliance client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: compliancepb.NewQueryClient(grpcConn),
		msgClient:   compliancepb.NewMsgClient(grpcConn),
	}
}

// SubmitKYCParams contains parameters for submitting KYC data
type SubmitKYCParams struct {
	Address       string
	KYCLevel      compliancepb.KYCLevel
	Provider      string
	PIICommitment []byte
	Jurisdiction  string
}

// SubmitKYC submits KYC data for an address
func (c *Client) SubmitKYC(ctx context.Context, params *SubmitKYCParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if params.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if len(params.PIICommitment) == 0 {
		return nil, fmt.Errorf("PII commitment is required")
	}
	if params.Jurisdiction == "" {
		return nil, fmt.Errorf("jurisdiction is required")
	}

	msg := &compliancepb.MsgSubmitKYC{
		Address:       params.Address,
		KycLevel:      params.KYCLevel,
		Provider:      params.Provider,
		PiiCommitment: params.PIICommitment,
		Jurisdiction:  params.Jurisdiction,
	}

	addr, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to submit KYC: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// ReportSuspiciousActivityParams contains parameters for reporting suspicious activity
type ReportSuspiciousActivityParams struct {
	Reporter        string
	Address         string
	TransactionHash string
	ActivityType    string
	Description     string
	Indicators      []string
}

// ReportSuspiciousActivity reports suspicious activity for AML compliance
func (c *Client) ReportSuspiciousActivity(ctx context.Context, params *ReportSuspiciousActivityParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Reporter == "" {
		return nil, fmt.Errorf("reporter is required")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if params.ActivityType == "" {
		return nil, fmt.Errorf("activity type is required")
	}

	msg := &compliancepb.MsgReportSuspiciousActivity{
		Reporter:        params.Reporter,
		Address:         params.Address,
		TransactionHash: params.TransactionHash,
		ActivityType:    params.ActivityType,
		Description:     params.Description,
		Indicators:      params.Indicators,
	}

	addr, err := sdk.AccAddressFromBech32(params.Reporter)
	if err != nil {
		return nil, fmt.Errorf("invalid reporter address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to report suspicious activity: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// ScreenSanctionsParams contains parameters for sanctions screening
type ScreenSanctionsParams struct {
	Address      string
	ForceRefresh bool
}

// ScreenSanctions performs sanctions screening for an address
func (c *Client) ScreenSanctions(ctx context.Context, params *ScreenSanctionsParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	msg := &compliancepb.MsgScreenSanctions{
		Address:      params.Address,
		ForceRefresh: params.ForceRefresh,
	}

	addr, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to screen sanctions: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// RecordGDPRConsentParams contains parameters for recording GDPR consent
type RecordGDPRConsentParams struct {
	Address        string
	ConsentType    string
	Consented      bool
	ConsentVersion string
}

// RecordGDPRConsent records GDPR consent
func (c *Client) RecordGDPRConsent(ctx context.Context, params *RecordGDPRConsentParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if params.ConsentType == "" {
		return nil, fmt.Errorf("consent type is required")
	}

	msg := &compliancepb.MsgRecordGDPRConsent{
		Address:        params.Address,
		ConsentType:    params.ConsentType,
		Consented:      params.Consented,
		ConsentVersion: params.ConsentVersion,
	}

	addr, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to record GDPR consent: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// RequestGDPRData requests GDPR data export
func (c *Client) RequestGDPRData(ctx context.Context, address string, requestType string) (*types.TxResponse, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if requestType == "" {
		return nil, fmt.Errorf("request type is required")
	}

	msg := &compliancepb.MsgRequestGDPRData{
		Address:     address,
		RequestType: requestType,
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to request GDPR data: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// EraseGDPRData requests GDPR data erasure (right to be forgotten)
func (c *Client) EraseGDPRData(ctx context.Context, address string, erasureReason string) (*types.TxResponse, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if erasureReason == "" {
		return nil, fmt.Errorf("erasure reason is required")
	}

	msg := &compliancepb.MsgEraseGDPRData{
		Address:       address,
		ErasureReason: erasureReason,
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to erase GDPR data: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// GenerateTaxReportParams contains parameters for generating a tax report
type GenerateTaxReportParams struct {
	Address      string
	TaxYear      string
	Jurisdiction string
	ReportType   string
	FilePath     string
}

// GenerateTaxReport generates a tax report for an address
func (c *Client) GenerateTaxReport(ctx context.Context, params *GenerateTaxReportParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if params.TaxYear == "" {
		return nil, fmt.Errorf("tax year is required")
	}

	msg := &compliancepb.MsgGenerateTaxReport{
		Address:      params.Address,
		TaxYear:      params.TaxYear,
		Jurisdiction: params.Jurisdiction,
		ReportType:   params.ReportType,
		FilePath:     params.FilePath,
	}

	addr, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tax report: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// GetKYCRecord retrieves a KYC record for an address
func (c *Client) GetKYCRecord(ctx context.Context, address string) (*compliancepb.KYCRecord, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &compliancepb.QueryKYCRecordRequest{
		Address: address,
	}

	resp, err := c.queryClient.KycRecord(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get KYC record: %w", err)
	}

	return resp.Record, nil
}

// GetKYCHistory retrieves KYC history for an address
func (c *Client) GetKYCHistory(ctx context.Context, address string) ([]*compliancepb.KYCHistoryEntry, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &compliancepb.QueryKYCHistoryRequest{
		Address: address,
	}

	resp, err := c.queryClient.KycHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get KYC history: %w", err)
	}

	return resp.History, nil
}

// GetAMLProfile retrieves an AML profile for an address
func (c *Client) GetAMLProfile(ctx context.Context, address string) (*compliancepb.AMLProfile, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &compliancepb.QueryAMLProfileRequest{
		Address: address,
	}

	resp, err := c.queryClient.AmlProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get AML profile: %w", err)
	}

	return resp.Profile, nil
}

// GetSanctionsScreening retrieves sanctions screening results for an address
func (c *Client) GetSanctionsScreening(ctx context.Context, address string) (*compliancepb.SanctionsScreeningResult, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &compliancepb.QuerySanctionsScreeningRequest{
		Address: address,
	}

	resp, err := c.queryClient.SanctionsScreening(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get sanctions screening: %w", err)
	}

	return resp.Result, nil
}

// GetTransactionAlerts retrieves transaction alerts for an address
func (c *Client) GetTransactionAlerts(ctx context.Context, address string, unreviewedOnly bool) ([]*compliancepb.TransactionAlert, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &compliancepb.QueryTransactionAlertsRequest{
		Address:        address,
		UnreviewedOnly: unreviewedOnly,
	}

	resp, err := c.queryClient.TransactionAlerts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction alerts: %w", err)
	}

	return resp.Alerts, nil
}

// GetTaxReport retrieves a tax report
func (c *Client) GetTaxReport(ctx context.Context, address string, taxYear string) (*compliancepb.TaxReport, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if taxYear == "" {
		return nil, fmt.Errorf("tax year is required")
	}

	req := &compliancepb.QueryTaxReportRequest{
		Address: address,
		TaxYear: taxYear,
	}

	resp, err := c.queryClient.TaxReport(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax report: %w", err)
	}

	return resp.Report, nil
}

// ListAllKYCRecords lists all KYC records with pagination
func (c *Client) ListAllKYCRecords(ctx context.Context) ([]*compliancepb.KYCRecord, error) {
	req := &compliancepb.QueryAllKYCRecordsRequest{}

	resp, err := c.queryClient.AllKYCRecords(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list KYC records: %w", err)
	}

	return resp.Records, nil
}

// ListAllAMLProfiles lists all AML profiles with pagination
func (c *Client) ListAllAMLProfiles(ctx context.Context) ([]*compliancepb.AMLProfile, error) {
	req := &compliancepb.QueryAllAMLProfilesRequest{}

	resp, err := c.queryClient.AllAMLProfiles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list AML profiles: %w", err)
	}

	return resp.Profiles, nil
}

// ListAllSanctionsResults lists all sanctions screening results with pagination
func (c *Client) ListAllSanctionsResults(ctx context.Context) ([]*compliancepb.SanctionsScreeningResult, error) {
	req := &compliancepb.QueryAllSanctionsResultsRequest{}

	resp, err := c.queryClient.AllSanctionsResults(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list sanctions results: %w", err)
	}

	return resp.Results, nil
}

// ListAllTransactionAlerts lists all transaction alerts with pagination
func (c *Client) ListAllTransactionAlerts(ctx context.Context) ([]*compliancepb.TransactionAlertList, error) {
	req := &compliancepb.QueryAllTransactionAlertsRequest{}

	resp, err := c.queryClient.AllTransactionAlerts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list transaction alerts: %w", err)
	}

	return resp.Alerts, nil
}

// ListAllGDPRConsents lists all GDPR consents with pagination
func (c *Client) ListAllGDPRConsents(ctx context.Context) ([]*compliancepb.GDPRConsentList, error) {
	req := &compliancepb.QueryAllGDPRConsentsRequest{}

	resp, err := c.queryClient.AllGDPRConsents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list GDPR consents: %w", err)
	}

	return resp.Consents, nil
}

// ListAllTaxReports lists all tax reports with pagination
func (c *Client) ListAllTaxReports(ctx context.Context) ([]*compliancepb.TaxReportList, error) {
	req := &compliancepb.QueryAllTaxReportsRequest{}

	resp, err := c.queryClient.AllTaxReports(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tax reports: %w", err)
	}

	return resp.Reports, nil
}

// ListAllGDPRRequests lists all GDPR requests with pagination
func (c *Client) ListAllGDPRRequests(ctx context.Context) ([]*compliancepb.GDPRDataRequest, error) {
	req := &compliancepb.QueryAllGDPRRequestsRequest{}

	resp, err := c.queryClient.AllGDPRRequests(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list GDPR requests: %w", err)
	}

	return resp.Requests, nil
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*compliancepb.ComplianceParams, error) {
	req := &compliancepb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return &resp.Params, nil
}
