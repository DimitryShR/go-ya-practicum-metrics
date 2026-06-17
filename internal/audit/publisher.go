package audit

import (
	"sync"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

// Auditor — интерфейс подписчика (наблюдатель), получающего уведомления о событиях аудита.
type Auditor interface {
	Handle(event AuditEvent) error
}

// Publisher — интерфейс издателя, управляющего подписками и рассылающего уведомления.
type Publisher interface {
	Register(auditor Auditor)
	Deregister(auditor Auditor)
	Notify(event AuditEvent)
}

// AuditPublisher — реализация Publisher, хранящая список подписчиков и оповещающая их о событиях.
// Отправка событий происходит асинхронно через buffered channel.
type AuditPublisher struct {
	mu           sync.Mutex
	observers    []Auditor
	ch           chan AuditEvent
	workers      int
	wg           sync.WaitGroup
	shutdownOnce sync.Once
}

const publisherBufferSize = 256
const defaultWorkers = 2

// NewAuditPublisher создаёт новый экземпляр AuditPublisher и запускает воркеры для обработки событий.
func NewAuditPublisher() *AuditPublisher {
	p := &AuditPublisher{
		ch:      make(chan AuditEvent, publisherBufferSize),
		workers: defaultWorkers,
	}
	p.startWorkers()
	return p
}

// startWorkers запускает горутины-воркеры для асинхронной обработки аудит-событий.
func (p *AuditPublisher) startWorkers() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker читает события из канала и передаёт их подписчикам.
func (p *AuditPublisher) worker() {
	defer p.wg.Done()
	for event := range p.ch {
		p.dispatchEvent(event)
	}
}

// dispatchEvent оповещает всех подписчиков о событии аудита.
// Копирует список observers, чтобы не удерживать RLock во время I/O-операций.
func (p *AuditPublisher) dispatchEvent(event AuditEvent) {
	p.mu.Lock()
	observers := make([]Auditor, len(p.observers))
	copy(observers, p.observers)
	p.mu.Unlock()

	for _, obs := range observers {
		if err := obs.Handle(event); err != nil {
			logger.Log.Error("audit observer handling failed", zap.Error(err))
		}
	}
}

// Register добавляет подписчика в список наблюдателей.
func (p *AuditPublisher) Register(auditor Auditor) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, auditor)
}

// Deregister удаляет подписчика из списка наблюдателей.
func (p *AuditPublisher) Deregister(auditor Auditor) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, obs := range p.observers {
		if obs == auditor {
			// Копируем все элементы после удаляемого и смещаем влево на 1.
			copy(p.observers[i:], p.observers[i+1:])
			// Зануляем последний (освободившийся) элемент для избежания утечки.
			p.observers[len(p.observers)-1] = nil
			// Усекаем длину слайса после обнуления последнего элемента.
			p.observers = p.observers[:len(p.observers)-1]
			return
		}
	}
}

// Notify отправляет событие аудита в buffered channel для асинхронной обработки.
// Если канал заполнен, событие не отправляется (non-blocking), чтобы не блокировать HTTP-обработчик.
func (p *AuditPublisher) Notify(event AuditEvent) {
	select {
	case p.ch <- event:
	default:
		logger.Log.Warn("audit event channel is full, event dropped")
	}
}

// Shutdown останавливает воркеры и ожидает обработки оставшихся событий.
func (p *AuditPublisher) Shutdown() {
	p.shutdownOnce.Do(func() {
		close(p.ch)
		p.wg.Wait()
	})
}
