package summary

import (
	"fmt"
	"io"
	
	"github.com/dustin/go-humanize"
)

// DashboardFormatter formats summaries as a speedometer dashboard
type DashboardFormatter struct {
	NoEmoji bool
}

// NewDashboardFormatter creates a new dashboard formatter
func NewDashboardFormatter() *DashboardFormatter {
	return &DashboardFormatter{}
}

// Format outputs a speedometer dashboard style summary
func (f *DashboardFormatter) Format(w io.Writer, stats Stats) error {
	ratio := getCompressionRatio(stats.OriginalBytes, stats.DistilledBytes)
	
	// Calculate speed percentage (base: 1000ms = 0%, 0ms = 100%)
	speedPct := 100.0
	if stats.Duration.Milliseconds() > 0 {
		speedPct = 100.0 - (float64(stats.Duration.Milliseconds()) / 10.0)
		if speedPct < 0 {
			speedPct = 0
		}
		if speedPct > 100 {
			speedPct = 100
		}
	}
	
	// Build progress bars
	speedBar := buildProgressBar(speedPct, 10)
	savedBar := buildProgressBar(ratio, 10)
	
	// Format the dashboard
	fmt.Fprintln(w, "╔═══ AI Distiller ═══╗")
	fmt.Fprintf(w, "║ Speed: %s %3.0f%% ║ %s\n", 
		speedBar, 
		speedPct,
		formatDuration(stats.Duration),
	)
	fmt.Fprintf(w, "║ Saved: %s %3.1f%% ║ %s→%s\n",
		savedBar,
		ratio,
		humanize.Bytes(uint64(stats.OriginalBytes)),
		humanize.Bytes(uint64(stats.DistilledBytes)),
	)
	
	// Add token savings if available
	if stats.OriginalTokens > 0 && stats.DistilledTokens > 0 {
		tokensSaved := stats.OriginalTokens - stats.DistilledTokens
		fmt.Fprintf(w, "║ Tokens saved: ~%-8s ║\n", formatTokenCount(tokensSaved))
	}
	
	fmt.Fprintln(w, "╚═════════════════════╝")
	
	// Add output path if not stdout below the box
	if !stats.IsStdout && stats.OutputPath != "" {
		fileEmoji := "📄"
		if f.NoEmoji {
			fileEmoji = ""
		}
		fmt.Fprintf(w, "%s %s\n", fileEmoji, stats.OutputPath)
	}
	
	return nil
}