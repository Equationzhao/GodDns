/*
 *     @Copyright
 *     @file: Hook_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/26 下午11:18
 *     @last modified: 2023/3/26 下午11:18
 *
 *
 *
 */

package Log

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogrusOriginally2Writer(t *testing.T) {

	// logrus.AddHook(NewLogrusOriginally2writer(nil)) panic

	logrus.SetLevel(logrus.DebugLevel)
	file, err := os.OpenFile("test.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Log(err)
	}
	writer := io.Writer(file)
	logrus.AddHook(NewLogrusOriginally2writer(os.Stdout, writer))
	logrus.Trace("test Trace level")
	logrus.Debug("test Debug level")
	logrus.Info("test Info level")
	logrus.Warn("test Warn level")
	logrus.Error("test Error level")

}

func BenchmarkLogrusOriginally2Writer(b *testing.B) {

	logrus.AddHook(NewLogrusOriginally2writer(os.Stdout))
	for i := 0; i < b.N; i++ {
		logrus.Trace("test Trace level")
		logrus.Debug("test Debug level")
		logrus.Info("test Info level")
		logrus.Warn("test Warn level")
		logrus.Error("test Error level")
	}

}

func BenchmarkFmtPln(b *testing.B) {

	for i := 0; i < b.N; i++ {
		logrus.Trace("test Trace level")
		fmt.Println("test Trace level")
		logrus.Debug("test Debug level")
		fmt.Println("test Debug level")
		logrus.Info("test Info level")
		fmt.Println("test Info level")
		logrus.Warn("test Warn level")
		fmt.Println("test Warn level")
		logrus.Error("test Error level")
		fmt.Println("test Error level")
	}

}
