package summary

import (
	"fmt"
	"io"
	"os"
	
	"github.com/dustin/go-humanize"
	"golang.org/x/term"
)

// BarFormatter formats summaries with visual progress bars for TTY environments
type BarFormatter struct {
	NoColor bool
	NoEmoji bool
}

// NewBarFormatter creates a new bar formatter
func NewBarFormatter() *BarFormatter {
	return &BarFormatter{
		NoColor: os.Getenv("NO_COLOR") != "",
		NoEmoji: false,
	}
}

// Format outputs a visually appealing summary with progress bar
func (f *BarFormatter) Format(w io.Writer, stats Stats) error {
	ratio := getCompressionRatio(stats.OriginalBytes, stats.DistilledBytes)
	
	// Get terminal width for responsive bar sizing
	barWidth := 40
	if fd, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(int(fd.Fd())); err == nil && width > 80 {
			// Use up to 50 chars for bar on wide terminals
			barWidth = min(50, width-60)
		}
	}
	
	// Build the components
	emoji := ""
	if !f.NoEmoji {
		emoji = getEmoji(ratio) + " "
	}
	
	bar := buildProgressBar(ratio, barWidth)
	
	// Determine what was processed
	subject := "Distilled"
	if stats.FileCount > 0 {
		if stats.FileCount == 1 {
			subject = "Distilled 1 file"
		} else {
			subject = fmt.Sprintf("Distilled %d files", stats.FileCount)
		}
	}
	
	// Format the line
	fmt.Fprintf(w, "%s%s [%s] %.1f%% (%s → %s) in %s",
		emoji,
		subject,
		bar,
		ratio,
		humanize.Bytes(uint64(stats.OriginalBytes)),
		humanize.Bytes(uint64(stats.DistilledBytes)),
		formatDuration(stats.Duration),
	)
	
	// Add token savings if significant
	tokensSaved := stats.OriginalTokens - stats.DistilledTokens
	if tokensSaved > 1000 {
		fmt.Fprintf(w, " 💰 ~%s tokens saved",
			formatTokenCount(tokensSaved),
		)
	}
	
	fmt.Fprintln(w)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}