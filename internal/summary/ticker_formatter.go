package summary

import (
	"fmt"
	"io"
	
	"github.com/dustin/go-humanize"
)

// TickerFormatter formats summaries in stock ticker style
type TickerFormatter struct {
	NoEmoji bool
}

// NewTickerFormatter creates a new ticker formatter
func NewTickerFormatter() *TickerFormatter {
	return &TickerFormatter{}
}

// Format outputs a stock ticker style summary
func (f *TickerFormatter) Format(w io.Writer, stats Stats) error {
	ratio := getCompressionRatio(stats.OriginalBytes, stats.DistilledBytes)
	tokensSaved := stats.OriginalTokens - stats.DistilledTokens
	
	// Format: 📊 AID 97.6% ▲ │ SIZE: 5.2MB→128KB │ TIME: 450ms │ EST: ~1.17M tokens saved
	icon := "📊"
	if f.NoEmoji {
		icon = "[AID]"
	}
	
	arrow := "▲"
	if ratio < 50 {
		arrow = "▼"
	}
	
	fmt.Fprintf(w, "%s AID %.1f%% %s │ SIZE: %s→%s │ TIME: %s",
		icon,
		ratio,
		arrow,
		humanize.Bytes(uint64(stats.OriginalBytes)),
		humanize.Bytes(uint64(stats.DistilledBytes)),
		formatDuration(stats.Duration),
	)
	
	// Add token savings if available
	if tokensSaved > 0 {
		fmt.Fprintf(w, " │ EST: ~%s tokens saved",
			formatTokenCount(tokensSaved),
		)
	}
	
	fmt.Fprintln(w)
	
	// Add output path if not stdout on a new line
	if !stats.IsStdout && stats.OutputPath != "" {
		fileEmoji := "📄"
		if f.NoEmoji {
			fileEmoji = ">"
		}
		fmt.Fprintf(w, "%s %s\n", fileEmoji, stats.OutputPath)
	}
	return nil
}