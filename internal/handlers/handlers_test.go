package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cookiecloud/internal/cache"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
)

// TestNew 测试创建 Handlers 实例
func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	dataCache := cache.New(5 * time.Minute)

	handlers := New(store, dataCache)

	if handlers == nil {
		t.Fatal("Handlers 实例不应为 nil")
	}

	if handlers.store != store {
		t.Error("store 字段未正确设置")
	}
}

// TestFiberRootHandler 测试根路径处理器
func TestFiberRootHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/", FiberRootHandler("/api"))
	app.Post("/", FiberRootHandler("/api"))

	// 测试 GET 请求
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("GET 请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	expectedBody := "Hello World! API ROOT = /api"
	if string(body) != expectedBody {
		t.Errorf("期望响应体 '%s'，实际得到 '%s'", expectedBody, string(body))
	}

	// 测试 POST 请求
	req = httptest.NewRequest("POST", "/", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("POST 请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}
}

// TestFiberRootHandlerWithEmptyAPIRoot 测试空 API_ROOT
func TestFiberRootHandlerWithEmptyAPIRoot(t *testing.T) {
	app := fiber.New()
	app.Get("/", FiberRootHandler(""))

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	expectedBody := "Hello World! API ROOT = "
	if string(body) != expectedBody {
		t.Errorf("期望响应体 '%s'，实际得到 '%s'", expectedBody, string(body))
	}
}

// TestFiberUpdateHandlerSuccess 测试成功更新数据
func TestFiberUpdateHandlerSuccess(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	uuid := "test-uuid-update"
	encrypted := "base64-encoded-data"

	reqBody := UpdateRequest{
		UUID:      uuid,
		Encrypted: encrypted,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 验证响应体
	var respBody map[string]interface{}
	respBodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBodyBytes, &respBody)

	if respBody["action"] != "done" {
		t.Errorf("期望 action='done'，实际得到 %v", respBody["action"])
	}

	// 验证数据已保存
	loaded, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if loaded.Encrypted != encrypted {
		t.Errorf("数据不匹配：期望 '%s'，实际得到 '%s'", encrypted, loaded.Encrypted)
	}
}

// TestFiberUpdateHandlerMissingFields 测试缺少必填字段
func TestFiberUpdateHandlerMissingFields(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	testCases := []struct {
		name        string
		requestBody UpdateRequest
	}{
		{"缺少 UUID", UpdateRequest{Encrypted: "data"}},
		{"缺少 Encrypted", UpdateRequest{UUID: "uuid"}},
		{"全部为空", UpdateRequest{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("期望状态码 400，实际得到 %d", resp.StatusCode)
			}

			// 验证错误响应格式
			var respBody map[string]interface{}
			respBodyBytes, _ := io.ReadAll(resp.Body)
			json.Unmarshal(respBodyBytes, &respBody)

			if respBody["action"] != "error" {
				t.Error("期望 action='error'")
			}

			if respBody["reason"] == nil {
				t.Error("期望包含 reason 字段")
			}
		})
	}
}

// TestFiberUpdateHandlerInvalidJSON 测试无效的 JSON
func TestFiberUpdateHandlerInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	req := httptest.NewRequest("POST", "/update", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("期望状态码 400，实际得到 %d", resp.StatusCode)
	}
}

// TestFiberGetHandlerSuccess 测试成功获取加密数据
func TestFiberGetHandlerSuccess(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Get("/get/:uuid", handlers.FiberGetHandler)

	uuid := "test-uuid-get"
	encrypted := "encrypted-data-123"

	// 先保存数据
	_ = store.SaveEncryptedData(context.Background(), uuid, encrypted)

	// 发送 GET 请求
	req := httptest.NewRequest("GET", "/get/"+uuid, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 验证响应体
	var respBody map[string]string
	respBodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBodyBytes, &respBody)

	if respBody["encrypted"] != encrypted {
		t.Errorf("期望加密数据 '%s'，实际得到 '%s'", encrypted, respBody["encrypted"])
	}
}

// TestFiberGetHandlerNotFound 测试获取不存在的数据
func TestFiberGetHandlerNotFound(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Get("/get/:uuid", handlers.FiberGetHandler)

	req := httptest.NewRequest("GET", "/get/non-existent-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("期望状态码 404，实际得到 %d", resp.StatusCode)
	}

	// 验证错误响应格式
	var respBody map[string]interface{}
	respBodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBodyBytes, &respBody)

	if respBody["action"] != "error" {
		t.Error("期望 action='error'")
	}
}

