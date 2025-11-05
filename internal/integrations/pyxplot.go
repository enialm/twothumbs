// File: internal/integrations/pyxplot.go

// This file contains code to execute Pyxplot commands.

package integrations

import (
	"fmt"
	"os"
	"os/exec"

	"twothumbs/internal/models"
)

// Produce the a digest plot using Pyxplot
func PlotAPlot(stats []models.PlotStats, filePath string) error {
	// Write the Pyxplot script with inline data to a temp file
	scriptFile, err := os.CreateTemp("", "plot_*.ppl")
	if err != nil {
		return fmt.Errorf("failed to create temp script file: %w", err)
	}
	defer os.Remove(scriptFile.Name())
	defer scriptFile.Close()

	fmt.Fprintln(scriptFile, "set terminal png dpi 600")
	fmt.Fprintf(scriptFile, "set output \"%s\"\n", filePath)
	fmt.Fprintln(scriptFile, "set width 11")
	fmt.Fprintln(scriptFile, "set xlabel \"Month\"")
	fmt.Fprintln(scriptFile, "set ylabel \"Score [\\%]\"")
	fmt.Fprintln(scriptFile, "set y2label \"Count [\\#]\"")
	fmt.Fprintln(scriptFile, "set key above")
	fmt.Fprintln(scriptFile, "set keycolumns 3")
	fmt.Fprintln(scriptFile, "set yrange [0:100]")
	fmt.Fprintln(scriptFile, "set y2range [0:*]")

	fmt.Fprintf(scriptFile, "set xtics (")
	for i, s := range stats {
		if i > 0 {
			fmt.Fprint(scriptFile, ", ")
		}
		fmt.Fprintf(scriptFile, "\"%s\" %d", s.Month.Format("Jan"), s.Month.Unix())
	}
	fmt.Fprintln(scriptFile, ")")

	fmt.Fprintln(scriptFile, "plot '--' using time.fromUnix($1):2 title 'Comment Count' with linespoints lw 2 pt 3 c Apricot axes xy2, \\")
	fmt.Fprintln(scriptFile, "     '--' using time.fromUnix($1):2 title 'Feedback Count' with linespoints lw 2 pt 4 c CarnationPink axes xy2, \\")
	fmt.Fprintln(scriptFile, "     '--' using time.fromUnix($1):2 title 'Feedback Score' with linespoints lw 2 pt 2 c Cerulean")

	// Feedback Score Data
	for _, s := range stats {
		fmt.Fprintf(scriptFile, "%d,%.15f\n", s.Month.Unix(), s.ThumbsUpPct)
	}
	fmt.Fprintln(scriptFile, "END")
	// Feedback Count Data
	for _, s := range stats {
		fmt.Fprintf(scriptFile, "%d,%d\n", s.Month.Unix(), s.NFeedback)
	}
	fmt.Fprintln(scriptFile, "END")
	// Comment Count Data
	for _, s := range stats {
		fmt.Fprintf(scriptFile, "%d,%d\n", s.Month.Unix(), s.NComments)
	}
	fmt.Fprintln(scriptFile, "END")

	scriptFile.Sync()

	// Run the script
	cmd := exec.Command("pyxplot", scriptFile.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pyxplot failed: %w", err)
	}

	// Add a white border using mogrify
	mogrifyCmd := exec.Command("mogrify", "-bordercolor", "white", "-border", "100x100", filePath)
	if err := mogrifyCmd.Run(); err != nil {
		return fmt.Errorf("mogrify failed: %w", err)
	}

	return nil
}
