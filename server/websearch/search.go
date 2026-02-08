package websearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	defaultLimit       = 3
	maxExtractChars    = 20000
	maxUserAgent       = "Mozilla/5.0 (compatible; NightwaysBot/1.0)"
	duckDuckGoEndpoint = "https://api.duckduckgo.com/"
	serpAPIEndpoint    = "https://serpapi.com/search.json"
)

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"`
}

// Search performs a web search. If SERPAPI_KEY is configured, it uses Google via SerpAPI.
// Otherwise it falls back to DuckDuckGo Instant Answer API.
func Search(query string, limit int) ([]SearchResult, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > 5 {
		limit = 5
	}

	if apiKey := strings.TrimSpace(os.Getenv("SERPAPI_KEY")); apiKey != "" {
		return searchSerpAPI(q, apiKey, limit)
	}

	return searchDuckDuckGo(q, limit)
}

// SearchAndExtract performs search and fetches text content from top results.
func SearchAndExtract(query string, limit int) (string, []SearchResult, error) {
	results, err := Search(query, limit)
	if err != nil {
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString("Web Research Results:\n")

	for i, res := range results {
		text, err := ExtractTextFromURL(res.URL)
		if err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("\n[%d] %s\nURL: %s\n", i+1, res.Title, res.URL))
		b.WriteString(text)
		b.WriteString("\n---\n")
	}

	return b.String(), results, nil
}

func searchDuckDuckGo(query string, limit int) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_redirect", "1")
	params.Set("no_html", "1")

	reqURL := duckDuckGoEndpoint + "?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("duckduckgo error: %s", string(body))
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := []SearchResult{}

	if topics, ok := data["RelatedTopics"].([]interface{}); ok {
		for _, t := range topics {
			if len(results) >= limit {
				break
			}
			item, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			if nested, ok := item["Topics"].([]interface{}); ok {
				for _, n := range nested {
					if len(results) >= limit {
						break
					}
					nm, ok := n.(map[string]interface{})
					if !ok {
						continue
					}
					appendDuckResult(&results, nm)
				}
				continue
			}
			appendDuckResult(&results, item)
		}
	}

	if len(results) == 0 {
		return nil, errors.New("no results found")
	}

	return results, nil
}

func appendDuckResult(results *[]SearchResult, item map[string]interface{}) {
	title, _ := item["Text"].(string)
	urlStr, _ := item["FirstURL"].(string)
	if title == "" || urlStr == "" {
		return
	}
	*results = append(*results, SearchResult{
		Title:   title,
		URL:     urlStr,
		Snippet: title,
		Source:  "duckduckgo",
	})
}

type serpAPIResponse struct {
	OrganicResults []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic_results"`
}

func searchSerpAPI(query, apiKey string, limit int) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("engine", "google")
	params.Set("q", query)
	params.Set("api_key", apiKey)

	reqURL := serpAPIEndpoint + "?" + params.Encode()

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serpapi error: %s", string(body))
	}

	var data serpAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := []SearchResult{}
	for _, r := range data.OrganicResults {
		if len(results) >= limit {
			break
		}
		if r.Link == "" || r.Title == "" {
			continue
		}
		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.Link,
			Snippet: r.Snippet,
			Source:  "serpapi",
		})
	}

	if len(results) == 0 {
		return nil, errors.New("no results found")
	}

	return results, nil
}

// ExtractTextFromURL fetches a URL and extracts readable text from HTML.
func ExtractTextFromURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", errors.New("url is required")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", maxUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch url: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	text := extractText(doc)
	text = strings.TrimSpace(text)
	if len(text) > maxExtractChars {
		text = text[:maxExtractChars] + "\n[TRUNCATED]"
	}

	if text == "" {
		return "", errors.New("no text extracted")
	}

	return text, nil
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data + " "
	}

	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "noscript") {
		return ""
	}

	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(extractText(c))
	}
	return b.String()
}
