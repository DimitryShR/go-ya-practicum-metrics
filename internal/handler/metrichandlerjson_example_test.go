package handler_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

// ExampleMetricHandler_UpdateMetricHandlerJSON демонстрирует обновление метрики
// через JSON API: POST /update с телом запроса в формате JSON.
// Хендлер возвращает обновлённую метрику в JSON.
func ExampleMetricHandler_UpdateMetricHandlerJSON() {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewMetricHandler(svc, nil)

	r := chi.NewRouter()
	r.Post("/update", h.UpdateMetricHandlerJSON)

	ts := httptest.NewServer(r)
	defer ts.Close()

	body := `{"id":"test_counter","type":"counter","delta":10}`
	resp, err := http.Post(ts.URL+"/update", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", strings.TrimSpace(string(respBody)))

	// Output:
	// Status: 200
	// Body: {"id":"test_counter","type":"counter","delta":10}
}

// ExampleMetricHandler_GetMetricValueJSON демонстрирует получение значения метрики
// через JSON API: POST /value с телом запроса {id, type}. Хендлер возвращает
// метрику с заполненным полем delta или value.
func ExampleMetricHandler_GetMetricValueJSON() {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewMetricHandler(svc, nil)

	r := chi.NewRouter()
	r.Post("/update", h.UpdateMetricHandlerJSON)
	r.Post("/value", h.GetMetricValueJSON)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Сначала обновляем метрику
	updateBody := `{"id":"test_gauge","type":"gauge","value":42.5}`
	resp, err := http.Post(ts.URL+"/update", "application/json", bytes.NewReader([]byte(updateBody)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	resp.Body.Close()

	// Получаем значение метрики
	getBody := `{"id":"test_gauge","type":"gauge"}`
	resp, err = http.Post(ts.URL+"/value", "application/json", bytes.NewReader([]byte(getBody)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", strings.TrimSpace(string(respBody)))

	// Output:
	// Status: 200
	// Body: {"id":"test_gauge","type":"gauge","value":42.5}
}

// ExampleMetricHandler_UpdateMetricsHandlerJSON демонстрирует пакетное обновление метрик
// через JSON API: POST /updates с массивом метрик в теле запроса.
func ExampleMetricHandler_UpdateMetricsHandlerJSON() {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewMetricHandler(svc, nil)

	r := chi.NewRouter()
	r.Post("/updates", h.UpdateMetricsHandlerJSON)

	ts := httptest.NewServer(r)
	defer ts.Close()

	body := `[{"id":"test_counter","type":"counter","delta":10},{"id":"test_gauge","type":"gauge","value":3.14}]`
	resp, err := http.Post(ts.URL+"/updates", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	// Output:
	// Status: 200
}
