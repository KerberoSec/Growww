package main
import "testing"
func TestWaterfallCascade(t *testing.T) {
	// Default waterfall: Margin -> SGF -> Insurance
	margin := 50000.0
	sgf := 100000.0
	insurance := 500000.0
	deficit := 75000.0
	covered := 0.0
	if deficit <= margin { covered = deficit } else { covered = margin; deficit -= margin
		if deficit <= sgf { covered += deficit } else { covered += sgf; deficit -= sgf; covered += min(deficit, insurance) }
	}
	if covered != 75000 { t.Errorf("expected 75000 covered, got %f", covered) }
}
func min(a, b float64) float64 { if a < b { return a }; return b }
