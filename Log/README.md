# Log

## Description

Log hook for logrus
send original message to `io.Writer(s)`

if io.Writer is nil, `Fire()` will panic when calling `logrus.info()` / `logrus.error()` /...
