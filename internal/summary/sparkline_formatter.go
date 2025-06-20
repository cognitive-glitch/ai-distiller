package summary

import (
	"fmt"
	"io"
	
	"github.com/dustin/go-humanize"
)

// SparklineFormatter formats summaries in minimalist sparkline style
type SparklineFormatter struct {
	NoEmoji bool
}

// NewSparklineFormatter creates a new sparkline formatter
func NewSparklineFormatter() *SparklineFormatter {
	return &SparklineFormatter{}
}

// Format outputs a minimalist sparkline style summary
func (f *SparklineFormatter) Format(w io.Writer, stats Stats) error {
	ratio := getCompressionRatio(stats.OriginalBytes, stats.DistilledBytes)
	
	// Format: ✨ Distilled in {time}ms. 📦 {original_size} → {distilled_size} ({ratio}% saved). 🎟️ Tokens: ~{original_tokens} → ~{distilled_tokens}.
	sparkEmoji := "✨"
	boxEmoji := "📦"
	ticketEmoji := "🎟️"
	
	if f.NoEmoji {
		sparkEmoji = "*"
		boxEmoji = ""
		ticketEmoji = ""
	}
	
	fmt.Fprintf(w, "%s Distilled in %s. ", 
		sparkEmoji,
		formatDuration(stats.Duration),
	)
	
	fmt.Fprintf(w, "%s %s → %s (%.1f%% saved). ",
		boxEmoji,
		humanize.Bytes(uint64(stats.OriginalBytes)),
		humanize.Bytes(uint64(stats.DistilledBytes)),
		ratio,
	)
	
	// Add token info if available
	if stats.OriginalTokens > 0 && stats.DistilledTokens > 0 {
		fmt.Fprintf(w, "%s Tokens: ~%s → ~%s.",
			ticketEmoji,
			formatTokenCount(stats.OriginalTokens),
			formatTokenCount(stats.DistilledTokens),
		)
	}
	
	fmt.Fprintln(w)
	return nil
}