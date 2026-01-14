package cache

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cache := New(5 * time.Minute)
	if cache == nil {
		t.Fatal("New() 返回 nil")
	}
	if cache.ttl != 5*time.Minute {
		t.Errorf("New() ttl = %v, 期望 %v", cache.ttl, 5*time.Minute)
	}
}

func TestSetAndGet(t *testing.T) {
	cache := New(5 * time.Minute)

	// 测试设置和获取
	cache.Set("uuid1", "encrypted1")

	data, ok := cache.Get("uuid1")
	if !ok {
		t.Fatal("Get() 返回 false, 期望 true")
	}
	if data != "encrypted1" {
		t.Errorf("Get() = %v, 期望 %v", data, "encrypted1")
	}

	// 测试获取不存在的key
	_, ok = cache.Get("uuid2")
	if ok {
		t.Fatal("Get() 不存在的key返回 true, 期望 false")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := New(10 * time.Millisecond) // 10毫秒过期

	// 设置缓存
	cache.Set("uuid1", "encrypted1")

	// 立即获取，应该存在
	_, ok := cache.Get("uuid1")
	if !ok {
		t.Fatal("缓存立即过期，期望存在")
	}

	// 等待过期
	time.Sleep(15 * time.Millisecond)

	// 再次获取，应该不存在
	_, ok = cache.Get("uuid1")
	if ok {
		t.Fatal("缓存未过期，期望已过期")
	}
}

func TestDelete(t *testing.T) {
	cache := New(5 * time.Minute)

	cache.Set("uuid1", "encrypted1")

	// 删除前应该存在
	_, ok := cache.Get("uuid1")
	if !ok {
		t.Fatal("删除前缓存不存在")
	}

	// 删除
	cache.Delete("uuid1")

	// 删除后应该不存在
	_, ok = cache.Get("uuid1")
	if ok {
		t.Fatal("删除后缓存仍存在")
	}
}

func TestClear(t *testing.T) {
	cache := New(5 * time.Minute)

	// 添加多个缓存项
	cache.Set("uuid1", "encrypted1")
	cache.Set("uuid2", "encrypted2")
	cache.Set("uuid3", "encrypted3")

	// 清空前应该有3项
	if size := cache.Size(); size != 3 {
		t.Errorf("清空前 Size() = %v, 期望 3", size)
	}

	// 清空
	cache.Clear()

	// 清空后应该为0
	if size := cache.Size(); size != 0 {
		t.Errorf("清空后 Size() = %v, 期望 0", size)
	}
}

func TestCleanExpired(t *testing.T) {
	cache := New(10 * time.Millisecond)

	// 添加多个缓存项
	cache.Set("uuid1", "encrypted1")
	cache.Set("uuid2", "encrypted2")

	// 等待部分过期
	time.Sleep(15 * time.Millisecond)

	// 添加新的缓存项
	cache.Set("uuid3", "encrypted3")

	// 清理过期项
	cache.CleanExpired()

	// 检查结果
	_, ok1 := cache.Get("uuid1")
	_, ok2 := cache.Get("uuid2")
	_, ok3 := cache.Get("uuid3")

	if ok1 || ok2 {
		t.Fatal("过期缓存未清理")
	}
	if !ok3 {
		t.Fatal("新缓存被清理")
	}
}

func TestSize(t *testing.T) {
	cache := New(5 * time.Minute)

	// 初始大小为0
	if size := cache.Size(); size != 0 {
		t.Errorf("初始 Size() = %v, 期望 0", size)
	}

	// 添加3项
	cache.Set("uuid1", "encrypted1")
	cache.Set("uuid2", "encrypted2")
	cache.Set("uuid3", "encrypted3")

	if size := cache.Size(); size != 3 {
		t.Errorf("添加3项后 Size() = %v, 期望 3", size)
	}

	// 删除1项
	cache.Delete("uuid1")

	if size := cache.Size(); size != 2 {
		t.Errorf("删除1项后 Size() = %v, 期望 2", size)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := New(5 * time.Minute)
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			uuid := "uuid" + string(rune(n))
			cache.Set(uuid, "encrypted"+string(rune(n)))
		}(i)
	}

	// 并发读取
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			uuid := "uuid" + string(rune(n))
			cache.Get(uuid)
		}(i)
	}

	wg.Wait()

	// 检查最终大小
	size := cache.Size()
	if size == 0 {
		t.Fatal("并发操作后缓存为空")
	}
}

func TestOverwrite(t *testing.T) {
	cache := New(5 * time.Minute)

	// 设置初始值
	cache.Set("uuid1", "encrypted1")

	// 覆盖
	cache.Set("uuid1", "encrypted2")

	data, ok := cache.Get("uuid1")
	if !ok {
		t.Fatal("Get() 返回 false")
	}
	if data != "encrypted2" {
		t.Errorf("覆盖后 Get() = %v, 期望 %v", data, "encrypted2")
	}
}
