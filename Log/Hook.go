/*
 *     @Copyright
 *     @file: Hook.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午8:04
 *     @last modified: 2023/3/17 下午8:03
 *
 *
 *
 */

package Log

import (
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

// LogrusOriginally2writer
// implements the logrus.Hook interface
// It is used to write the original message to the writer
type LogrusOriginally2writer struct {
	writers []io.Writer
}

func (l LogrusOriginally2writer) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire writes the original message to the writer
// It is used by LogrusOriginally2writer

func (l LogrusOriginally2writer) Fire(entry *logrus.Entry) error {
	line := entry.Message
	var err error
	for _, writer := range l.writers {

		_, err_ := fmt.Fprintln(writer, line)
		if err_ != nil {
			err = errors.Join(err, err_)
		}
	}
	return err
}

// NewLogrusOriginally2writer
// creates a new instance of the LogrusOriginally2writer struct.
func NewLogrusOriginally2writer(writer ...io.Writer) *LogrusOriginally2writer {
	return &LogrusOriginally2writer{writers: writer}
}
