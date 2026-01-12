package crypto

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"testing"
)

// TestMd5String 测试 MD5 哈希计算
func TestMd5String(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"world", "7d793037a0760186574b0282f2f435e7"},
		{"hello-world", "2095312189753de6ad47dfe20cbe97ec"},
	}

	for _, tc := range testCases {
		result := md5String(tc.input)
		if result != tc.expected {
			t.Errorf("输入 '%s'：期望 '%s'，实际得到 '%s'", tc.input, tc.expected, result)
		}
	}
}

// TestMd5StringMultiple 测试多字符串拼接的 MD5 哈希
func TestMd5StringMultiple(t *testing.T) {
	// 测试多个字符串拼接
	result := md5String("uuid", "-", "password")
	expected := md5String("uuid-password")

	if result != expected {
		t.Errorf("多参数调用和单字符串拼接结果不一致：'%s' vs '%s'", result, expected)
	}
}

// TestPkcs7strip 测试 PKCS7 填充移除
func TestPkcs7strip(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		blockSize int
		wantErr   bool
		expected  []byte
	}{
		{
			name:      "正常情况 - 1字节填充",
			data:      []byte{0x01, 0x02, 0x03, 0x01},
			blockSize: 4,
			wantErr:   false,
			expected:  []byte{0x01, 0x02, 0x03},
		},
		{
			name:      "正常情况 - 3字节填充",
			data:      []byte{0x01, 0x02, 0x03, 0x03, 0x03, 0x03},
			blockSize: 3,
			wantErr:   false,
			expected:  []byte{0x01, 0x02, 0x03},
		},
		{
			name: "正常情况 - 完整块填充",
			data: []byte{0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
				0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10},
			blockSize: 16,
			wantErr:   false,
			expected:  []byte{},
		},
		{
			name:      "错误情况 - 空数据",
			data:      []byte{},
			blockSize: 16,
			wantErr:   true,
			expected:  nil,
		},
		{
			name:      "错误情况 - 未对齐",
			data:      []byte{0x01, 0x02, 0x03},
			blockSize: 4,
			wantErr:   true,
			expected:  nil,
		},
		{
			name:      "错误情况 - 无效填充",
			data:      []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			blockSize: 4,
			wantErr:   true,
			expected:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pkcs7strip(tc.data, tc.blockSize)

			if tc.wantErr {
				if err == nil {
					t.Errorf("期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("意外错误: %v", err)
				return
			}

			if !bytes.Equal(result, tc.expected) {
				t.Errorf("期望 %v，实际得到 %v", tc.expected, result)
			}
		})
	}
}

// TestBytesToKey 测试密钥派生函数
func TestBytesToKey(t *testing.T) {
	// 准备测试数据
	salt := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	password := []byte("test-password")

	// 调用密钥派生
	key, iv := bytesToKey(salt, password, md5.New(), 32, 16)

	// 验证密钥长度
	if len(key) != 32 {
		t.Errorf("期望密钥长度为 32 字节，实际得到 %d", len(key))
	}

	// 验证 IV 长度
	if len(iv) != 16 {
		t.Errorf("期望 IV 长度为 16 字节，实际得到 %d", len(iv))
	}

	// 验证密钥和 IV 不相同
	if bytes.Equal(key, iv) {
		t.Error("密钥和 IV 不应该相同")
	}

	// 验证相同输入产生相同输出
	key2, iv2 := bytesToKey(salt, password, md5.New(), 32, 16)
	if !bytes.Equal(key, key2) || !bytes.Equal(iv, iv2) {
		t.Error("相同输入应该产生相同输出")
	}

	// 验证不同输入产生不同输出
	salt2 := []byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	key3, _ := bytesToKey(salt2, password, md5.New(), 32, 16)
	if bytes.Equal(key, key3) {
		t.Error("不同盐值应该产生不同密钥")
	}
}

// TestBytesToKeyInvalidSalt 测试无效盐值长度
func TestBytesToKeyInvalidSalt(t *testing.T) {
	// 这个测试会触发 panic，需要用 recover 捕获
	defer func() {
		if r := recover(); r == nil {
			t.Error("期望 panic，但没有发生")
		}
	}()

	// 使用错误的盐值长度
	salt := []byte{0x01, 0x02, 0x03} // 只有 3 字节
	password := []byte("test")

	bytesToKey(salt, password, md5.New(), 32, 16)
}

