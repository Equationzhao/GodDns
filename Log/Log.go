/*
 *     @Copyright
 *     @file: Log.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/17 下午9:54
 *     @last modified: 2023/3/17 下午8:07
 *
 *
 *
 */

package Log

import (
	"io"

	"github.com/sirupsen/logrus"
)

// To sets the output destination for the logger.
// You can set the output destination to any io.Writer,
// such as a file, a network connection, or a bytes.Buffer.
func To(logger *logrus.Logger, writer ...io.Writer) {
	mw := io.MultiWriter(writer...)
	logger.SetOutput(mw)
}
