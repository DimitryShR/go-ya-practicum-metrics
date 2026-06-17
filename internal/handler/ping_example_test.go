package handler_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	_ "modernc.org/sqlite"
)

// ExamplePingHandler_Ping демонстрирует проверку соединения с базой данных
// через GET /ping. В примере используется in-memory SQLite.
// При успешном подключении хендлер возвращает 200 OK.
func ExamplePingHandler_Ping() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()

	// Проверяем, что соединение работает
	if err := db.Ping(); err != nil {
		fmt.Println("error:", err)
		return
	}

	h := handler.NewPingHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.Ping(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 200
}
