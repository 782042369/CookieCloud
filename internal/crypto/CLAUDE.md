# internal/crypto

[根目录](../../CLAUDE.md) > **internal/crypto**

## 模块快照

**职责**：AES-256-CBC 加密解密，兼容 CryptoJS 和 OpenSSL

**入口文件**：`crypto.go`（130 行）

**测试文件**：`crypto_test.go`

## 核心功能

### 1. 解密 Cookie 数据

```go
decrypted := crypto.Decrypt(uuid, encrypted, password)
```

- 返回解密后的 JSON 数据（`[]byte`）
- 解密失败返回 `{}`

**密钥生成**：
```go
key := md5(uuid + "-" + password)[:16]
```

### 2. CryptoJS 兼容性

**加密格式**（Base64 编码）：
```
"Salted__" + [8字节盐值] + [PKCS7填充的密文]
```

**密钥派生**：OpenSSL `EVP_BytesToKey`
- 算法：MD5
- 密钥长度：32 字节
- IV 长度：16 字节

**加密算法**：AES-256-CBC

## 实现细节

### bytesToKey（密钥派生）

```go
func bytesToKey(salt, data []byte, h hash.Hash, keyLen, blockLen int) (key, iv []byte)
```

循环哈希直到生成足够的密钥和 IV：
```
result = MD5(lastHash + data + salt)
```

### pkcs7strip（去除填充）

验证并移除 PKCS7 填充：
- 填充长度 ≤ blockSize
- 所有填充字节 = 填充长度

## 安全注意事项

- MD5 用于密钥派生（兼容性要求），但仅用于内部密钥生成
- 用户密钥：`MD5(uuid + "-" + password)[:16]`（16 字节）
- 实际加密密钥：32 字节（通过 EVP_BytesToKey 派生）

## 参考文档

- 完整实现：@internal/crypto/crypto.go
- 测试文件：@internal/crypto/crypto_test.go
- CryptoJS 文档：https://cryptojs.gitbook.io/docs/
