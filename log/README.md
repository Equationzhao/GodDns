# Log

## Description

wrapper for [slog] 

## Usage

```go
log.Infof("hello %s", "world") -> time=2023-03-28T15:39:15.380+08:00 level=INFO msg="hello world"

log.Info("hello", "toWhom" ,"world") -> time=2023-03-28T15:39:15.380+08:00 level=INFO msg="hello" toWhom="world"

log.Info("hello", "toWhom" ,"world", "age", 18) -> time=2023-03-28T15:39:15.380+08:00 level=INFO msg="hello" toWhom="world" age=18

log.Info("hello", log.String("toWhom", "world"), log.Int("age", 18), log.Bool("isMale", true)) -> time=2023-03-28T15:39:15.380+08:00 level=INFO msg="hello" toWhom="world" age=18 isMale=true

// ......

```