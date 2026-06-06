package audit

// AuditEvent представляет событие аудита, формируемое после успешной обработки метрик.
type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
