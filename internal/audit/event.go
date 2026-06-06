// Package audit предоставляет систему асинхронного аудита событий обновления метрик.
// Использует паттерн Publisher/Subscriber для рассылки аудит-событий подписчикам.
package audit

// AuditEvent представляет событие аудита, формируемое после успешной обработки метрик.
type AuditEvent struct {
	// Timestamp — Unix timestamp события.
	Timestamp int64 `json:"ts"`
	// Metrics — список имён обновлённых метрик.
	Metrics []string `json:"metrics"`
	// IPAddress — IP-адрес клиента, инициировавшего событие.
	IPAddress string `json:"ip_address"`
}
