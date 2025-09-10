package network_traffic

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"golang.org/x/time/rate"
)

type NetworkTraffic struct {
	Burst   TrafficLimitUnit
	Limit   TrafficLimitUnit
	Limiter *rate.Limiter
	io.ReadSeekCloser
}

// 流量限制单位
type TrafficLimitUnit int

const (
	// 字节
	TrafficLimitUnitByte TrafficLimitUnit = 1
	// KB
	TrafficLimitUnitKB TrafficLimitUnit = 1024
	// MB
	TrafficLimitUnitMB TrafficLimitUnit = 1024 * 1024
	// GB
	TrafficLimitUnitGB TrafficLimitUnit = 1024 * 1024 * 1024
)

// NewNetworkTraffic 创建一个网络流量限制器
func NewNetworkTraffic(limit, burst TrafficLimitUnit, src io.ReadSeekCloser) *NetworkTraffic {
	limiter := rate.NewLimiter(rate.Limit(limit), int(burst))
	return &NetworkTraffic{
		Burst:          burst,
		Limit:          limit,
		ReadSeekCloser: src,
		Limiter:        limiter,
	}
}

// Read 实现 io.ReadSeekCloser 接口
func (n *NetworkTraffic) Read(p []byte) (int, error) {
	length := len(p)
	if length == 0 {
		return 0, nil
	}

	// 判断length需要多少次token，分块读取传输
	burst := n.Limiter.Burst()

	// 确保burst不会超过缓冲区大小
	if burst > length {
		burst = length
	}

	totalRead := 0
	remaining := length

	for remaining > 0 {
		// 计算当前块大小
		chunkSize := burst
		if chunkSize > remaining {
			chunkSize = remaining
		}

		slog.Info("chunkInfo", "chunkSize", chunkSize, "chunkSizeKB", chunkSize/1024, "requestContentLength", length, "requestContentLengthKB", length/1024)

		// 为这个块预留令牌
		reservation := n.Limiter.ReserveN(time.Now(), chunkSize)
		if !reservation.OK() {
			return totalRead, errors.New("rate limit exceeded")
		}

		startTime := time.Now()
		if reservation.Delay() > 0 {
			time.Sleep(reservation.Delay())
		}
		delayTime := time.Since(startTime).Seconds()
		slog.Info("rateLimit", "chunkSize", chunkSize, "chunkSizeKB", chunkSize/1024, "delayTime", delayTime)

		// 读取数据到正确的缓冲区位置
		readLen, err := n.ReadSeekCloser.Read(p[totalRead : totalRead+chunkSize])
		if err != nil && err != io.EOF {
			return totalRead, err
		}

		totalRead += readLen
		remaining -= readLen

		// 如果读取到EOF，返回已读取的数据和EOF
		if err == io.EOF {
			return totalRead, io.EOF
		}

		// 如果没有读取到数据但没有EOF，说明可能是临时问题，结束本次读取
		if readLen == 0 {
			break
		}
	}

	return totalRead, nil
}

func (n *NetworkTraffic) Handler(dst io.Writer, Transferize int) (int64, error) {
	// 确保文件指针在开始位置
	_, err := n.ReadSeekCloser.Seek(0, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("failed to seek to start: %w", err)
	}

	ret, err := io.CopyBuffer(dst, n, make([]byte, Transferize))
	if err != nil {
		fmt.Println("===========================network traffic error:", err)
		return 0, fmt.Errorf("network traffic error: %w", err)
	}
	return ret, nil
}
