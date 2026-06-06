package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTimeout = 5 * time.Second
	testTick    = 10 * time.Millisecond
)

// newTestPublisher создаёт AuditPublisher без запуска воркеров (только для unit-тестов publisher).
func newTestPublisher() *AuditPublisher {
	p := &AuditPublisher{
		ch:      make(chan AuditEvent, publisherBufferSize),
		workers: defaultWorkers,
	}
	p.startWorkers()
	return p
}

// newTestRemoteAuditor создаёт RemoteAuditor с уменьшенными retry параметрами для быстрых тестов.
func newTestRemoteAuditor(ctx context.Context, url string) *RemoteAuditor {
	return NewRemoteAuditorWithRetry(ctx, url, 5*time.Second, 2, 10*time.Millisecond, 50*time.Millisecond)
}

// mockAuditor — простой мок для тестирования Publisher
type mockAuditor struct {
	handleFunc func(event AuditEvent) error
}

func (m *mockAuditor) Handle(event AuditEvent) error {
	return m.handleFunc(event)
}

func TestAuditPublisher_RegisterAndNotify(t *testing.T) {
	publisher := newTestPublisher()

	var received []AuditEvent
	var mu sync.Mutex

	mockAuditor := &mockAuditor{
		handleFunc: func(event AuditEvent) error {
			mu.Lock()
			defer mu.Unlock()
			received = append(received, event)
			return nil
		},
	}

	publisher.Register(mockAuditor)

	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.1",
	}

	publisher.Notify(event)

	// Даём время воркеру обработать событие (асинхронная обработка)
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(received) == 1
	}, testTimeout, testTick, "expected 1 event to be received")

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, event, received[0])
}

func TestAuditPublisher_MultipleSubscribers(t *testing.T) {
	publisher := newTestPublisher()

	callCount := 0
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		publisher.Register(&mockAuditor{
			handleFunc: func(event AuditEvent) error {
				mu.Lock()
				defer mu.Unlock()
				callCount++
				return nil
			},
		})
	}

	event := AuditEvent{Timestamp: 1, Metrics: []string{"test"}, IPAddress: "127.0.0.1"}
	publisher.Notify(event)

	// Даём время воркеру обработать событие
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return callCount == 3
	}, testTimeout, testTick, "expected 3 subscribers to be notified")
}

func TestAuditPublisher_Deregister(t *testing.T) {
	publisher := newTestPublisher()

	called := false
	aud := &mockAuditor{
		handleFunc: func(event AuditEvent) error {
			called = true
			return nil
		},
	}

	publisher.Register(aud)
	publisher.Deregister(aud)

	event := AuditEvent{Timestamp: 1, Metrics: []string{"test"}, IPAddress: "127.0.0.1"}
	publisher.Notify(event)

	// Даём время воркеру обработать событие
	assert.Eventually(t, func() bool {
		// called остаётся false если дегистрированный аудитор не был вызван
		return !called
	}, testTimeout, testTick, "deregistered auditor should not be called")
}

func TestAuditPublisher_NotifyErrorDoesNotStopOthers(t *testing.T) {
	publisher := newTestPublisher()

	firstCalled := false
	thirdCalled := false
	var mu sync.Mutex

	publisher.Register(&mockAuditor{
		handleFunc: func(event AuditEvent) error {
			mu.Lock()
			defer mu.Unlock()
			firstCalled = true
			return nil
		},
	})
	publisher.Register(&mockAuditor{
		handleFunc: func(event AuditEvent) error {
			return assert.AnError
		},
	})
	publisher.Register(&mockAuditor{
		handleFunc: func(event AuditEvent) error {
			mu.Lock()
			defer mu.Unlock()
			thirdCalled = true
			return nil
		},
	})

	event := AuditEvent{Timestamp: 1, Metrics: []string{"test"}, IPAddress: "127.0.0.1"}
	publisher.Notify(event)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return firstCalled && thirdCalled
	}, testTimeout, testTick, "first and third subscribers should be called even if second fails")
}

