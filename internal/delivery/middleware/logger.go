package middleware

import (
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func Logger() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		zlog.Logger.Info().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Int("body_size", bodySize).
			Msg("HTTP Request")
	}
}
