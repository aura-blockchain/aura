package types

import (
	"time"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// Supplemental types that extend or wrap proto types for keeper use
// These provide backward compatibility with the old keeper implementation

// BlacklistEntry represents a blacklisted peer or address
type BlacklistEntry struct {
	Identifier string     `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Reason     string     `protobuf:"bytes,2,opt,name=reason,proto3" json:"reason,omitempty"`
	Permanent  bool       `protobuf:"varint,3,opt,name=permanent,proto3" json:"permanent,omitempty"`
	ExpiresAt  *time.Time `protobuf:"bytes,4,opt,name=expires_at,json=expiresAt,proto3,stdtime" json:"expires_at,omitempty"`
	AddedAt    *time.Time `protobuf:"bytes,5,opt,name=added_at,json=addedAt,proto3,stdtime" json:"added_at,omitempty"`
}

// Reset implements proto.Message
func (m *BlacklistEntry) Reset() { *m = BlacklistEntry{} }

// String implements proto.Message
func (m *BlacklistEntry) String() string { return "BlacklistEntry" }

// ProtoMessage implements proto.Message
func (*BlacklistEntry) ProtoMessage() {}

// DeviceFingerprint represents a trusted device
type DeviceFingerprint struct {
	Id            string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	WalletAddress string     `protobuf:"bytes,2,opt,name=wallet_address,json=walletAddress,proto3" json:"wallet_address,omitempty"`
	Fingerprint   string     `protobuf:"bytes,3,opt,name=fingerprint,proto3" json:"fingerprint,omitempty"`
	DeviceName    string     `protobuf:"bytes,4,opt,name=device_name,json=deviceName,proto3" json:"device_name,omitempty"`
	TrustedAt     *time.Time `protobuf:"bytes,5,opt,name=trusted_at,json=trustedAt,proto3,stdtime" json:"trusted_at,omitempty"`
	LastUsed      *time.Time `protobuf:"bytes,6,opt,name=last_used,json=lastUsed,proto3,stdtime" json:"last_used,omitempty"`
	IsActive      bool       `protobuf:"varint,7,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
}

func (m *DeviceFingerprint) Reset()         { *m = DeviceFingerprint{} }
func (m *DeviceFingerprint) String() string { return "DeviceFingerprint" }
func (*DeviceFingerprint) ProtoMessage()    {}

// WalletSession represents an active wallet session
type WalletSession struct {
	Id            string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	WalletAddress string     `protobuf:"bytes,2,opt,name=wallet_address,json=walletAddress,proto3" json:"wallet_address,omitempty"`
	DeviceId      string     `protobuf:"bytes,3,opt,name=device_id,json=deviceId,proto3" json:"device_id,omitempty"`
	CreatedAt     *time.Time `protobuf:"bytes,4,opt,name=created_at,json=createdAt,proto3,stdtime" json:"created_at,omitempty"`
	ExpiresAt     *time.Time `protobuf:"bytes,5,opt,name=expires_at,json=expiresAt,proto3,stdtime" json:"expires_at,omitempty"`
	IpAddress     string     `protobuf:"bytes,6,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`
	UserAgent     string     `protobuf:"bytes,7,opt,name=user_agent,json=userAgent,proto3" json:"user_agent,omitempty"`
}

func (m *WalletSession) Reset()         { *m = WalletSession{} }
func (m *WalletSession) String() string { return "WalletSession" }
func (*WalletSession) ProtoMessage()    {}

// AnomalyDetection represents detected anomalous activity
type AnomalyDetection struct {
	Id            string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	WalletAddress string     `protobuf:"bytes,2,opt,name=wallet_address,json=walletAddress,proto3" json:"wallet_address,omitempty"`
	AnomalyType   string     `protobuf:"bytes,3,opt,name=anomaly_type,json=anomalyType,proto3" json:"anomaly_type,omitempty"`
	Description   string     `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Severity      string     `protobuf:"bytes,5,opt,name=severity,proto3" json:"severity,omitempty"`
	DetectedAt    *time.Time `protobuf:"bytes,6,opt,name=detected_at,json=detectedAt,proto3,stdtime" json:"detected_at,omitempty"`
	TxHash        string     `protobuf:"bytes,7,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	Resolved      bool       `protobuf:"varint,8,opt,name=resolved,proto3" json:"resolved,omitempty"`
}

func (m *AnomalyDetection) Reset()         { *m = AnomalyDetection{} }
func (m *AnomalyDetection) String() string { return "AnomalyDetection" }
func (*AnomalyDetection) ProtoMessage()    {}

// PauseState represents the system pause state
type PauseState struct {
	IsPaused    bool       `protobuf:"varint,1,opt,name=is_paused,json=isPaused,proto3" json:"is_paused,omitempty"`
	PauseLevel  uint32     `protobuf:"varint,2,opt,name=pause_level,json=pauseLevel,proto3" json:"pause_level,omitempty"`
	PausedAt    *time.Time `protobuf:"bytes,3,opt,name=paused_at,json=pausedAt,proto3,stdtime" json:"paused_at,omitempty"`
	PausedBy    string     `protobuf:"bytes,4,opt,name=paused_by,json=pausedBy,proto3" json:"paused_by,omitempty"`
	Reason      string     `protobuf:"bytes,5,opt,name=reason,proto3" json:"reason,omitempty"`
	ResumeAfter *time.Time `protobuf:"bytes,6,opt,name=resume_after,json=resumeAfter,proto3,stdtime" json:"resume_after,omitempty"`
}

func (m *PauseState) Reset()         { *m = PauseState{} }
func (m *PauseState) String() string { return "PauseState" }
func (*PauseState) ProtoMessage()    {}

// WalletLimit defines transaction limits for a wallet during incidents
type WalletLimit struct {
	WalletAddress   string     `protobuf:"bytes,1,opt,name=wallet_address,json=walletAddress,proto3" json:"wallet_address,omitempty"`
	MaxTxAmount     string     `protobuf:"bytes,2,opt,name=max_tx_amount,json=maxTxAmount,proto3" json:"max_tx_amount,omitempty"`
	MaxDailyTxs     uint32     `protobuf:"varint,3,opt,name=max_daily_txs,json=maxDailyTxs,proto3" json:"max_daily_txs,omitempty"`
	CooldownPeriod  string     `protobuf:"bytes,4,opt,name=cooldown_period,json=cooldownPeriod,proto3" json:"cooldown_period,omitempty"`
	SetAt           *time.Time `protobuf:"bytes,5,opt,name=set_at,json=setAt,proto3,stdtime" json:"set_at,omitempty"`
	ExpiresAt       *time.Time `protobuf:"bytes,6,opt,name=expires_at,json=expiresAt,proto3,stdtime" json:"expires_at,omitempty"`
	Reason          string     `protobuf:"bytes,7,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (m *WalletLimit) Reset()         { *m = WalletLimit{} }
func (m *WalletLimit) String() string { return "WalletLimit" }
func (*WalletLimit) ProtoMessage()    {}

// ThresholdScheme is an alias for ThresholdSignatureScheme
type ThresholdScheme = securitypb.ThresholdSignatureScheme

// SecureEnclave represents a secure enclave configuration
type SecureEnclave struct {
	Id           string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	EnclaveType  string     `protobuf:"bytes,2,opt,name=enclave_type,json=enclaveType,proto3" json:"enclave_type,omitempty"`
	PublicKey    string     `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Attestation  string     `protobuf:"bytes,4,opt,name=attestation,proto3" json:"attestation,omitempty"`
	RegisteredAt *time.Time `protobuf:"bytes,5,opt,name=registered_at,json=registeredAt,proto3,stdtime" json:"registered_at,omitempty"`
	IsVerified   bool       `protobuf:"varint,6,opt,name=is_verified,json=isVerified,proto3" json:"is_verified,omitempty"`
}

func (m *SecureEnclave) Reset()         { *m = SecureEnclave{} }
func (m *SecureEnclave) String() string { return "SecureEnclave" }
func (*SecureEnclave) ProtoMessage()    {}

// RandomSource represents a verified randomness source
type RandomSource struct {
	Id           string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	SourceType   string     `protobuf:"bytes,2,opt,name=source_type,json=sourceType,proto3" json:"source_type,omitempty"`
	Endpoint     string     `protobuf:"bytes,3,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	PublicKey    string     `protobuf:"bytes,4,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	LastVerified *time.Time `protobuf:"bytes,5,opt,name=last_verified,json=lastVerified,proto3,stdtime" json:"last_verified,omitempty"`
	IsActive     bool       `protobuf:"varint,6,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
}

func (m *RandomSource) Reset()         { *m = RandomSource{} }
func (m *RandomSource) String() string { return "RandomSource" }
func (*RandomSource) ProtoMessage()    {}

// CertificatePin represents a pinned certificate
type CertificatePin struct {
	Id         string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Domain     string     `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	PinHash    string     `protobuf:"bytes,3,opt,name=pin_hash,json=pinHash,proto3" json:"pin_hash,omitempty"`
	PinType    string     `protobuf:"bytes,4,opt,name=pin_type,json=pinType,proto3" json:"pin_type,omitempty"`
	ValidFrom  *time.Time `protobuf:"bytes,5,opt,name=valid_from,json=validFrom,proto3,stdtime" json:"valid_from,omitempty"`
	ValidUntil *time.Time `protobuf:"bytes,6,opt,name=valid_until,json=validUntil,proto3,stdtime" json:"valid_until,omitempty"`
	IsActive   bool       `protobuf:"varint,7,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
}

func (m *CertificatePin) Reset()         { *m = CertificatePin{} }
func (m *CertificatePin) String() string { return "CertificatePin" }
func (*CertificatePin) ProtoMessage()    {}

// ViewKey is a simplified view key for privacy
type ViewKey struct {
	Id            string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	WalletAddress string     `protobuf:"bytes,2,opt,name=wallet_address,json=walletAddress,proto3" json:"wallet_address,omitempty"`
	ViewKey       string     `protobuf:"bytes,3,opt,name=view_key,json=viewKey,proto3" json:"view_key,omitempty"`
	RegisteredAt  *time.Time `protobuf:"bytes,4,opt,name=registered_at,json=registeredAt,proto3,stdtime" json:"registered_at,omitempty"`
	ValidUntil    *time.Time `protobuf:"bytes,5,opt,name=valid_until,json=validUntil,proto3,stdtime" json:"valid_until,omitempty"`
	Permissions   []string   `protobuf:"bytes,6,rep,name=permissions,proto3" json:"permissions,omitempty"`
}

func (m *ViewKey) Reset()         { *m = ViewKey{} }
func (m *ViewKey) String() string { return "ViewKey" }
func (*ViewKey) ProtoMessage()    {}

// RateLimit tracks operation rate limiting
// This is stored in memory and cleared on block boundaries
type RateLimit struct {
	Identifier string      `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Timestamps []time.Time `protobuf:"bytes,2,rep,name=timestamps,proto3,stdtime" json:"timestamps,omitempty"`
}

func (m *RateLimit) Reset()         { *m = RateLimit{} }
func (m *RateLimit) String() string { return "RateLimit" }
func (*RateLimit) ProtoMessage()    {}
