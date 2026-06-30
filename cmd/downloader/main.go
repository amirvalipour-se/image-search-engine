package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	csvPath    = "data/amazon_products_all.csv"
	outputDir  = "images"
	numWorkers = 50
)

type Job struct {
	ID  int
	URL string
}

func main() {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	header, err := reader.Read()
	if err != nil {
		panic(err)
	}

	imageIndex := -1

	for i, col := range header {
		if col == "Image" {
			imageIndex = i
			break
		}
	}

	if imageIndex == -1 {
		panic("Image column not found")
	}

	jobs := make(chan Job, 100)

	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go worker(client, jobs, &wg)
	}

	id := 0

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			continue
		}

		jobs <- Job{
			ID:  id,
			URL: record[imageIndex],
		}

		id++
	}

	close(jobs)

	wg.Wait()

	fmt.Println("Download complete")
}

func worker(client *http.Client, jobs <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {

		filename := filepath.Join(outputDir, fmt.Sprintf("%d.jpg", job.ID))

		// Skip already downloaded files
		if _, err := os.Stat(filename); err == nil {
			continue
		}

		resp, err := client.Get(job.URL)
		if err != nil {
			fmt.Printf("Download failed %d\n", job.ID)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			fmt.Printf("HTTP %d for %d\n", resp.StatusCode, job.ID)
			continue
		}

		file, err := os.Create(filename)
		if err != nil {
			resp.Body.Close()
			continue
		}

		_, err = io.Copy(file, resp.Body)

		file.Close()
		resp.Body.Close()

		if err != nil {
			os.Remove(filename)
			continue
		}

		fmt.Printf("Downloaded %d\n", job.ID)
	}
}
