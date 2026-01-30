package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestNew 测试创建 Storage 实例
func TestNew(t *testing.T) {
	// 使用临时目录
	tempDir := t.TempDir()

	store, err := New(tempDir)

	if err != nil {
		t.Fatalf("创建 Storage 实例失败: %v", err)
	}

	if store == nil {
		t.Fatal("Storage 实例不应为 nil")
	}

	if store.dataDir != tempDir {
		t.Errorf("期望 dataDir 为 '%s'，实际得到 '%s'", tempDir, store.dataDir)
	}

	// 验证目录已创建
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("数据目录未创建")
	}
}

// TestNewWithInvalidPath 测试使用无效路径创建 Storage
func TestNewWithInvalidPath(t *testing.T) {
	// 使用一个不可能创建的路径（比如在只读文件系统中）
	// 这里我们用一个空字符串来触发错误
	_, err := New("")

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestSaveAndLoad 测试保存和加载数据
func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "test-uuid-123"
	encryptedData := "base64-encoded-encrypted-data"

	// 保存数据
	err := store.SaveEncryptedData(context.Background(), uuid, encryptedData)
	if err != nil {
		t.Fatalf("保存数据失败: %v", err)
	}

	// 验证文件存在
	filePath := filepath.Join(tempDir, uuid+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("数据文件未创建")
	}

	// 加载数据
	data, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if data.Encrypted != encryptedData {
		t.Errorf("期望加密数据为 '%s'，实际得到 '%s'", encryptedData, data.Encrypted)
	}
}

// TestLoadNonExistent 测试加载不存在的数据
func TestLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	// 尝试加载不存在的 UUID
	_, err := store.LoadEncryptedData(context.Background(), "non-existent-uuid")

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestOverwriteData 测试覆盖已存在的数据
func TestOverwriteData(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "test-uuid-456"

	// 保存初始数据
	err := store.SaveEncryptedData(context.Background(), uuid, "first-data")
	if err != nil {
		t.Fatalf("保存初始数据失败: %v", err)
	}

	// 覆盖为新数据
	newData := "second-data"
	err = store.SaveEncryptedData(context.Background(), uuid, newData)
	if err != nil {
		t.Fatalf("覆盖数据失败: %v", err)
	}

	// 加载并验证
	loaded, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if loaded.Encrypted != newData {
		t.Errorf("期望数据为 '%s'，实际得到 '%s'", newData, loaded.Encrypted)
	}
}

// TestConcurrentWrites 测试并发写入安全性
func TestConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "concurrent-uuid"
	numGoroutines := 100
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			data := "data-" + string(rune('0'+index%10))
			if err := store.SaveEncryptedData(context.Background(), uuid, data); err != nil {
				t.Errorf("并发保存失败: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 验证文件存在
	filePath := filepath.Join(tempDir, uuid+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("并发写入后文件不存在")
	}

	// 加载并验证数据格式
	_, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Errorf("并发写入后加载数据失败: %v", err)
	}
}

// TestMultipleUUIDs 测试多个 UUID 的独立存储
func TestMultipleUUIDs(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuids := []string{
		"uuid-1",
		"uuid-2",
		"uuid-3",
	}

	expectedData := map[string]string{
		"uuid-1": "encrypted-data-1",
		"uuid-2": "encrypted-data-2",
		"uuid-3": "encrypted-data-3",
	}

	// 保存所有数据
	for _, uuid := range uuids {
		err := store.SaveEncryptedData(context.Background(), uuid, expectedData[uuid])
		if err != nil {
			t.Fatalf("保存 %s 失败: %v", uuid, err)
		}
	}

	// 加载并验证所有数据
	for _, uuid := range uuids {
		data, err := store.LoadEncryptedData(context.Background(), uuid)
		if err != nil {
			t.Fatalf("加载 %s 失败: %v", uuid, err)
		}

		if data.Encrypted != expectedData[uuid] {
			t.Errorf("%s: 期望 '%s'，实际得到 '%s'", uuid, expectedData[uuid], data.Encrypted)
		}
	}
}

