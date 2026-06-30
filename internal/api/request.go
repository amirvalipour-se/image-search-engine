package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func readImageFile(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func parseK(r *http.Request, fallback int) int {
	k := fallback
	if value := r.FormValue("k"); value != "" {
		fmt.Sscanf(value, "%d", &k)
	} else if value := r.URL.Query().Get("k"); value != "" {
		fmt.Sscanf(value, "%d", &k)
	}

	if k <= 0 {
		return fallback
	}
	return k
}

func parsePrice(value string) float64 {
	price, _ := strconv.ParseFloat(value, 64)
	return price
}
