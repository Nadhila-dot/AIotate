package ai

// AIProvider represents the AI service provider
type AIProvider string

const (
	ProviderGemini     AIProvider = "gemini"
	ProviderOpenRouter AIProvider = "openrouter"
)

// AIConfig holds the AI configuration
type AIConfig struct {
	Provider         AIProvider
	GeminiAPIKey     string
	OpenRouterAPIKey string
	MainModel        string // For LaTeX generation
	UtilityModel     string // For descriptions, tags, etc.
}

// TaskType represents different AI task types
type TaskType string

const (
	TaskLaTeXGeneration TaskType = "latex_generation"
	TaskUtility         TaskType = "utility"
)

// ModelConfig holds model configuration for different tasks
type ModelConfig struct {
	Provider AIProvider
	Model    string
	APIKey   string
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// Attachment represents an uploaded file or extracted content passed to the AI
type Attachment struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"` // "utf-8" or "base64"
}
