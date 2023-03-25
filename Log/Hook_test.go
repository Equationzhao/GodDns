/*
 *     @Copyright
 *     @file: Hook_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/25 下午5:41
 *     @last modified: 2023/3/25 上午1:46
 *
 *
 *
 */

package Log

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"testing"
)

func TestLogrusOriginally2Writer(t *testing.T) {
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
