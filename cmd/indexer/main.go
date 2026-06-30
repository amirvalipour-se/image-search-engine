package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"search-engine/internal/domain"
)

func main() {
	csvPath := "data/amazon_products_all.csv"
	idsPath := "index/image_ids.json"
	outputPath := "index/metadata.json"

	ids, err := readImageIDs(idsPath)
	if err != nil {
		panic(err)
	}

	records, err := readCSV(csvPath)
	if err != nil {
		panic(err)
	}

	products := buildMetadata(ids, records)
	if err := writeMetadata(outputPath, products); err != nil {
		panic(err)
	}

	fmt.Printf("Generated metadata.json with %d products\n", len(products))
}

func readImageIDs(path string) ([]int, error) {
	idsData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal(idsData, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func readCSV(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func buildMetadata(ids []int, records [][]string) map[int]domain.Product {
	products := make(map[int]domain.Product, len(ids))
	for faissID, csvRow := range ids {
		if csvRow+1 >= len(records) {
			continue
		}
		row := records[csvRow+1]
		if len(row) < 6 {
			continue
		}

		products[faissID] = domain.Product{
			ID:          faissID,
			ProductName: row[0],
			Category:    row[1],
			Description: row[2],
			Price:       parsePrice(row[3]),
			Image:       row[5],
		}
	}

	return products
}

func writeMetadata(path string, products map[int]domain.Product) error {
	out, err := json.MarshalIndent(products, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0644)
}

func parsePrice(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	s = strings.ReplaceAll(s, ",", "")

	var price float64
	fmt.Sscanf(s, "%f", &price)
	return price
}
