package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

const (
	defaultFlushInterval = 1 * time.Second
	defaultBufferSize    = 4096
)

// FileAuditor — подписчик (наблюдатель), записывающий аудит-события в файл.
// Файл открывается один раз при создании, запись идёт через bufio.Writer.
// Периодический Flush гарантирует, что данные попадают на диск.
type FileAuditor struct {
	filePath string
	file     *os.File
	writer   *bufio.Writer
	mu       sync.Mutex
	ctx      context.Context
	wg       sync.WaitGroup
}

// NewFileAuditor создаёт FileAuditor с указанным путём к файлу.
// Файл открывается в режиме append. При необходимости создаётся.
// При отмене ctx flushLoop автоматически завершается.
func NewFileAuditor(ctx context.Context, filePath string) (*FileAuditor, error) {
	fa := &FileAuditor{
		filePath: filePath,
		ctx:      ctx,
	}
	if err := fa.openFile(); err != nil {
		return nil, err
	}
	fa.wg.Add(1)
	go fa.flushLoop()
	return fa, nil
}

// openFile открывает файл для записи в режиме append.
func (f *FileAuditor) openFile() error {
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Log.Error("failed to open audit file",
			zap.String("path", f.filePath), zap.Error(err))
		return err
	}
	f.file = file
	f.writer = bufio.NewWriterSize(f.file, defaultBufferSize)
	return nil
}

// flushLoop периодически сбрасывает буфер на диск.
// При отмене ctx автоматически выполняет финальный flush и закрывает файл.
func (f *FileAuditor) flushLoop() {
	defer f.wg.Done()
	defer f.close()
	ticker := time.NewTicker(defaultFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-f.ctx.Done():
			return
		case <-ticker.C:
			_ = f.Flush()
		}
	}
}

// Handle записывает событие аудита в буфер.
func (f *FileAuditor) Handle(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	data = append(data, '\n')
	_, err = f.writer.Write(data)
	return err
}

// Flush сбрасывает буфер writer на диск.
func (f *FileAuditor) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.writer == nil {
		return nil
	}
	if err := f.writer.Flush(); err != nil {
		return err
	}
	return f.file.Sync()
}

// Wait блокируется до полного завершения flushLoop (включая финальный flush и закрытие файла).
func (f *FileAuditor) Wait() {
	f.wg.Wait()
}

// close выполняет финальный flush буфера и закрывает файл.
// Вызывается автоматически из flushLoop при отмене ctx.
func (f *FileAuditor) close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.writer != nil {
		_ = f.writer.Flush()
	}
	if f.file != nil {
		_ = f.file.Close()
	}
}
