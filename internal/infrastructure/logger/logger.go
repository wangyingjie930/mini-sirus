package logger

import (
	"fmt"
	"log"
	"time"
)

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// SimpleLogger 简单日志实现
type SimpleLogger struct {
	prefix string
}

// NewSimpleLogger 创建简单日志
func NewSimpleLogger(prefix string) *SimpleLogger {
	return &SimpleLogger{
		prefix: prefix,
	}
}

// Info 信息日志
func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	l.log("INFO", msg, fields...)
}

// Error 错误日志
func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	l.log("ERROR", msg, fields...)
}

// Warn 警告日志
func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	l.log("WARN", msg, fields...)
}

// Debug 调试日志
func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {
	l.log("DEBUG", msg, fields...)
}

// log 统一日志输出
func (l *SimpleLogger) log(level string, msg string, fields ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}

	logMsg := fmt.Sprintf("[%s] %s%s: %s", timestamp, prefix, level, msg)
	if len(fields) > 0 {
		logMsg += fmt.Sprintf(" %v", fields)
	}

	log.Println(logMsg)
}