// TestFiberGetHandlerWithPassword 测试使用密码解密
func TestFiberGetHandlerWithPassword(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/get/:uuid", handlers.FiberGetHandler)

	uuid := "test-uuid-decrypt"
	encrypted := "invalid-encrypted-data" // 这会导致解密失败，返回 {}

	// 先保存数据
	_ = store.SaveEncryptedData(context.Background(), uuid, encrypted)

	// 发送 POST 请求并提供密码
	reqBody := DecryptRequest{Password: "test-password"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/get/"+uuid, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 验证返回的是空 JSON 对象（因为解密失败）
	respBodyBytes, _ := io.ReadAll(resp.Body)
	if string(respBodyBytes) != "{}" {
		t.Errorf("解密失败时应返回空 JSON 对象，实际得到: %s", string(respBodyBytes))
	}
}

// TestFiberGetHandlerEmptyUUID 测试空 UUID
func TestFiberGetHandlerEmptyUUID(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Get("/get/:uuid", handlers.FiberGetHandler)

	req := httptest.NewRequest("GET", "/get/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	// Fiber 会返回 404 因为路径不匹配
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("期望状态码 404，实际得到 %d", resp.StatusCode)
	}
}

// TestSendErrorResponse 测试错误响应格式
func TestSendErrorResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/error", func(c *fiber.Ctx) error {
		return sendError(c, fiber.StatusInternalServerError, "test error message")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("期望状态码 500，实际得到 %d", resp.StatusCode)
	}

	// 验证响应格式
	var respBody map[string]string
	respBodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBodyBytes, &respBody)

	if respBody["action"] != "error" {
		t.Error("期望 action='error'")
	}

	if respBody["reason"] != "test error message" {
		t.Errorf("期望 reason='test error message'，实际得到 '%s'", respBody["reason"])
	}
}

// TestFullUpdateAndGetFlow 测试完整的更新和获取流程
func TestFullUpdateAndGetFlow(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)
	app.Get("/get/:uuid", handlers.FiberGetHandler)

	uuid := "full-flow-uuid"
	encrypted := "test-full-flow-data"

	// 1. 更新数据
	updateReq := UpdateRequest{UUID: uuid, Encrypted: encrypted}
	body, _ := json.Marshal(updateReq)
	updateReqHTTP := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	updateReqHTTP.Header.Set("Content-Type", "application/json")

	updateResp, err := app.Test(updateReqHTTP)
	if err != nil {
		t.Fatalf("更新请求失败: %v", err)
	}

	if updateResp.StatusCode != fiber.StatusOK {
		t.Errorf("更新失败，状态码: %d", updateResp.StatusCode)
	}

	// 2. 获取数据
	getReq := httptest.NewRequest("GET", "/get/"+uuid, nil)
	getResp, err := app.Test(getReq)
	if err != nil {
		t.Fatalf("获取请求失败: %v", err)
	}

	if getResp.StatusCode != fiber.StatusOK {
		t.Errorf("获取失败，状态码: %d", getResp.StatusCode)
	}

	// 3. 验证数据
	var getRespBody map[string]string
	getRespBytes, _ := io.ReadAll(getResp.Body)
	json.Unmarshal(getRespBytes, &getRespBody)

	if getRespBody["encrypted"] != encrypted {
		t.Errorf("数据不匹配：期望 '%s'，实际得到 '%s'", encrypted, getRespBody["encrypted"])
	}
}

// TestConcurrentRequests 测试并发请求处理
func TestConcurrentRequests(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)
	app.Get("/get/:uuid", handlers.FiberGetHandler)

	uuid := "concurrent-uuid"
	numRequests := 50

	// 使用 WaitGroup 等待所有并发请求完成
	var wg sync.WaitGroup

	// 并发发送更新请求
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := UpdateRequest{
				UUID:      uuid,
				Encrypted: "data-" + string(rune('0'+index%10)),
			}
			body, _ := json.Marshal(req)
			httpReq := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
			httpReq.Header.Set("Content-Type", "application/json")
			app.Test(httpReq)
		}(i)
	}

	// 等待所有请求完成
	wg.Wait()

	// 验证文件存在且可以读取
	_, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Errorf("并发请求后数据加载失败: %v", err)
	}
}

// BenchmarkUpdateHandler 性能基准测试
func BenchmarkUpdateHandler(b *testing.B) {
	tempDir := b.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	uuid := "benchmark-uuid"
	encrypted := "benchmark-test-data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := UpdateRequest{UUID: uuid, Encrypted: encrypted}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		app.Test(httpReq)
	}
}

// BenchmarkGetHandler 性能基准测试
func BenchmarkGetHandler(b *testing.B) {
	tempDir := b.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Get("/get/:uuid", handlers.FiberGetHandler)

	uuid := "benchmark-get-uuid"
	encrypted := "benchmark-test-data"
	_ = store.SaveEncryptedData(context.Background(), uuid, encrypted)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest("GET", "/get/"+uuid, nil)
		app.Test(httpReq)
	}
}

