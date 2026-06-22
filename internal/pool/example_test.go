package pool_test

import (
	"fmt"
	"sync"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/pool"
)

// buffer — тестовая структура, представляющая буфер с методом Reset()
type buffer struct {
	data []byte
	pos  int
}

func (b *buffer) Reset() {
	b.data = b.data[:0]
	b.pos = 0
}

func (b *buffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *buffer) String() string {
	return string(b.data)
}

// newBuffer — фабрика для создания буферов
func newBuffer() *buffer {
	return &buffer{
		data: make([]byte, 0, 1024),
	}
}

// Example_usage демонстрирует базовое использование Pool
func Example_usage() {
	// Создаём пул с фабрикой
	p := pool.New(newBuffer)

	// Получаем буфер из пула
	buf := p.Get()
	buf.Write([]byte("Hello, "))
	buf.Write([]byte("Pool!"))
	fmt.Println(buf)

	// Возвращаем буфер в пул
	p.Put(buf)

	// Получаем буфер снова — он будет сброшен
	buf2 := p.Get()
	fmt.Println(buf2)

	// Output:
	// Hello, Pool!
	//
}

// Example_concurrent демонстрирует многопоточное использование Pool.
// Несколько горутин одновременно получают, используют и возвращают объекты.
func Example_concurrent() {
	p := pool.New(newBuffer)

	var wg sync.WaitGroup

	// Запускаем 3 горутины, каждая получает буфер, пишет данные и возвращает
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := p.Get()
			buf.Write([]byte("data"))
			p.Put(buf)
		}()
	}

	// Ждём завершения всех горутин
	wg.Wait()

	// После параллельной работы получаем буфер — он должен быть пустым
	buf := p.Get()
	fmt.Println(buf)

	// Output:
	//
}