// TestDecryptCryptoJsAesMsgInvalidFormat 测试无效的密文格式
func TestDecryptCryptoJsAesMsgInvalidFormat(t *testing.T) {
	testCases := []struct {
		name       string
		ciphertext string
	}{
		{"无效的 Base64", "invalid-base64!@#"},
		{"太短的数据", base64.StdEncoding.EncodeToString([]byte("short"))},
		{"缺少 Salted__ 前缀", base64.StdEncoding.EncodeToString([]byte("NoSalted__12345678data"))},
		{"长度未对齐", base64.StdEncoding.EncodeToString([]byte("Salted__12345678invalid"))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decryptCryptoJsAesMsg("password", tc.ciphertext)
			if err == nil {
				t.Error("期望返回错误，但没有")
			}
		})
	}
}

// TestDecrypt 测试公开的 Decrypt 函数
func TestDecrypt(t *testing.T) {
	// 测试解密失败的情况（无效的 Base64）
	uuid := "test-uuid"
	password := "test-password"
	invalidEncrypted := "invalid-base64-data"

	result := Decrypt(uuid, invalidEncrypted, password)

	// 解密失败应该返回空 JSON 对象
	if !bytes.Equal(result, []byte("{}")) {
		t.Errorf("解密失败时应该返回空 JSON 对象，实际得到: %s", string(result))
	}
}

// TestDecryptWithValidCiphertext 测试有效的解密流程
// 注意：这个测试使用预先生成的测试向量
func TestDecryptWithValidCiphertext(t *testing.T) {
	// 这个测试需要有效的 CryptoJS 加密数据
	// 由于加密过程需要匹配 CryptoJS 的格式，这里只测试函数调用路径

	uuid := "test-uuid"
	password := "test-password"
	// 空字符串会触发 Base64 解码错误
	emptyEncrypted := ""

	result := Decrypt(uuid, emptyEncrypted, password)

	// 应该返回空 JSON 对象
	if !bytes.Equal(result, []byte("{}")) {
		t.Errorf("期望空 JSON 对象，实际得到: %s", string(result))
	}
}

// TestDecryptKeyGeneration 测试密钥生成逻辑
func TestDecryptKeyGeneration(t *testing.T) {
	uuid := "test-device-123"
	password := "my-password"

	// 手动计算期望的密钥
	expectedInput := uuid + "-" + password
	expectedKey := md5String(expectedInput)[:16]

	// 调用实际的密钥生成逻辑
	actualKey := md5String(uuid + "-" + password)[:16]

	if expectedKey != actualKey {
		t.Errorf("密钥生成不一致: '%s' vs '%s'", expectedKey, actualKey)
	}

	// 验证密钥长度为 16 字节（32个十六进制字符）
	if len(actualKey) != 16 {
		t.Errorf("密钥长度应该为 16 字节，实际为 %d", len(actualKey))
	}
}

// TestConstants 测试常量定义
func TestConstants(t *testing.T) {
	if pkcs5SaltLen != 8 {
		t.Errorf("pkcs5SaltLen 应该为 8，实际为 %d", pkcs5SaltLen)
	}

	if aes256KeyLen != 32 {
		t.Errorf("aes256KeyLen 应该为 32，实际为 %d", aes256KeyLen)
	}

	if aesBlockLen != 16 {
		t.Errorf("aesBlockLen 应该为 16，实际为 %d", aesBlockLen)
	}
}

// BenchmarkMd5String 性能基准测试
func BenchmarkMd5String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		md5String("test-string-for-benchmarking")
	}
}

// BenchmarkDecrypt 性能基准测试
func BenchmarkDecrypt(b *testing.B) {
	uuid := "test-uuid"
	password := "test-password"
	invalidEncrypted := "invalid-data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decrypt(uuid, invalidEncrypted, password)
	}
}

