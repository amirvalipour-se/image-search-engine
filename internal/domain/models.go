package domain

type Product struct {
	ID          int     `json:"id"`
	ProductName string  `json:"productName"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
}

type SearchResult struct {
	ID          int     `json:"id"`
	ProductName string  `json:"productName"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Score       float64 `json:"score"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type PythonMatch struct {
	ID    int     `json:"id"`
	Score float64 `json:"score"`
}

type PythonSearchResponse struct {
	Matches []PythonMatch `json:"matches"`
	Count   int           `json:"count"`
	TookS   float64       `json:"took_s"`
}
