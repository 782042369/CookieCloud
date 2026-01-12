// Package handlers 提供 HTTP 请求处理功能
// 包含路由处理器、请求验证和响应格式化
package handlers

import (
	"log"

	"cookiecloud/internal/crypto"
	"cookiecloud/internal/storage"

	"github.com/gofiber/fiber/v2"
)

// Handlers 处理器集合，持有依赖的存储实例
type Handlers struct {
	store *storage.Storage
}

// New 创建一个新的 Handlers 实例（依赖注入 storage）
func New(store *storage.Storage) *Handlers {
	return &Handlers{store: store}
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

// FiberUpdateHandler 处理更新请求，保存加密数据
func (h *Handlers) FiberUpdateHandler(c *fiber.Ctx) error {
	var req UpdateRequest

	if err := c.BodyParser(&req); err != nil {
		logRequestError(c, "JSON 解析失败", err)
		return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: failed to parse JSON")
	}

	if req.Encrypted == "" || req.UUID == "" {
		logWarn(c, "参数缺失", "encrypted, uuid")
		return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: both 'encrypted' and 'uuid' fields are required")
	}

	if err := h.store.SaveEncryptedData(req.UUID, req.Encrypted); err != nil {
		logRequestError(c, "文件写入失败", err)
		return sendErrorResponse(c, fiber.StatusInternalServerError, "Internal Server Error: failed to save data: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"action": "done"})
}

// DecryptRequest 解密请求的数据结构
type DecryptRequest struct {
	Password string `json:"password"`
}

// FiberGetHandler 处理获取数据请求
func (h *Handlers) FiberGetHandler(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	if uuid == "" {
		logWarn(c, "参数缺失", "uuid")
		return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: 'uuid' is required")
	}

	data, err := h.store.LoadEncryptedData(uuid)
	if err != nil {
		logWarn(c, "数据不存在", uuid)
		return sendErrorResponse(c, fiber.StatusNotFound, "Not Found: data not found for uuid")
	}

	if c.Method() == "POST" {
		var req DecryptRequest
		if err := c.BodyParser(&req); err != nil {
			logRequestError(c, "JSON 解析失败", err)
			return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: failed to parse JSON")
		}

		if req.Password != "" {
			decrypted := crypto.Decrypt(uuid, data.Encrypted, req.Password)
			c.Set("Content-Type", "application/json")
			return c.Send(decrypted)
		}
	}

	return c.Status(fiber.StatusOK).JSON(data)
}

// sendErrorResponse 统一错误响应处理
func sendErrorResponse(ctx *fiber.Ctx, statusCode int, reason string) error {
	return ctx.Status(statusCode).JSON(fiber.Map{
		"action": "error",
		"reason": reason,
	})
}

// logRequestError 记录请求错误日志
func logRequestError(c *fiber.Ctx, msg string, err error) {
	log.Printf("[ERROR] %s: %v | 路径: %s | IP: %s\n", msg, err, c.Path(), c.IP())
}

// logWarn 记录警告日志
func logWarn(c *fiber.Ctx, msg, detail string) {
	log.Printf("[WARN] %s | 路径: %s | IP: %s | 缺少字段: %s\n", msg, c.Path(), c.IP(), detail)
}
