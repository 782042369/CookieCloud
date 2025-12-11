package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cookiecloud/internal/crypto"
	"cookiecloud/internal/storage"

	"github.com/gorilla/mux"
)

// RootHandler 根路径处理器，返回欢迎信息
func RootHandler(apiRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World! API ROOT = %s", apiRoot)
	}
}

// UpdateRequest 更新请求的数据结构
type UpdateRequest struct {
	Encrypted string `json:"encrypted"`
	UUID      string `json:"uuid"`
}

// UpdateHandler 处理更新请求，保存加密数据
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request: failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Bad Request: failed to parse JSON", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Encrypted == "" || req.UUID == "" {
		http.Error(w, "Bad Request: both 'encrypted' and 'uuid' fields are required", http.StatusBadRequest)
		return
	}

	// 保存加密数据到文件
	err = storage.SaveEncryptedData(req.UUID, req.Encrypted)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: failed to save data: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"action": "done"})
}

// DecryptRequest 解密请求的数据结构
type DecryptRequest struct {
	Password string `json:"password"`
}

// GetHandler 处理获取数据请求
func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 验证必填字段
	if uuid == "" {
		http.Error(w, "Bad Request: 'uuid' is required", http.StatusBadRequest)
		return
	}

	// 从文件获取加密数据
	data, err := storage.LoadEncryptedData(uuid)
	if err != nil {
		http.Error(w, "Not Found: data not found for uuid", http.StatusNotFound)
		return
	}

	// 如果是POST请求且提供了密码，则解密后返回数据
	if r.Method == "POST" {
		var req DecryptRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request: failed to read request body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &req); err == nil && req.Password != "" {
			// 解密数据
			decrypted := crypto.Decrypt(uuid, data.Encrypted, req.Password)

			w.Header().Set("Content-Type", "application/json")
			w.Write(decrypted)
			return
		}
	}

	// 返回加密数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
