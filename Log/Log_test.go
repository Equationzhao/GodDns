package Log

import (
	"testing"
)

func TestInfo(t *testing.T) {
	Infof("hello %s", "world")

	Info("hello", "toWhom", "world")

	Info("hello", "toWhom", "world", "age", 18)

	Info("hello", String("toWhom", "world"), Int("age", 18), Bool("isMale", true))
}
