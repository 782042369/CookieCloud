package handlers

import (
	"cookiecloud/internal/crypto"
	"cookiecloud/internal/storage"
	"log"

	"github.com/gofiber/fiber/v2"
)

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
func FiberUpdateHandler(c *fiber.Ctx) error {
	var req UpdateRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] JSON 解析失败: %v | 路径: %s | IP: %s\n", err, c.Path(), c.IP())
		return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: failed to parse JSON")
	}

	// 验证必填字段
	if req.Encrypted == "" || req.UUID == "" {
		log.Printf("[WARN] 参数缺失 | 路径: %s | IP: %s | 缺少字段: encrypted, uuid\n", c.Path(), c.IP())
		return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: both 'encrypted' and 'uuid' fields are required")
	}

	// 保存加密数据到文件
	if err := storage.SaveEncryptedData(req.UUID, req.Encrypted); err != nil {
		log.Printf("[ERROR] 文件写入失败: %v | UUID: %s | IP: %s\n", err, req.UUID, c.IP())
		return sendErrorResponse(c, fiber.StatusInternalServerError, "Internal Server Error: failed to save data: "+err.Error())
	}

	// 返回成功响应
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"action": "done",
	})
}

// DecryptRequest 解密请求的数据结构
type DecryptRequest struct {
	Password string `json:"password"`
}

// FiberGetHandler 处理获取数据请求
func FiberGetHandler(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	// 验证必填字段
	if uuid == "" {
		log.Printf("[WARN] 参数缺失 | 路径: %s | IP: %s | 缺少字段: uuid\n", c.Path(), c.IP())
		return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: 'uuid' is required")
	}

	// 从文件获取加密数据
	data, err := storage.LoadEncryptedData(uuid)
	if err != nil {
		log.Printf("[WARN] 数据不存在 | UUID: %s | IP: %s\n", uuid, c.IP())
		return sendErrorResponse(c, fiber.StatusNotFound, "Not Found: data not found for uuid")
	}

	// 如果是POST请求且提供了密码，就解密后返回数据
	if c.Method() == "POST" {
		var req DecryptRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("[ERROR] JSON 解析失败: %v | 路径: %s | IP: %s\n", err, c.Path(), c.IP())
			return sendErrorResponse(c, fiber.StatusBadRequest, "Bad Request: failed to parse JSON")
		}

		if req.Password != "" {
			// 解密数据
			decrypted := crypto.Decrypt(uuid, data.Encrypted, req.Password)
			c.Set("Content-Type", "application/json")
			return c.Send(decrypted)
		}
	}

	// 返回加密数据
	return c.Status(fiber.StatusOK).JSON(data)
}

// sendErrorResponse 统一错误响应处理
func sendErrorResponse(ctx *fiber.Ctx, statusCode int, reason string) error {
	return ctx.Status(statusCode).JSON(fiber.Map{
		"action": "error",
		"reason": reason,
	})
}