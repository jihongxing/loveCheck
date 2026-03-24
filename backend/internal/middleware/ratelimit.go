package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"lovecheck/pkg/logger"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// RedisClient holds the global redis client instance.
// RedisClient 保存全局范围内可以被调用的 Redis 客户端实例。
var RedisClient *redis.Client

// InitRedis initializes the connection to the Redis container.
// InitRedis 根据我们配置的 docker-compose 环境初始化连接 Redis。
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", "lovecheck_redis_pwd"),
		DB:       0,
	})

	// Test the Redis connection
	// 验证连通性测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	logger.Log.Info().Msg("Redis connected")
}

// RateLimitMiddleware creates a Gin middleware to rate limit incoming requests using a sliding window algorithm to defend against malicious bots/crawlers.
// RateLimitMiddleware 创建一个基于 IP （或其他指纹）拦截防御恶意机器人、女巫攻击的 Gin 路由中间件。
func RateLimitMiddleware(maxReqs int64, window time.Duration) gin.HandlerFunc {
	// Derive a bucket name from maxReqs so different limiters use separate Redis keys.
	bucket := fmt.Sprintf("rl:%d:%d", maxReqs, int64(window.Seconds()))

	return func(c *gin.Context) {
		if RedisClient == nil {
			c.Next()
			return
		}

		ctx := context.Background()
		clientIP := c.ClientIP()
		key := bucket + ":" + clientIP

		pipe := RedisClient.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		val := incr.Val()
		if val > maxReqs {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate_limit_exceeded",
			})
			return
		}

		c.Next()
	}
}
