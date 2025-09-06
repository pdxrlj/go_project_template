package http

import (
	"fmt"
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
}
