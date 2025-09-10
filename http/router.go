package http

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"telecommunications_repair_hub/pkg/network_traffic"
	"telecommunications_repair_hub/pkg/response"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type BaseRouter struct {
	*Server
}

func NewBaseRouter(e *Server) *BaseRouter {
	return &BaseRouter{
		Server: e,
	}
}

type HealthRequest struct {
	Message string `json:"message" query:"message" validate:"required,min=3,max=100"`
}

// UserRequest 用户注册请求示例，展示更多验证规则
type UserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=18,max=120"`
	Phone    string `json:"phone" validate:"required,numeric,len=11"`
	Username string `json:"username" validate:"required,alphanum,min=3,max=20"`
}

func (r *BaseRouter) RegisterRoutes() {
	r.GET("/health", func(ctx *TelecommunicationsContext, request *HealthRequest) error {
		fmt.Println(request.Message)
		return response.NewResponse(ctx.Context).Success(request)
	})

	// 用户注册示例端点
	r.POST("/register", func(ctx *TelecommunicationsContext, request *UserRequest) error {
		fmt.Printf("注册用户: %+v\n", request)
		return response.NewResponse(ctx.Context).Success(map[string]interface{}{
			"message": "用户注册成功",
			"user":    request,
		})
	})

	r.GET("/test", func(ctx *TelecommunicationsContext) error {
		return response.NewResponse(ctx.Context).Success(map[string]interface{}{
			"message": "测试成功",
		})
	})

	// prometheus
	r.GET("/metrics", func(ctx *TelecommunicationsContext) error {
		promhttp.InstrumentMetricHandler(reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          reg,
		})).ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	})

	if _, err := os.Stat("test-network-traffic"); os.IsNotExist(err) {
		// 使用dd生成100M的文件
		tempData := slices.Repeat([]byte{'a'}, 100*1024*1024)
		os.WriteFile("test-network-traffic", tempData, 0644)
	}

	fd, err := os.OpenFile("test-network-traffic", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	var OneMB = 1024 * 1024

	r.GET("/test-network-traffic", func(ctx *TelecommunicationsContext) error {
		// 获取文件大小
		fileSize, err := fd.Seek(0, io.SeekEnd)
		if err != nil {
			return response.NewResponse(ctx.Context).Error(err)
		}

		// 重置文件指针到开始位置
		_, err = fd.Seek(0, io.SeekStart)
		if err != nil {
			return response.NewResponse(ctx.Context).Error(err)
		}

		networkTraffic := network_traffic.NewNetworkTraffic(
			10*network_traffic.TrafficLimitUnitMB, // 10MB限制
			10*network_traffic.TrafficLimitUnitMB, // 10MB限制
			fd,
		)

		ctx.Response().Header().Set("Content-Length", strconv.Itoa(int(fileSize)))
		ctx.Response().Header().Set("Content-Type", "application/octet-stream")
		ctx.Response().Header().Set("Content-Disposition", "attachment; filename=\"test-file.dat\"")
		ctx.Response().Header().Set("Cache-Control", "no-cache")

		_, err = networkTraffic.Handler(ctx.Response().Writer, OneMB)
		if err != nil {
			return fmt.Errorf("network traffic error: %w", err)
		}

		// 直接返回 nil，不要调用 NoContent()，因为我们已经写入了响应体
		return nil
	})
}
