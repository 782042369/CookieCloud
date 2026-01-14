// Package logger 提供结构化日志记录功能
// 支持 INFO、WARN、ERROR 三个级别，使用 key-value 格式便于解析
package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

var (
	infoLogger  = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile)
	warnLogger  = log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
)

// Info 记录信息日志（结构化 key-value 格式）
// 用法：logger.Info("服务启动", "port", "8088")
func Info(msg string, keyvals ...interface{}) {
	logStructured(infoLogger, msg, keyvals...)
}

// Warn 记录警告日志（结构化 key-value 格式）
// 用法：logger.Warn("缓存未命中", "uuid", uuid)
func Warn(msg string, keyvals ...interface{}) {
	logStructured(warnLogger, msg, keyvals...)
}

// Error 记录错误日志（结构化 key-value 格式）
// 用法：logger.Error("保存数据失败", "uuid", uuid, "error", err)
func Error(msg string, keyvals ...interface{}) {
	logStructured(errorLogger, msg, keyvals...)
}

// RequestError 记录 HTTP 请求错误（包含请求上下文）
func RequestError(path, method, ip, msg string, err error) {
	Error(msg, "path", path, "method", method, "ip", ip, "error", err)
}

// logStructured 结构化日志记录的核心函数
func logStructured(logger *log.Logger, msg string, keyvals ...interface{}) {
	if len(keyvals) == 0 {
		logger.Println(msg)
		return
	}

	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "")
	}

	var buf bytes.Buffer
	buf.WriteString(msg)

	for i := 0; i < len(keyvals); i += 2 {
		buf.WriteString(" | ")
		fmt.Fprintf(&buf, "%v=%v", keyvals[i], keyvals[i+1])
	}

	logger.Println(buf.String())
}
