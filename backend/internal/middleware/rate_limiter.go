package middleware

import (
	"electronic-muyu-backend/internal/services"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterMiddleware 包装限流服务
type RateLimiterMiddleware struct {
	rateLimiter *services.RateLimiterService
}

// NewRateLimiterMiddleware 创建限流中间件实例
func NewRateLimiterMiddleware(rateLimiter *services.RateLimiterService) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiter: rateLimiter,
	}
}

// AuthRateLimit 返回认证相关的限流中间件
func (m *RateLimiterMiddleware) AuthRateLimit() gin.HandlerFunc {
	if m.rateLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return m.createRateLimiter(services.LoginRateLimit)
}

// ScoreSubmitRateLimit 返回分数提交的限流中间件
func (m *RateLimiterMiddleware) ScoreSubmitRateLimit() gin.HandlerFunc {
	if m.rateLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return m.createUserBasedRateLimiter(services.ScoreSubmitRateLimit)
}

// APIRateLimit 返回通用API的限流中间件
func (m *RateLimiterMiddleware) APIRateLimit() gin.HandlerFunc {
	if m.rateLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return m.createRateLimiter(services.GeneralAPIRateLimit)
}

// createRateLimiter 创建基于IP的限流中间件
func (m *RateLimiterMiddleware) createRateLimiter(rateLimit services.RateLimit) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP作为标识符
		clientIP := getClientIP(c)

		// 检查是否允许请求
		allowed, err := m.rateLimiter.IsAllowed(clientIP, rateLimit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiter service error",
			})
			c.Abort()
			return
		}

		if !allowed {
			// 获取剩余请求数用于响应头
			remaining, _ := m.rateLimiter.GetRemainingRequests(clientIP, rateLimit)

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit.Limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rateLimit.Window).Unix()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded",
				"message":   "Too many requests, please try again later",
				"limit":     rateLimit.Limit,
				"window":    rateLimit.Window.String(),
				"remaining": remaining,
			})
			c.Abort()
			return
		}

		// 添加限流信息到响应头
		remaining, _ := m.rateLimiter.GetRemainingRequests(clientIP, rateLimit)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rateLimit.Window).Unix()))

		c.Next()
	}
}

// createUserBasedRateLimiter 创建基于用户ID的限流中间件
func (m *RateLimiterMiddleware) createUserBasedRateLimiter(rateLimit services.RateLimit) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从JWT中获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			c.Abort()
			return
		}

		identifier := fmt.Sprintf("user_%v", userID)

		// 检查是否允许请求
		allowed, err := m.rateLimiter.IsAllowed(identifier, rateLimit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiter service error",
			})
			c.Abort()
			return
		}

		if !allowed {
			remaining, _ := m.rateLimiter.GetRemainingRequests(identifier, rateLimit)

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit.Limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rateLimit.Window).Unix()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded",
				"message":   "Too many requests, please try again later",
				"limit":     rateLimit.Limit,
				"window":    rateLimit.Window.String(),
				"remaining": remaining,
			})
			c.Abort()
			return
		}

		// 添加限流信息到响应头
		remaining, _ := m.rateLimiter.GetRemainingRequests(identifier, rateLimit)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rateLimit.Window).Unix()))

		c.Next()
	}
}

// getClientIP 获取客户端真实IP地址
func getClientIP(c *gin.Context) string {
	// 检查 X-Forwarded-For 头
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 检查 X-Real-IP 头
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		if net.ParseIP(xRealIP) != nil {
			return xRealIP
		}
	}

	// 检查 X-Forwarded-Host 头
	xForwardedHost := c.GetHeader("X-Forwarded-Host")
	if xForwardedHost != "" {
		if net.ParseIP(xForwardedHost) != nil {
			return xForwardedHost
		}
	}

	// 使用 RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}