// TestJSONFormat 测试保存的文件格式
func TestJSONFormat(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "json-test-uuid"
	encryptedData := "test-encrypted-data"

	err := store.SaveEncryptedData(context.Background(), uuid, encryptedData)
	if err != nil {
		t.Fatalf("保存数据失败: %v", err)
	}

	// 读取文件内容
	filePath := filepath.Join(tempDir, uuid+".json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 验证 JSON 格式
	var cookieData CookieData
	err = json.Unmarshal(content, &cookieData)
	if err != nil {
		t.Fatalf("JSON 解析失败: %v", err)
	}

	if cookieData.Encrypted != encryptedData {
		t.Errorf("JSON 数据不匹配：期望 '%s'，实际得到 '%s'", encryptedData, cookieData.Encrypted)
	}
}

// TestSpecialCharactersInUUID 测试 UUID 中包含特殊字符
func TestSpecialCharactersInUUID(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	// 测试各种特殊字符
	testUUIDs := []string{
		"uuid-with-dash",
		"uuid_with_underscore",
		"uuid.with.dot",
		"uuid@with@at",
	}

	for _, uuid := range testUUIDs {
		encryptedData := "data-for-" + uuid
		err := store.SaveEncryptedData(context.Background(), uuid, encryptedData)
		if err != nil {
			t.Errorf("保存 %s 失败: %v", uuid, err)
		}

		loaded, err := store.LoadEncryptedData(context.Background(), uuid)
		if err != nil {
			t.Errorf("加载 %s 失败: %v", uuid, err)
			continue
		}

		if loaded.Encrypted != encryptedData {
			t.Errorf("%s 数据不匹配", uuid)
		}
	}
}

// TestFileLock 测试文件锁机制
func TestFileLock(t *testing.T) {
	uuid := "lock-test-uuid"

	// 获取同一个 UUID 的锁两次，应该返回相同的锁
	lock1 := getFileLock(uuid)
	lock2 := getFileLock(uuid)

	if lock1 != lock2 {
		t.Error("同一个 UUID 应该返回相同的锁")
	}

	// 不同 UUID 应该有不同的锁
	uuid2 := "lock-test-uuid-2"
	lock3 := getFileLock(uuid2)

	if lock1 == lock3 {
		t.Error("不同 UUID 应该有不同的锁")
	}
}

// TestEmptyData 测试空数据
func TestEmptyData(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "empty-data-uuid"
	emptyData := ""

	// 保存空数据
	err := store.SaveEncryptedData(context.Background(), uuid, emptyData)
	if err != nil {
		t.Fatalf("保存空数据失败: %v", err)
	}

	// 加载空数据
	data, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载空数据失败: %v", err)
	}

	if data.Encrypted != "" {
		t.Errorf("期望空字符串，实际得到 '%s'", data.Encrypted)
	}
}

// TestLongData 测试长数据
func TestLongData(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "long-data-uuid"

	// 创建一个较长的数据（模拟加密后的 Cookie 数据）
	longData := ""
	for i := 0; i < 1000; i++ {
		longData += "a"
	}

	err := store.SaveEncryptedData(context.Background(), uuid, longData)
	if err != nil {
		t.Fatalf("保存长数据失败: %v", err)
	}

	loaded, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载长数据失败: %v", err)
	}

	if len(loaded.Encrypted) != len(longData) {
		t.Errorf("数据长度不匹配：期望 %d，实际得到 %d", len(longData), len(loaded.Encrypted))
	}
}

// BenchmarkSaveAndLoad 性能基准测试
func BenchmarkSaveAndLoad(b *testing.B) {
	tempDir := b.TempDir()
	store, _ := New(tempDir)
	uuid := "benchmark-uuid"
	data := "test-encrypted-data-for-benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.SaveEncryptedData(context.Background(), uuid, data); err != nil {
			b.Fatalf("基准测试保存失败: %v", err)
		}
		if _, err := store.LoadEncryptedData(context.Background(), uuid); err != nil {
			b.Fatalf("基准测试加载失败: %v", err)
		}
	}
}

// BenchmarkConcurrentWrites 并发写入性能测试
func BenchmarkConcurrentWrites(b *testing.B) {
	tempDir := b.TempDir()
	store, _ := New(tempDir)
	uuid := "concurrent-benchmark-uuid"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := store.SaveEncryptedData(context.Background(), uuid, "test-data"); err != nil {
				b.Errorf("并发保存失败: %v", err)
			}
		}
	})
}

