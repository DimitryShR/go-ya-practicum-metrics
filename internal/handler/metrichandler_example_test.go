package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

// ExampleMetricHandler_UpdateMetricHandler демонстрирует обновление метрики
// через URL-путь /update/{type}/{name}/{value}. В примере используется chi роутер
// с middleware ParseUpdatePathHandler, которая парсит URL и помещает метрику в контекст запроса.
func ExampleMetricHandler_UpdateMetricHandler() {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewMetricHandler(svc, nil)

	r := chi.NewRouter()
	// ParseUpdatePathHandler применяется только к маршруту /update/{type}/{name}/{value}
	r.With(middleware.ParseUpdatePathHandler).Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/update/counter/test_counter/10", "text/plain", nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	// Output:
	// Status: 200
}

// ExampleMetricHandler_GetMetricValue демонстрирует получение значения метрики
// через URL-путь /value/{type}/{name}. Предварительно метрика создаётся через
// POST /update/{type}/{name}/{value}.
func ExampleMetricHandler_GetMetricValue() {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewMetricHandler(svc, nil)

	r := chi.NewRouter()
	r.With(middleware.ParseUpdatePathHandler).Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricValue)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Сначала обновляем метрику через полный HTTP-запрос
	updateResp, err := http.Post(ts.URL+"/update/counter/test_counter/10", "text/plain", nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	updateResp.Body.Close()

	// Получаем значение метрики
	resp, err := http.Get(ts.URL + "/value/counter/test_counter")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", string(body))

	// Output:
	// Status: 200
	// Body: 10
}

// ExampleMetricHandler_GetAllMetrics демонстрирует получение HTML-страницы
// со всеми сохранёнными метриками через GET /. Предварительно создаются
// метрики типов gauge и counter.
func ExampleMetricHandler_GetAllMetrics() {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewMetricHandler(svc, nil)

	r := chi.NewRouter()
	r.With(middleware.ParseUpdatePathHandler).Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/", h.GetAllMetrics)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Обновляем несколько метрик
	resp, err := http.Post(ts.URL+"/update/gauge/test_gauge/42.5", "text/plain", nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	resp.Body.Close()

	resp, err = http.Post(ts.URL+"/update/counter/test_counter/10", "text/plain", nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	resp.Body.Close()

	// Получаем HTML-страницу со всеми метриками
	resp, err = http.Get(ts.URL + "/")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
	fmt.Println("Contains gauge:", strings.Contains(string(body), "test_gauge"))
	fmt.Println("Contains counter:", strings.Contains(string(body), "test_counter"))

	// Output:
	// Status: 200
	// Content-Type: text/html
	// Contains gauge: true
	// Contains counter: true
}
