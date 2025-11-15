package prompts

// Template represents a prompt template with metadata and content
type Template struct {
	Name         string   `yaml:"-"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Tags         []string `yaml:"tags"`
	Category     string   `yaml:"category"`
	Dependencies []string `yaml:"dependencies"`
	Related      []string `yaml:"related"`
	Content      string   `yaml:"-"`
}

// DiscoverResult represents the result of pattern discovery
type DiscoverResult struct {
	Query      string     `json:"query"`
	Matches    []Template `json:"matches"`
	Default    *Template  `json:"default,omitempty"`
	Message    string     `json:"message"`
}

// Pattern represents a template in discovery results (lightweight)
type Pattern struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Category    string   `json:"category"`
	Score       float64  `json:"score,omitempty"`
}
