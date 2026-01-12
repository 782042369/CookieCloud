# 加密解密模块 (internal/crypto)

[根目录](../../CLAUDE.md) > [internal](../) > **crypto**

> 最后更新：2026-01-11 16:25:37

## 变更记录

### 2026-01-11 16:25:37
- 初始化模块文档

---

## 这个模块干啥的

这个模块负责数据的加密和解密处理，主要干这些事：

1. **生成密钥**：从 UUID 和密码派生加密密钥
2. **解密数据**：解密客户端加密的数据
3. **算法兼容**：和 CryptoJS 的加密算法保持兼容
4. **格式处理**：处理 OpenSSL 兼容的加密消息格式

---

## 对外接口

### Decrypt
解密 Cookie 数据

```go
func Decrypt(uuid, encrypted, password string) []byte
```

**参数**：
- `uuid`：用户设备的唯一标识符
- `encrypted`：Base64 编码的加密数据（CryptoJS 格式）
- `password`：解密密码

**返回**：
- `[]byte`：解密后的数据（JSON 格式）
- 解密失败时返回 `[]byte("{}")`

**密钥生成**：
```
key = MD5(uuid + "-" + password)[0:16]
```

---

## 加密算法详解

### 算法栈
```
AES-256-CBC
├── 密钥派生: EVP_BytesToKey (OpenSSL)
│   ├── Hash: MD5
│   ├── Key Length: 32 bytes
│   └── IV Length: 16 bytes
├── 加密模式: CBC (Cipher Block Chaining)
└── 填充: PKCS7
```

### 加密消息格式

CryptoJS 生成的加密数据格式：
```
Base64(
    "Salted__" +        // 8 字节魔术字符串
    [8 字节盐值] +       // 随机盐值
    [实际密文]          // PKCS7 填充的密文
)
```

### 解密流程


---

## 核心函数

### decryptCryptoJsAesMsg
解密 CryptoJS 生成的 AES 加密消息

```go
func decryptCryptoJsAesMsg(password string, ciphertext string) ([]byte, error)
```

**功能**：
1. Base64 解码密文
2. 验证格式（以 "Salted__" 开头）
3. 提取盐值和实际密文
4. 使用 EVP_BytesToKey 派生密钥和 IV
5. AES-256-CBC 解密
6. 移除 PKCS7 填充

**错误处理**：
- Base64 解码失败 → 返回错误
- 格式验证失败 → 返回错误
- 解密失败 → 返回错误
- 填充移除失败 → 返回错误（可能是密码错误）

---

### bytesToKey
实现 OpenSSL EVP_BytesToKey 密钥派生算法

```go
func bytesToKey(salt, data []byte, h hash.Hash, keyLen, blockLen int) (key, iv []byte)
```

**参数**：
- `salt`：8 字节盐值
- `data`：密码数据
- `h`：哈希函数（MD5）
- `keyLen`：密钥长度（32 字节）
- `blockLen`：块长度（16 字节）

**返回**：
- `key`：32 字节加密密钥
- `iv`：16 字节初始化向量

**算法**：
```
concat = ""
lastHash = ""

while len(concat) < keyLen + blockLen:
    lastHash = MD5(lastHash + data + salt)
    concat += lastHash

key = concat[0:keyLen]
iv = concat[keyLen:keyLen+blockLen]
```

---

### md5String
计算字符串的 MD5 哈希值

```go
func md5String(inputs ...string) string
```

**功能**：将多个字符串拼接后计算 MD5 哈希，返回十六进制字符串

**用途**：生成解密密钥

---

### pkcs7strip
移除 PKCS7 填充

```go
func pkcs7strip(data []byte, blockSize int) ([]byte, error)
```

**功能**：
1. 验证填充格式
2. 移除填充字节
3. 返回原始数据

**错误**：
- 数据为空
- 数据长度不是块长度的倍数
- 填充格式无效（密码错误时会出现）

---

## 常量定义

```go
const (
    pkcs5SaltLen = 8    // PKCS5 盐值长度
    aes256KeyLen = 32   // AES-256 密钥长度
)
```

---

## 关键依赖

```go
import (
    "bytes"           // 字节缓冲操作
    "crypto/aes"       // AES 加密算法
    "crypto/cipher"    // 加密接口
    "crypto/md5"       // MD5 哈希
    "encoding/base64"  // Base64 编解码
    "encoding/hex"     // 十六进制编解码
    "errors"           // 错误处理
    "hash"             // 哈希接口
    "io"               // I/O 操作
    "strings"          // 字符串操作
)
```

---

## 兼容性要求

- 必须与 CryptoJS 的加密格式兼容
- 必须与 OpenSSL 的 EVP_BytesToKey 算法兼容
- 必须正确处理 PKCS7 填充

---

## 代码结构

### 文件组织
```
internal/crypto/
└── crypto.go         # 加密解密功能
```

### 常量列表

| 常量名 | 值 | 说明 |
|-------|---|------|
| `pkcs5SaltLen` | `8` | PKCS5 盐值长度 |
| `aes256KeyLen` | `32` | AES-256 密钥长度 |

### 函数列表

| 函数名 | 可见性 | 说明 |
|-------|-------|------|
| `Decrypt` | ✅ 公开 | 解密 Cookie 数据 |
| `decryptCryptoJsAesMsg` | 🔒 私有 | 解密 CryptoJS 格式消息 |
| `bytesToKey` | 🔒 私有 | EVP_BytesToKey 密钥派生 |
| `md5String` | 🔒 私有 | MD5 哈希计算 |
| `pkcs7strip` | 🔒 私有 | 移除 PKCS7 填充 |

---

## 常见问题

### Q1: 为什么使用 MD5 而不是更安全的哈希算法？
**A**: 为了与 CryptoJS 和 OpenSSL 的默认实现保持兼容。CryptoJS 使用 MD5 作为 EVP_BytesToKey 的默认哈希函数。

### Q2: 密钥为什么只取前 16 字节？
**A**: 这是为了与原版 Node.js 实现保持一致。虽然 AES-256 支持 32 字节密钥，但当前实现使用 16 字节密钥（AES-128）。

### Q3: 解密失败时为什么不返回错误？
**A**: 这是为了简化错误处理流程。调用方只需要检查返回值是否为 `{}` 即可判断解密是否成功。

---

## 加密算法兼容性

### 与 CryptoJS 的对应关系

| CryptoJS | 本实现 |
|---------|-------|
| `CryptoJS.AES.encrypt(msg, password)` | `decryptCryptoJsAesMsg()` |
| 默认模式：CBC | AES-CBC |
| 默认填充：Pkcs7 | PKCS7 |
| 密钥派生：EVP_BytesToKey | `bytesToKey()` |
| 哈希算法：MD5 | MD5 |

---

## 相关文件清单

### 源代码文件
- `internal/crypto/crypto.go` - 加密解密功能

### 依赖模块
- `internal/handlers/handlers.go` - 调用 Decrypt 函数

### 参考资源
- [CryptoJS 文档](https://cryptojs.gitbook.io/docs/)
- [OpenSSL EVP_BytesToKey](https://www.openssl.org/docs/man3.0/man3/EVP_BytesToKey.html)
- [PKCS#7 填充](https://datatracker.ietf.org/doc/html/rfc2315)

---

**模块维护者**：782042369
**最后审核**：2026-01-11
**文档版本**：1.0.0
