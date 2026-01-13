// Package logger 提供结构化错误日志记录功能
// 只记录错误级别日志，使用 key-value 格式便于解析
package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

var errorLogger = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)

// Error 记录错误日志（结构化 key-value 格式）
// 用法：logger.Error("保存数据失败", "uuid", uuid, "error", err)
func Error(msg string, keyvals ...interface{}) {
	if len(keyvals) == 0 {
		errorLogger.Println(msg)
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

	errorLogger.Println(buf.String())
}

// RequestError 记录 HTTP 请求错误（包含请求上下文）
func RequestError(path, method, ip, msg string, err error) {
	Error(msg, "path", path, "method", method, "ip", ip, "error", err)
}
