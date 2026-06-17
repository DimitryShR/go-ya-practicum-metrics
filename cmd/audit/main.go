package main

import (
	"flag"
	"io"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	addr := flag.String("a", ":8083", "Audit Server address")
	flag.Parse()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("failed to read audit request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Log.Info("audit event received", zap.String("body", string(body)))
		w.WriteHeader(http.StatusOK)
	})

	logger.Log.Info("audit server started", zap.String("address", *addr))
	if err := http.ListenAndServe(*addr, nil); err != nil {
		logger.Log.Fatal("audit server error", zap.Error(err))
	}
}
