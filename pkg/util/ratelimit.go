package util

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitedReader 带速率限制的 Reader
type RateLimitedReader struct {
	reader  io.Reader
	limiter *rate.Limiter
}

// NewRateLimitedReader 创建带速率限制的 Reader
// bytesPerSecond: 每秒字节数限制，0 表示无限制
func NewRateLimitedReader(reader io.Reader, bytesPerSecond int64) *RateLimitedReader {
	if bytesPerSecond <= 0 {
		return &RateLimitedReader{
			reader:  reader,
			limiter: nil,
		}
	}

	// 创建令牌桶限流器
	// 速率设置为 bytesPerSecond，桶容量设置为 1 秒的数据量
	limiter := rate.NewLimiter(rate.Limit(bytesPerSecond), int(bytesPerSecond))

	return &RateLimitedReader{
		reader:  reader,
		limiter: limiter,
	}
}

func (r *RateLimitedReader) Read(p []byte) (int, error) {
	if r.limiter == nil {
		// 无限制，直接读取
		return r.reader.Read(p)
	}

	n, err := r.reader.Read(p)
	if n > 0 {
		// 等待令牌
		now := time.Now()
		reservation := r.limiter.ReserveN(now, n)
		if !reservation.OK() {
			// 无法获取足够的令牌
			return n, err
		}
		delay := reservation.DelayFrom(now)
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return n, err
}

// RateLimitedWriter 带速率限制的 Writer
type RateLimitedWriter struct {
	writer  io.Writer
	limiter *rate.Limiter
}

// NewRateLimitedWriter 创建带速率限制的 Writer
// bytesPerSecond: 每秒字节数限制，0 表示无限制
func NewRateLimitedWriter(writer io.Writer, bytesPerSecond int64) *RateLimitedWriter {
	if bytesPerSecond <= 0 {
		return &RateLimitedWriter{
			writer:  writer,
			limiter: nil,
		}
	}

	limiter := rate.NewLimiter(rate.Limit(bytesPerSecond), int(bytesPerSecond))

	return &RateLimitedWriter{
		writer:  writer,
		limiter: limiter,
	}
}

func (w *RateLimitedWriter) Write(p []byte) (int, error) {
	if w.limiter == nil {
		// 无限制，直接写入
		return w.writer.Write(p)
	}

	n, err := w.writer.Write(p)
	if n > 0 {
		// 等待令牌
		now := time.Now()
		reservation := w.limiter.ReserveN(now, n)
		if !reservation.OK() {
			return n, err
		}
		delay := reservation.DelayFrom(now)
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return n, err
}

// RateLimitedReadWriter 带速率限制的 ReadWriter
type RateLimitedReadWriter struct {
	*RateLimitedReader
	*RateLimitedWriter
}

// NewRateLimitedReadWriter 创建带速率限制的 ReadWriter
func NewRateLimitedReadWriter(rw io.ReadWriter, readBytesPerSecond, writeBytesPerSecond int64) *RateLimitedReadWriter {
	return &RateLimitedReadWriter{
		RateLimitedReader: NewRateLimitedReader(rw, readBytesPerSecond),
		RateLimitedWriter: NewRateLimitedWriter(rw, writeBytesPerSecond),
	}
}

// ParseRateLimit 解析速率限制字符串
// 例如: "1M" = 1MB/s, "500K" = 500KB/s, "10" = 10B/s
func ParseRateLimit(s string) (int64, error) {
	if s == "" || s == "0" {
		return 0, nil
	}

	var value int64
	var unit string

	// 解析数值和单位
	_, err := fmt.Sscanf(s, "%d%s", &value, &unit)
	if err != nil {
		// 尝试只解析数值
		_, err = fmt.Sscanf(s, "%d", &value)
		if err != nil {
			return 0, fmt.Errorf("无效的速率限制格式: %s", s)
		}
		return value, nil
	}

	// 转换单位
	switch unit {
	case "K", "k", "KB", "kb":
		return value * 1024, nil
	case "M", "m", "MB", "mb":
		return value * 1024 * 1024, nil
	case "G", "g", "GB", "gb":
		return value * 1024 * 1024 * 1024, nil
	case "B", "b", "":
		return value, nil
	default:
		return 0, fmt.Errorf("不支持的单位: %s", unit)
	}
}
