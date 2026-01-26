// Package handlers 提供 HTTP 请求处理功能
// 包含路由处理器、请求验证和响应格式化
package handlers

import (
	"cookiecloud/internal/cache"
	"cookiecloud/internal/crypto"
	"cookiecloud/internal/logger"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
	"unicode/utf8"
)

const (
	// MaxUUIDLength UUID的最大长度（防止超长字符串攻击）
	MaxUUIDLength = 256
	// MaxEncryptedDataLength 加密数据的最大长度（10MB）
	MaxEncryptedDataLength = 10 * 1024 * 1024
)

// Handlers 处理器集合，持有依赖的存储实例和缓存实例
type Handlers struct {
	store *storage.Storage
	cache *cache.Cache
}

// New 创建一个新的 Handlers 实例（依赖注入 storage 和 cache）
func New(store *storage.Storage, cache *cache.Cache) *Handlers {
	return &Handlers{
		store: store,
		cache: cache,
	}
}

// FiberRootHandler 根路径处理器，返回欢迎信息
func FiberRootHandler(apiRoot string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString("Hello World! API ROOT = " + apiRoot)
	}
}

// UpdateRequest 更新请求的数据结构
type UpdateRequest struct {
	Encrypted string `json:"encrypted"`
	UUID      string `json:"uuid"`
}

// DecryptRequest 解密请求的数据结构
type DecryptRequest struct {
	Password string `json:"password"`
}

// FiberUpdateHandler 处理更新请求，保存加密数据并更新缓存
// 使用 Fiber 的 Context 获取标准库的 context.Context
func (h *Handlers) FiberUpdateHandler(c *fiber.Ctx) error {
	var req UpdateRequest

	if err := c.BodyParser(&req); err != nil {
		logger.RequestError(c.Path(), c.Method(), c.IP(), "JSON解析失败", err)
		return sendError(c, fiber.StatusBadRequest, "Bad Request: failed to parse JSON")
	}

	if req.Encrypted == "" || req.UUID == "" {
		logger.Error("参数缺失", "path", c.Path(), "method", c.Method(), "ip", c.IP())
		return sendError(c, fiber.StatusBadRequest, "Bad Request: both 'encrypted' and 'uuid' fields are required")
	}

	if !validateUUID(req.UUID) {
		logger.Error("UUID长度超限", "uuid", req.UUID, "ip", c.IP())
		return sendError(c, fiber.StatusBadRequest, "Bad Request: uuid length exceeds maximum limit")
	}

	if len(req.Encrypted) > MaxEncryptedDataLength {
		logger.Error("加密数据长度超限", "uuid", req.UUID, "length", len(req.Encrypted), "ip", c.IP())
		return sendError(c, fiber.StatusBadRequest, "Bad Request: encrypted data length exceeds maximum limit")
	}

	// 获取 Fiber 的标准 context.Context
	ctx := c.Context()

	if err := h.store.SaveEncryptedData(ctx, req.UUID, req.Encrypted); err != nil {
		logger.RequestError(c.Path(), c.Method(), c.IP(), "文件写入失败", err)
		return sendError(c, fiber.StatusInternalServerError, "Internal Server Error: failed to save data")
	}

	h.cache.Set(req.UUID, req.Encrypted)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"action": "done"})
}

// FiberGetHandler 处理获取数据请求，优先从缓存读取
// 使用 Fiber 的 Context 获取标准库的 context.Context
func (h *Handlers) FiberGetHandler(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return sendError(c, fiber.StatusBadRequest, "Bad Request: 'uuid' is required")
	}

	if !validateUUID(uuid) {
		logger.Error("UUID长度超限", "uuid", uuid, "ip", c.IP())
		return sendError(c, fiber.StatusBadRequest, "Bad Request: uuid length exceeds maximum limit")
	}

	encrypted, found := h.cache.Get(uuid)
	if !found {
		// 获取 Fiber 的标准 context.Context
		ctx := c.Context()

		data, err := h.store.LoadEncryptedData(ctx, uuid)
		if err != nil {
			logger.Error("数据不存在", "uuid", uuid, "ip", c.IP())
			return sendError(c, fiber.StatusNotFound, "Not Found: data not found")
		}
		encrypted = data.Encrypted
		h.cache.Set(uuid, encrypted)
	}

	if c.Method() == "POST" {
		var req DecryptRequest
		if err := c.BodyParser(&req); err != nil {
			logger.RequestError(c.Path(), c.Method(), c.IP(), "JSON解析失败", err)
			return sendError(c, fiber.StatusBadRequest, "Bad Request: failed to parse JSON")
		}
		if req.Password != "" {
			decrypted := crypto.Decrypt(uuid, encrypted, req.Password)
			c.Set("Content-Type", "application/json")
			return c.Send(decrypted)
		}
	}

	return c.JSON(&storage.CookieData{Encrypted: encrypted})
}

// sendError 统一错误响应处理
func sendError(ctx *fiber.Ctx, statusCode int, reason string) error {
	return ctx.Status(statusCode).JSON(fiber.Map{
		"action": "error",
		"reason": reason,
	})
}

// validateUUID 验证UUID长度
func validateUUID(uuid string) bool {
	return utf8.RuneCountInString(uuid) <= MaxUUIDLength
}
