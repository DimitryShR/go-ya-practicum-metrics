package pool

import (
	"sync"
	"sync/atomic"
	"testing"
)

// testStruct — тестовая структура с методом Reset()
type testStruct struct {
	data       []int
	name       string
	mu         sync.Mutex
	resetCount atomic.Int64
}

func (t *testStruct) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resetCount.Add(1)
	t.data = t.data[:0]
	t.name = ""
}

// newTestStruct — фабрика для создания тестовых структур
func newTestStruct() *testStruct {
	return &testStruct{
		data: make([]int, 0, 10),
		name: "initial",
	}
}

// TestNew проверяет создание пула
func TestNew(t *testing.T) {
	p := New(newTestStruct)
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.pool == nil {
		t.Fatal("pool field is nil")
	}
}

// TestGetEmptyPool проверяет Get() когда пул пуст.
// sync.Pool создаёт новый объект через фабрику, поэтому объект приходит с начальным состоянием, заданным фабрикой.
func TestGetEmptyPool(t *testing.T) {
	p := New(newTestStruct)
	obj := p.Get()

	if obj == nil {
		t.Fatal("Get() returned nil")
	}
}

// TestGetReuse проверяет повторное использование объекта из пула
func TestGetReuse(t *testing.T) {
	p := New(newTestStruct)

	// Получаем объект, модифицируем и возвращаем в пул
	obj1 := p.Get()
	obj1.name = "modified"
	obj1.data = append(obj1.data, 1, 2, 3)
	p.Put(obj1)

	// Получаем объект снова — должен быть сброшен
	obj2 := p.Get()
	if obj2.name != "" {
		t.Errorf("expected empty name after Reset(), got %q", obj2.name)
	}
	if len(obj2.data) != 0 {
		t.Errorf("expected empty data after Reset(), got %v", obj2.data)
	}
}

// TestPutReset проверяет, что Put() вызывает Reset()
func TestPutReset(t *testing.T) {
	p := New(newTestStruct)

	obj := p.Get()
	obj.name = "should_be_reset"
	obj.data = append(obj.data, 1, 2, 3)

	resetBefore := obj.resetCount.Load()
	p.Put(obj)
	resetAfter := obj.resetCount.Load()

	if resetAfter-resetBefore != 1 {
		t.Errorf("expected Reset() to be called once, called %d times", resetAfter-resetBefore)
	}
}

// TestConcurrent проверяет многопоточное использование пула
func TestConcurrent(t *testing.T) {
	p := New(newTestStruct)
	var wg sync.WaitGroup
	iterations := 100
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				obj := p.Get()
				obj.name = "goroutine"
				obj.data = append(obj.data, int(id))
				p.Put(obj)
			}
		}(i)
	}

	wg.Wait()
}

// TestMultipleGets проверяет получение нескольких объектов
func TestMultipleGets(t *testing.T) {
	p := New(newTestStruct)

	objs := make([]*testStruct, 5)
	for i := 0; i < 5; i++ {
		objs[i] = p.Get()
		if objs[i] == nil {
			t.Fatal("Get() returned nil")
		}
	}

	// Возвращаем все объекты в пул
	for _, obj := range objs {
		p.Put(obj)
	}

	// Получаем объекты снова — они должны быть сброшены
	for i := 0; i < 5; i++ {
		obj := p.Get()
		if obj.name != "" {
			t.Errorf("expected empty name after Reset(), got %q", obj.name)
		}
	}
}

// TestConcurrentResetCount проверяет, что Reset() вызывается при каждом Get/Put
func TestConcurrentResetCount(t *testing.T) {
	p := New(newTestStruct)

	var wg sync.WaitGroup
	iterations := 50
	goroutines := 5

	// Запускаем горутины, которые получают и возвращают объект
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				o := p.Get()
				o.name = "used"
				o.data = append(o.data, j)
				p.Put(o)
			}
		}()
	}

	wg.Wait()

	// Получаем объект — он должен быть сброшен
	obj := p.Get()
	if obj.name != "" {
		t.Errorf("expected empty name after concurrent use, got %q", obj.name)
	}
	if len(obj.data) != 0 {
		t.Errorf("expected empty data after concurrent use, got %v", obj.data)
	}
}

// TestNilSafety проверяет безопасность при nil объектах
func TestNilSafety(t *testing.T) {
	// Тест с указательным типом
	p := New(func() *testStruct {
		return newTestStruct()
	})

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get() returned nil for pointer type")
	}

	// Возвращаем объект — не должно быть паники
	p.Put(obj)

	// Получаем снова
	obj2 := p.Get()
	if obj2 == nil {
		t.Fatal("Get() returned nil after Put()")
	}
}
