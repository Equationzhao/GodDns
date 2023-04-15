package Log

import (
	"fmt"
	"io"
	"os"
	"time"

	log "golang.org/x/exp/slog"
)

type Color = string

const (
	Red              = "\x1b[31m"
	Green            = "\x1b[32m"
	Yellow           = "\x1b[33m"
	White            = "\x1b[37m"
	DefaultColor     = White
	BackgroundPurple = "\x1b[45m"
)

type ColoredPrinter struct {
	Color   Color
	Disable bool
}

func (gp ColoredPrinter) Sprint(args ...any) string {
	if gp.Disable {
		return fmt.Sprint(args...)
	}
	return fmt.Sprint(gp.Color + fmt.Sprint(args...) + "\x1b[0m")
}

func (gp ColoredPrinter) Printf(format string, args ...interface{}) (int, error) {
	if gp.Disable {
		return fmt.Printf(format, args...)
	}
	return fmt.Printf(gp.Color+format+"\x1b[0m", args...)
}

func (gp ColoredPrinter) Println(args ...interface{}) (int, error) {
	if gp.Disable {
		return fmt.Println(args...)
	}
	return fmt.Println(gp.Color + fmt.Sprint(args...) + "\x1b[0m")
}

func (gp ColoredPrinter) Fprintf(writer io.Writer, format string, args ...interface{}) (int, error) {
	if gp.Disable {
		return fmt.Fprintf(writer, format, args...)
	}
	return fmt.Fprintf(writer, gp.Color+format+"\x1b[0m", args...)
}

func (gp ColoredPrinter) Fprintln(writer io.Writer, args ...interface{}) (int, error) {
	if gp.Disable {
		return fmt.Fprintln(writer, args...)
	}
	return fmt.Fprintln(writer, gp.Color+fmt.Sprint(args...)+"\x1b[0m")
}

var (
	SuccessPP = ColoredPrinter{Color: Green}
	// InfoPP is a pretty printer for info with default color(white)
	InfoPP = ColoredPrinter{Color: DefaultColor}
	// ErrPP is a pretty printer for error with red color
	ErrPP = ColoredPrinter{Color: Red}
	// WarnPP is a pretty printer for warning with yellow color
	WarnPP = ColoredPrinter{Color: Yellow}
	// DebugPP is a pretty printer for debug with green color
	DebugPP = ColoredPrinter{Color: BackgroundPurple}
)

var (
	output []io.Writer
	level  log.Level
)

// TxtTo sets the output destination for a new logger and return it
// You can set the output destination to any io.Writer,
// such as a file, a network connection, or a bytes.Buffer.
func TxtTo(opts log.HandlerOptions, writer ...io.Writer) *log.Logger {
	mw := io.MultiWriter(writer...)
	return log.New(opts.NewTextHandler(mw))
}

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
			switch l {
			case log.LevelError:
				_, _ = ErrPP.Fprintln(mw, v...)
			case log.LevelInfo:
				_, _ = InfoPP.Fprintln(mw, v...)
			case log.LevelWarn:
				_, _ = WarnPP.Fprintln(mw, v...)
			case log.LevelDebug:
				_, _ = DebugPP.Fprintln(mw, v...)

			default:
				_, _ = InfoPP.Fprintln(mw, v...)
			}
		}
	}
}

type Logger log.Logger

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	(*log.Logger)(l).Info(msg, keysAndValues...)
}

func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "error", err)
	(*log.Logger)(l).Error(msg, keysAndValues...)
}

func NewLogger(w io.Writer) *Logger {
	l := log.New(log.NewTextHandler(w))
	return (*Logger)(l)
}

func (l *Logger) WithGroup(g string) *Logger {
	newLogger := (*log.Logger)(l)
	newLogger = newLogger.WithGroup(g)
	return (*Logger)(newLogger)
}

func (l *Logger) Raw() *log.Logger {
	return (*log.Logger)(l)
}

func (l *Logger) Printf(msg string, v ...interface{}) {
	msg = fmt.Sprintf(msg, v...)
	(*log.Logger)(l).Info(msg)
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

func InfoRaw(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Info(msg, vLeft...)
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

func ErrorRaw(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Error(msg, vLeft...)
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

func WarnRaw(v ...any) {
	msg := fmt.Sprintf("%s", v[0])
	vLeft := v[1:]
	log.Warn(msg, vLeft...)
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

func String(key string, value string) log.Attr {
	return log.String(key, value)
}

func Time(key string, value time.Time) log.Attr {
	return log.Time(key, value)
}

func Duration(key string, value time.Duration) log.Attr {
	return log.Duration(key, value)
}

func Any(key string, value any) log.Attr {
	return log.Any(key, value)
}

func Bool(key string, value bool) log.Attr {
	return log.Bool(key, value)
}

func Float64(key string, value float64) log.Attr {
	return log.Float64(key, value)
}

func Group(key string, as ...log.Attr) log.Attr {
	return log.Group(key, as...)
}

func Int(key string, value int) log.Attr {
	return log.Int(key, value)
}

func Int64(key string, value int64) log.Attr {
	return log.Int64(key, value)
}

func Uint64(key string, value uint64) log.Attr {
	return log.Uint64(key, value)
}
