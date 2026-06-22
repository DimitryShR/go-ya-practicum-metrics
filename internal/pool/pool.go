// Package pool предоставляет generic-пул для повторного использования объектов
// с методом Reset(). Использует sync.Pool для thread-safe хранения объектов.
package pool

import "sync"

// Resetter — интерфейс для типов, поддерживающих сброс состояния.
type Resetter interface {
	Reset()
}

// Pool — generic-пул для повторного использования объектов с методом Reset().
// Тип T должен реализовывать интерфейс Resetter.
type Pool[T Resetter] struct {
	pool *sync.Pool
}

// New создаёт и возвращает новый Pool для объектов типа T.
// Функция newFunc используется для создания новых объектов, когда пул пуст.
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get возвращает объект типа T из пула.
// Если пул пуст, создаётся новый объект через фабрику, переданную в New.
func (p *Pool[T]) Get() T {
	// Type assertion безопасен, так как Pool контролирует все Put().
	return p.pool.Get().(T)
}

// Put помещает объект типа T в пул для повторного использования.
// Перед помещением вызывается Reset() для очистки состояния объекта.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
