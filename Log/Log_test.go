package Log

import (
	"os"
	"testing"
)

func TestInfo(t *testing.T) {
	Infof("hello %s", "world")

	Info("hello", "toWhom", "world")

	Info("hello", "toWhom", "world", "age", 18)

	Info("hello", String("toWhom", "world"), Int("age", 18), Bool("isMale", true))
}

func TestGreenPrinter_Fprintf(t *testing.T) {
	coloredPrinter := ColoredPrinter{Color: BackgroundPurple}
	coloredPrinter.Fprintf(os.Stdout, "hello %s\n", "world")
	coloredPrinter.Color = DefaultColor
	coloredPrinter.Fprintf(os.Stdout, "hello %s\n", "world")
	coloredPrinter.Color = Red
	coloredPrinter.Fprintf(os.Stdout, "hello %s\n", "world")
	coloredPrinter.Color = Yellow
	coloredPrinter.Fprintf(os.Stdout, "hello %s\n", "world")
	coloredPrinter.Color = Green
	coloredPrinter.Fprintf(os.Stdout, "hello %s\n", "world")
	coloredPrinter.Color = BackgroundPurple
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = DefaultColor
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = Red
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = Yellow
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = Green
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Disable = true
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = DefaultColor
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = Red
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = Yellow
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
	coloredPrinter.Color = Green
	coloredPrinter.Fprintln(os.Stdout, "hello ", "world")
}
