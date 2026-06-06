package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("error reading body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Printf("AUDIT EVENT: %s\n", string(body))
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Audit server listening on :8083")
	if err := http.ListenAndServe(":8083", nil); err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
