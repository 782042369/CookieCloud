package logger

import (
	"fmt"
	"sync"
	"testing"
)

func TestError_Basic(t *testing.T) {
	// 测试基本错误日志
	// 由于日志是全局的，这里只测试不 panic
	Error("测试错误日志")
	_ = t // 避免未使用参数警告
}

func TestError_WithKeyValues(t *testing.T) {
	// 测试带 key-value 的错误日志
	Error("操作失败",
		"component", "storage",
		"operation", "save",
		"uuid", "test-uuid",
		"error", "test error")

	// 如果没有 panic，测试通过
	_ = t
}

func TestError_OddKeyValues(t *testing.T) {
	// 测试奇数个 key-value（应该补空字符串）
	Error("操作失败",
		"component", "storage",
		"operation") // 缺少 value

	// 如果没有 panic，测试通过
	_ = t
}

func TestRequestError(t *testing.T) {
	// 测试请求错误日志
	RequestError("/update", "POST", "127.0.0.1", "保存数据失败",
		fmt.Errorf("test error"))

	// 如果没有 panic，测试通过
	_ = t
}

func TestConcurrentLogging(t *testing.T) {
	// 测试并发日志记录
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			Error("并发测试",
				"index", idx,
				"error", "concurrent error")
		}(i)
	}
	wg.Wait()

	// 如果没有 panic 和死锁，测试通过
}