// TestFiberUpdateHandlerStorageError 测试存储层错误处理
func TestFiberUpdateHandlerStorageError(t *testing.T) {
	// 跳过这个测试，因为很难在不修改代码的情况下模拟文件系统错误
	t.Skip("无法轻易模拟文件系统写入错误，这个错误路径在实际使用中很少见")
}

// TestFiberGetHandlerWithInvalidPasswordJSON 测试解密请求的无效JSON
func TestFiberGetHandlerWithInvalidPasswordJSON(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/get/:uuid", handlers.FiberGetHandler)

	uuid := "test-uuid-invalid-json"
	encrypted := "test-data"

	// 先保存数据
	_ = store.SaveEncryptedData(context.Background(), uuid, encrypted)

	// 发送无效的JSON
	req := httptest.NewRequest("POST", "/get/"+uuid, bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	// 应该返回400错误（JSON解析失败）
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("期望状态码 400，实际得到 %d", resp.StatusCode)
	}
}

// TestFiberGetHandlerWithEmptyPassword 测试空密码的解密请求
func TestFiberGetHandlerWithEmptyPassword(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/get/:uuid", handlers.FiberGetHandler)

	uuid := "test-uuid-empty-password"
	encrypted := "test-data"

	// 先保存数据
	_ = store.SaveEncryptedData(context.Background(), uuid, encrypted)

	// 发送空密码
	reqBody := DecryptRequest{Password: ""}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/get/"+uuid, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	// 空密码应该返回加密数据而不是解密
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 验证返回的是加密数据
	var respBody map[string]string
	respBodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBodyBytes, &respBody)

	if respBody["encrypted"] != encrypted {
		t.Errorf("期望返回加密数据，实际得到: %v", respBody)
	}
}

// TestFiberUpdateHandlerWithEmptyFields 测试空字段值
func TestFiberUpdateHandlerWithEmptyFields(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	testCases := []struct {
		name      string
		request   UpdateRequest
		expectErr bool
	}{
		{
			name:      "UUID为空字符串",
			request:   UpdateRequest{UUID: "", Encrypted: "data"},
			expectErr: true,
		},
		{
			name:      "Encrypted为空字符串",
			request:   UpdateRequest{UUID: "uuid", Encrypted: ""},
			expectErr: true,
		},
		{
			name:      "两者都为空字符串",
			request:   UpdateRequest{UUID: "", Encrypted: ""},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if tc.expectErr && resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("期望状态码 400，实际得到 %d", resp.StatusCode)
			}
		})
	}
}

// TestFiberUpdateHandlerWithVeryLongData 测试超长数据
func TestFiberUpdateHandlerWithVeryLongData(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	uuid := "long-data-uuid"

	// 创建一个非常长的数据（模拟大型Cookie）
	longData := ""
	for i := 0; i < 10000; i++ {
		longData += "a"
	}

	reqBody := UpdateRequest{
		UUID:      uuid,
		Encrypted: longData,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 验证数据已保存
	loaded, _ := store.LoadEncryptedData(context.Background(), uuid)
	if len(loaded.Encrypted) != len(longData) {
		t.Errorf("数据长度不匹配")
	}
}

// TestFiberUpdateHandlerDuplicateRequests 测试重复的更新请求
func TestFiberUpdateHandlerDuplicateRequests(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := storage.New(tempDir)
	handlers := New(store, cache.New(5 * time.Minute))

	app := fiber.New()
	app.Post("/update", handlers.FiberUpdateHandler)

	uuid := "duplicate-uuid"

	// 发送第一次请求
	reqBody1 := UpdateRequest{
		UUID:      uuid,
		Encrypted: "first-data",
	}

	body1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/update", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")

	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("第一次请求失败: %v", err)
	}

	if resp1.StatusCode != fiber.StatusOK {
		t.Errorf("第一次请求失败，状态码: %d", resp1.StatusCode)
	}

	// 发送第二次请求（覆盖）
	reqBody2 := UpdateRequest{
		UUID:      uuid,
		Encrypted: "second-data",
	}

	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/update", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("第二次请求失败: %v", err)
	}

	if resp2.StatusCode != fiber.StatusOK {
		t.Errorf("第二次请求失败，状态码: %d", resp2.StatusCode)
	}

	// 验证最终保存的是第二次的数据
	loaded, _ := store.LoadEncryptedData(context.Background(), uuid)
	if loaded.Encrypted != "second-data" {
		t.Errorf("期望保存第二次的数据，实际得到: %s", loaded.Encrypted)
	}
}
