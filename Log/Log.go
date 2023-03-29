/*
 *
 *     @file: Log.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/28 下午3:59
 *     @last modified: 2023/3/28 下午3:59
 *
 *
 *
 */

package Log

import (
	"fmt"
	"io"
	"os"
	"time"

	log "golang.org/x/exp/slog"
)

var output []io.Writer

// TxtTo sets the output destination for a new logger and return it
// You can set the output destination to any io.Writer,
// such as a file, a network connection, or a bytes.Buffer.
func TxtTo(opts log.HandlerOptions, writer ...io.Writer) *log.Logger {
	mw := io.MultiWriter(writer...)
	return log.New(opts.NewTextHandler(mw))
}

var level log.Level

// InitLog
// initialize the log file with fileMode and log level
// print information to output
// return a function to close the log file
// if error occurs, return error
func InitLog(filename string, filePerm os.FileMode, loglevel string, _output ...io.Writer) (func(), error) {

	switch loglevel {
	// case "Panic", "panic", "PANIC":
	// 	level = log.PanicLevel
	// case "Fatal", "fatal", "FATAL":
	// 	level = log.FatalLevel
	case "Error", "error", "ERROR":
		level = log.LevelError
	case "Warn", "warn", "WARN":
		level = log.LevelWarn
	case "Info", "info", "INFO":
		level = log.LevelInfo
	case "Debug", "debug", "DEBUG":
		level = log.LevelDebug
	case "Trace", "trace", "TRACE": // [deprecated]
		level = log.LevelDebug
	default:
		log.Error("invalid log level")
	}

	// output to log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
	if err != nil {
		return nil, err
	}

	cleanUp := func() {
		err := file.Close()
		fmt.Println("close log file")
		if err != nil {
			log.Error("failed to close log file ", err)
		}
	}

	// AddSource := false
	// if level <= log.LevelDebug {
	// 	AddSource = true
	// }

	opts := log.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: nil,
	}

	output = _output

	log.SetDefault(TxtTo(opts, file))
	log.Info(fmt.Sprintf("init log file at %s\n", filename))
	_, err = file.WriteString(fmt.Sprintf("---------start at %s---------\n", time.Now().Format(time.DateTime)))
	if err != nil {
		return cleanUp, err
	}

	return cleanUp, nil
}

func toOutput(l log.Level, v ...any) {
	if l >= level {
		if output != nil {
			mw := io.MultiWriter(output...)
			_, _ = fmt.Fprintln(mw, v...)
		}
	}
}

func Debug(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Debug(msg, vLeft...)
	toOutput(log.LevelDebug, v...)
}

func Debugf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Debug(msg)
	toOutput(log.LevelDebug, msg)
}

func Info(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Info(msg, vLeft...)
	toOutput(log.LevelInfo, v...)
}

func Infof(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Info(msg)
	toOutput(log.LevelInfo, msg)
}

func Error(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Error(msg, vLeft...)
	toOutput(log.LevelError, v...)
}

func Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Error(msg)
	toOutput(log.LevelError, msg)
}

func Trace(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Debug(msg, vLeft...)
	toOutput(log.LevelDebug, v...)
}

func Tracef(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Debug(msg)
	toOutput(log.LevelDebug, msg)
}

func Warn(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Warn(msg, vLeft...)
	toOutput(log.LevelWarn, v...)
}

func Warnf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Warn(msg)
	toOutput(log.LevelWarn, msg)
}

func Fatal(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Error(msg, vLeft...)
	toOutput(log.LevelError, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Error(msg)
	toOutput(log.LevelError, msg)
	os.Exit(1)
}

func String(Key string, Value string) log.Attr {
	return log.String(Key, Value)
}

func Time(Key string, Value time.Time) log.Attr {
	return log.Time(Key, Value)
}

func Duration(Key string, Value time.Duration) log.Attr {
	return log.Duration(Key, Value)
}

func Any(Key string, Value any) log.Attr {
	return log.Any(Key, Value)
}

func Bool(Key string, Value bool) log.Attr {
	return log.Bool(Key, Value)
}

func Float64(Key string, Value float64) log.Attr {
	return log.Float64(Key, Value)
}

func Group(Key string, as ...log.Attr) log.Attr {
	return log.Group(Key, as...)
}

func Int(Key string, Value int) log.Attr {
	return log.Int(Key, Value)
}

func Int64(key string, value int64) log.Attr {
	return log.Int64(key, value)
}

func Uint64(key string, value uint64) log.Attr {
	return log.Uint64(key, value)
}