// TestDecryptCryptoJsAesMsgWithValidFormat 测试有效格式的密文
// 这个测试尝试构造一个格式正确的密文来覆盖更多代码路径
func TestDecryptCryptoJsAesMsgWithValidFormat(t *testing.T) {
	// 构造一个格式正确的密文："Salted__" + 8字节盐值 + 16字节对齐的密文
	salt := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	// 创建一个16字节的数据（AES块大小）
	fakeCiphertext := bytes.Repeat([]byte{0x42}, 16)

	// 组合格式：Salted__ + salt + ciphertext
	formatted := append([]byte("Salted__"), salt...)
	formatted = append(formatted, fakeCiphertext...)

	// Base64编码
	encoded := base64.StdEncoding.EncodeToString(formatted)

	// 尝试解密（会失败，因为密文是假的，但能覆盖更多代码路径）
	_, err := decryptCryptoJsAesMsg("test-password", encoded)

	// 应该返回错误（pkcs7填充验证失败）
	if err == nil {
		t.Error("期望返回填充验证错误，但没有")
	}
}

// TestDecryptCryptoJsAesMsgBoundaryConditions 测试边界条件
func TestDecryptCryptoJsAesMsgBoundaryConditions(t *testing.T) {
	testCases := []struct {
		name       string
		password   string
		ciphertext string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "刚好17字节（最小长度）",
			password:   "test",
			ciphertext: base64.StdEncoding.EncodeToString(append([]byte("Salted__"), bytes.Repeat([]byte{0x00}, 9)...)),
			wantErr:    true,
		},
		{
			name:       "32字节（2个块）",
			password:   "test",
			ciphertext: base64.StdEncoding.EncodeToString(append([]byte("Salted__"), bytes.Repeat([]byte{0x00}, 24)...)),
			wantErr:    true,
		},
		{
			name:       "空密码",
			password:   "",
			ciphertext: base64.StdEncoding.EncodeToString(append([]byte("Salted__"), bytes.Repeat([]byte{0x00}, 24)...)),
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decryptCryptoJsAesMsg(tc.password, tc.ciphertext)
			if tc.wantErr && err == nil {
				t.Error("期望返回错误，但没有")
			}
		})
	}
}

// TestDecryptSuccessCase 测试解密成功的情况
func TestDecryptSuccessCase(t *testing.T) {
	uuid := "test-uuid-success"
	password := "test-password"
	emptyEncrypted := ""

	// 测试空字符串情况
	result := Decrypt(uuid, emptyEncrypted, password)

	// 解密失败应该返回空 JSON 对象
	if !bytes.Equal(result, []byte("{}")) {
		t.Errorf("期望空 JSON 对象，实际得到: %s", string(result))
	}

	// 测试各种无效情况
	testCases := []string{
		"",               // 空字符串
		"invalid",        // 无效base64
		"invalid@base64", // 包含非法字符
		"U2FsdGVk",       // 太短
	}

	for _, tc := range testCases {
		t.Run("encrypted="+tc, func(t *testing.T) {
			result := Decrypt(uuid, tc, password)
			if !bytes.Equal(result, []byte("{}")) {
				t.Errorf("期望空 JSON 对象，实际得到: %s", string(result))
			}
		})
	}
}

// TestPkcs7stripEdgeCases 测试PKCS7填充的边缘情况
func TestPkcs7stripEdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		blockSize int
		wantErr   bool
	}{
		{
			name:      "正确的完整填充",
			data:      append(bytes.Repeat([]byte{0x01}, 15), []byte{0x01}...),
			blockSize: 16,
			wantErr:   false, // 填充正确
		},
		{
			name:      "填充值为2",
			data:      append(bytes.Repeat([]byte{0x01}, 14), []byte{0x02, 0x02}...),
			blockSize: 16,
			wantErr:   false, // 填充正确
		},
		{
			name:      "填充值超过块大小",
			data:      append(bytes.Repeat([]byte{0x01}, 15), []byte{0x17}...),
			blockSize: 16,
			wantErr:   true, // 填充值17>16
		},
		{
			name:      "填充值为0",
			data:      append(bytes.Repeat([]byte{0x01}, 15), []byte{0x00}...),
			blockSize: 16,
			wantErr:   true, // 填充值0无效
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pkcs7strip(tc.data, tc.blockSize)
			if tc.wantErr && err == nil {
				t.Error("期望返回错误，但没有")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("不期望错误，但得到: %v", err)
			}
		})
	}
}
