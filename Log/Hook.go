/*
 *     @Copyright
 *     @file: Hook.go
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
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

// LogrusOriginally2writer
// implements the logrus.Hook interface
// It is used to write the original message to the writer
// if writers is nil, panic raises when firing the hook.
type LogrusOriginally2writer struct {
	writers []io.Writer
}

// Levels return applied levels
func (l LogrusOriginally2writer) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire writes the original message to the writer
// It is used by LogrusOriginally2writer
// if writers is nil, panic raises when firing the hook.
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