// TestLoadEncryptedDataInvalidJSON 测试加载无效的JSON数据
func TestLoadEncryptedDataInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "invalid-json-uuid"
	filePath := filepath.Join(tempDir, uuid+".json")

	// 创建一个无效的JSON文件
	err := os.WriteFile(filePath, []byte("{invalid json content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 尝试加载
	_, err = store.LoadEncryptedData(context.Background(), uuid)
	if err == nil {
		t.Error("期望返回JSON解析错误，但没有")
	}
}

// TestSaveEncryptedDataWithSpecialChars 测试保存包含特殊字符的数据
func TestSaveEncryptedDataWithSpecialChars(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "special-chars-uuid"

	// 包含各种特殊字符的数据
	specialData := "data with \"quotes\" and\nnewlines and\ttabs and \\backslashes\\ and /slashes/ and emoji 🎉"

	err := store.SaveEncryptedData(context.Background(), uuid, specialData)
	if err != nil {
		t.Fatalf("保存特殊字符数据失败: %v", err)
	}

	loaded, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载特殊字符数据失败: %v", err)
	}

	if loaded.Encrypted != specialData {
		t.Errorf("特殊字符数据不匹配")
	}
}

// TestLoadEncryptedDataEmptyFile 测试加载空文件
func TestLoadEncryptedDataEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "empty-file-uuid"
	filePath := filepath.Join(tempDir, uuid+".json")

	// 创建一个空文件
	err := os.WriteFile(filePath, []byte{}, 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 尝试加载（空文件会导致JSON解析失败）
	_, err = store.LoadEncryptedData(context.Background(), uuid)
	if err == nil {
		t.Error("期望返回JSON解析错误，但没有")
	}
}

// TestLoadEncryptedDataPartialJSON 测试加载部分JSON数据
func TestLoadEncryptedDataPartialJSON(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "partial-json-uuid"
	filePath := filepath.Join(tempDir, uuid+".json")

	// 创建一个缺少encrypted字段的JSON文件
	err := os.WriteFile(filePath, []byte(`{"other_field":"value"}`), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 尝试加载（应该成功，但Encrypted字段为空）
	loaded, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载部分JSON失败: %v", err)
	}

	if loaded.Encrypted != "" {
		t.Errorf("部分JSON应该得到空字符串，实际得到 '%s'", loaded.Encrypted)
	}
}

// TestMultipleReads 测试多次读取同一个文件
func TestMultipleReads(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "multi-read-uuid"
	data := "test-data-for-multiple-reads"

	// 保存数据
	err := store.SaveEncryptedData(context.Background(), uuid, data)
	if err != nil {
		t.Fatalf("保存数据失败: %v", err)
	}

	// 多次读取
	for i := 0; i < 10; i++ {
		loaded, err := store.LoadEncryptedData(context.Background(), uuid)
		if err != nil {
			t.Errorf("第%d次读取失败: %v", i+1, err)
		}
		if loaded.Encrypted != data {
			t.Errorf("第%d次读取数据不匹配", i+1)
		}
	}
}

// TestOverwriteWithDifferentSizes 测试用不同大小的数据覆盖
func TestOverwriteWithDifferentSizes(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := New(tempDir)

	uuid := "size-change-uuid"

	// 保存小数据
	smallData := "small"
	err := store.SaveEncryptedData(context.Background(), uuid, smallData)
	if err != nil {
		t.Fatalf("保存小数据失败: %v", err)
	}

	// 保存大数据
	largeData := ""
	for i := 0; i < 100; i++ {
		largeData += "x"
	}
	err = store.SaveEncryptedData(context.Background(), uuid, largeData)
	if err != nil {
		t.Fatalf("保存大数据失败: %v", err)
	}

	// 验证最终保存的是大数据
	loaded, err := store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if loaded.Encrypted != largeData {
		t.Error("覆盖后数据不匹配")
	}

	// 再保存小数据
	err = store.SaveEncryptedData(context.Background(), uuid, smallData)
	if err != nil {
		t.Fatalf("再次保存小数据失败: %v", err)
	}

	// 验证最终保存的是小数据
	loaded, err = store.LoadEncryptedData(context.Background(), uuid)
	if err != nil {
		t.Fatalf("再次加载数据失败: %v", err)
	}

	if loaded.Encrypted != smallData {
		t.Error("再次覆盖后数据不匹配")
	}
}
