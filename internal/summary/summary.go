package summary

import (
	"io"
	"os"
	
	"github.com/mattn/go-isatty"
)

// Options configures summary output behavior
type Options struct {
	Format  string // "auto", "ci", "bar", "json", "off"
	NoColor bool
	NoEmoji bool
}

// GetFormatter returns the appropriate formatter based on options and environment
func GetFormatter(opts Options) Formatter {
	// Handle explicit format selection
	switch opts.Format {
	case "off":
		return nil
	case "json":
		return NewJSONFormatter()
	case "ci":
		return NewCIFormatter()
	case "bar":
		formatter := NewBarFormatter()
		formatter.NoColor = opts.NoColor
		formatter.NoEmoji = opts.NoEmoji
		return formatter
	}
	
	// Auto-detection (default)
	// Check if we're in a CI environment or output is not a TTY
	if os.Getenv("CI") != "" || !isatty.IsTerminal(os.Stderr.Fd()) {
		return NewCIFormatter()
	}
	
	// Interactive terminal - use bar formatter
	formatter := NewBarFormatter()
	formatter.NoColor = opts.NoColor || os.Getenv("NO_COLOR") != ""
	formatter.NoEmoji = opts.NoEmoji
	return formatter
}

// Print outputs the summary using the specified formatter
func Print(w io.Writer, stats Stats, opts Options) error {
	formatter := GetFormatter(opts)
	if formatter == nil {
		return nil // "off" format
	}
	
	return formatter.Format(w, stats)
}

// EstimateTokens provides a rough estimate of token count from byte size
// Using the cl100k_base tokenizer approximation (GPT-4)
// This is a very rough estimate: ~1 token per 4 bytes for code
func EstimateTokens(bytes int64) int64 {
	if bytes == 0 {
		return 0
	}
	// For code, the ratio is approximately 1 token per 4 bytes
	// This is a conservative estimate that works well for most programming languages
	return bytes / 4
}