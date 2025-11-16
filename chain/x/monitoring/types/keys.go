package types

const (
	// ModuleName defines the module name
	ModuleName = "monitoring"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_monitoring"
)

// Store key prefixes
var (
	// TransactionPrefix is the prefix for transaction monitoring data
	TransactionPrefix = []byte{0x01}

	// AlertPrefix is the prefix for alert data
	AlertPrefix = []byte{0x02}

	// AnomalyPrefix is the prefix for anomaly detection data
	AnomalyPrefix = []byte{0x03}

	// ValidatorUptimePrefix is the prefix for validator uptime data
	ValidatorUptimePrefix = []byte{0x04}

	// NetworkHealthPrefix is the prefix for network health metrics
	NetworkHealthPrefix = []byte{0x05}

	// GasPricePrefix is the prefix for gas price tracking
	GasPricePrefix = []byte{0x06}

	// TVLPrefix is the prefix for Total Value Locked data
	TVLPrefix = []byte{0x07}

	// LargeTransactionPrefix is the prefix for large transaction alerts
	LargeTransactionPrefix = []byte{0x08}

	// FailedTransactionPrefix is the prefix for failed transaction patterns
	FailedTransactionPrefix = []byte{0x09}

	// SecurityEventPrefix is the prefix for SIEM security events
	SecurityEventPrefix = []byte{0x0a}

	// LogAggregationPrefix is the prefix for centralized log data
	LogAggregationPrefix = []byte{0x0b}
)
