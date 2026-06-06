package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// ExamplePingHandler_Ping демонстрирует проверку соединения с базой данных.
// В реальном коде используется sql.Open для подключения к БД.
func ExamplePingHandler_Ping() {
	// Пример с моком: в реальности здесь будет sql.Open("postgres", dsn)
	// Для примера создадим nil db, чтобы показать ошибку
	var db *sql.DB = nil

	h := NewPingHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.Ping(w, req)

	fmt.Println("Status:", w.Code)
	// Output:
	// Status: 500
}

// ExamplePingHandler_Ping_second демонстрирует ответ при отсутствии подключения к БД.
func ExamplePingHandler_Ping_second() {
	// Этот пример показывает поведение при nil db
	h := NewPingHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.Ping(w, req)

	fmt.Println("Status:", w.Code)
	// Output:
	// Status: 500
}
