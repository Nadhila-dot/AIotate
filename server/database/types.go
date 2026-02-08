package store

import "time"

type Store struct {
	Name string
	Path string
}

type DB struct {
	Path string
}

type Optional struct {
	Tags        []string `json:"tags,omitempty"`
	Color       string   `json:"color,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Notebook struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Username    string            `json:"username"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Optional    Optional          `json:"optional,omitempty"`
	Items       map[string]string `json:"items"`
}

type Style struct {
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	Prompt      string    `json:"prompt"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"isDefault"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
