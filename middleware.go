package main

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

const authTokenEnvVar = "XHS_AUTH_TOKEN"

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		expectedToken := strings.TrimSpace(os.Getenv(authTokenEnvVar))
		if expectedToken == "" {
			respondError(c, http.StatusServiceUnavailable, "AUTH_NOT_CONFIGURED",
				"服务鉴权未配置", authTokenEnvVar+" is not set")
			c.Abort()
			return
		}

		providedToken, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			respondError(c, http.StatusUnauthorized, "UNAUTHORIZED",
				"未授权访问", "Authorization 请求头格式必须为 Bearer <token>")
			c.Abort()
			return
		}

		if subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
			respondError(c, http.StatusUnauthorized, "UNAUTHORIZED",
				"未授权访问", "bearer token 无效")
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractBearerToken(header string) (string, bool) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || fields[1] == "" {
		return "", false
	}

	return fields[1], true
}

// errorHandlingMiddleware 错误处理中间件
func errorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logrus.Errorf("服务器内部错误: %v, path: %s", recovered, c.Request.URL.Path)

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR",
			"服务器内部错误", recovered)
	})
}