func TestFileAuditor_Handle(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit.log")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fa, err := NewFileAuditor(ctx, filePath)
	require.NoError(t, err)

	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	err = fa.Handle(event)
	require.NoError(t, err)

	// Flush чтобы данные гарантированно попали в файл
	err = fa.Flush()
	require.NoError(t, err)

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Проверяем, что файл содержит одну строку с JSON
	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	require.Len(t, lines, 1)

	var parsed AuditEvent
	err = json.Unmarshal(lines[0], &parsed)
	require.NoError(t, err)
	assert.Equal(t, event, parsed)

	// Останавливаем flushLoop и закрываем файл
	cancel()
	fa.Wait()
}

func TestFileAuditor_MultipleEvents(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit.json")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fa, err := NewFileAuditor(ctx, filePath)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		event := AuditEvent{
			Timestamp: int64(i),
			Metrics:   []string{string(rune('A' + i))},
			IPAddress: "127.0.0.1",
		}

		err = fa.Handle(event)
		require.NoError(t, err)
	}

	// Flush чтобы все данные попали в файл
	err = fa.Flush()
	require.NoError(t, err)

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	require.Len(t, lines, 3)

	for i, line := range lines {
		var parsed AuditEvent
		err = json.Unmarshal(line, &parsed)
		require.NoError(t, err)
		assert.Equal(t, int64(i), parsed.Timestamp)
	}

	// Останавливаем flushLoop и закрываем файл
	cancel()
	fa.Wait()
}

func TestRemoteAuditor_Handle(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ra := NewRemoteAuditor(context.Background(), server.URL, 5*time.Second)

	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   []string{"test"},
		IPAddress: "10.0.0.1",
	}

	err := ra.Handle(event)
	require.NoError(t, err)

	var parsed AuditEvent
	err = json.Unmarshal(receivedBody, &parsed)
	require.NoError(t, err)
	assert.Equal(t, event, parsed)
}

func TestRemoteAuditor_HandleServerError(t *testing.T) {
	// Сервер всегда возвращает 500, resty retry'ит и возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Используем маленький таймаут и retry count для быстрого теста
	ra := newTestRemoteAuditor(context.Background(), server.URL)

	event := AuditEvent{
		Timestamp: 1,
		Metrics:   []string{"test"},
		IPAddress: "127.0.0.1",
	}

	err := ra.Handle(event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestRemoteAuditor_RetryOn5xx(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count <= 2 {
			// Первые 2 запроса — ошибка
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Третий запрос — успех
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ra := newTestRemoteAuditor(context.Background(), server.URL)

	event := AuditEvent{
		Timestamp: 12345678,
		Metrics:   []string{"test"},
		IPAddress: "127.0.0.1",
	}

	err := ra.Handle(event)
	require.NoError(t, err)
	assert.Equal(t, int32(3), requestCount.Load(), "expected 3 requests (2 retries + 1 success)")
}

func TestRemoteAuditor_NoRetryOn4xx(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	ra := newTestRemoteAuditor(context.Background(), server.URL)

	event := AuditEvent{
		Timestamp: 1,
		Metrics:   []string{"test"},
		IPAddress: "127.0.0.1",
	}

	err := ra.Handle(event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 400")
	assert.Equal(t, int32(1), requestCount.Load(), "expected no retry for 4xx errors")
}

func TestRemoteAuditor_NoRetryOnContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // сразу отменяем

	ra := newTestRemoteAuditor(ctx, "http://localhost:1")

	event := AuditEvent{
		Timestamp: 1,
		Metrics:   []string{"test"},
		IPAddress: "127.0.0.1",
	}

	err := ra.Handle(event)
	require.Error(t, err)
	// Ошибка должна быть связана с отменой контекста, а не с retry
	assert.ErrorIs(t, err, context.Canceled)
}
