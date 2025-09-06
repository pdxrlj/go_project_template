package http

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"telecommunications_repair_hub/config"
	"telecommunications_repair_hub/pkg"
	"telecommunications_repair_hub/pkg/db"
	"telecommunications_repair_hub/pkg/response"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/go-playground/validator/v10"
)

type Server struct {
	*echo.Echo
	config                *config.Config
	db                    *db.DB
	globalMiddlewares     map[string]echo.MiddlewareFunc
	globalMiddlewaresName string
}

type Validator struct {
	validator *validator.Validate
}

type ValidationErrors struct {
	Errors []validator.FieldError `json:"errors"`
}

func (v *ValidationErrors) Error() string {
	errors := []string{}
	for _, error := range v.Errors {
		errors = append(errors, error.Error())
	}
	return strings.Join(errors, ", ")
}

func (v *Validator) Validate(i interface{}) error {
	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}
	// 注释掉调试输出，如果需要可以使用正式的日志记录
	// utils.PP("err", err)
	// 处理验证错误
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		return &ValidationErrors{
			Errors: validatorErrors,
		}
	}

	return err
}

func NewServer(config *config.Config) *Server {
	dbInstance, err := db.New(config)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Validator = &Validator{
		validator: validator.New(),
	}
	e.HideBanner = true
	e.HidePort = true
	s := &Server{
		Echo:              e,
		config:            config,
		db:                dbInstance,
		globalMiddlewares: make(map[string]echo.MiddlewareFunc),
	}

	s.UseGlobalMiddleware()

	return s
}

func (s *Server) UseGlobalMiddleware() {
	useMiddlewares := map[string]echo.MiddlewareFunc{
		"cors":      middleware.CORS(),
		"bodyLimit": middleware.BodyLimit("5M"),
		"secure":    middleware.Secure(),
		"recover": middleware.RecoverWithConfig(middleware.RecoverConfig{
			DisableStackAll:   true,
			DisablePrintStack: true,
		}),
		"logger": middleware.LoggerWithConfig(middleware.LoggerConfig{
			Output:           os.Stdout,
			Format:           fmt.Sprintf(`%s time":"%s","method":"%s","uri":"%s","status":%s,"latency_human":"%s","bytes_in":%s,"bytes_out":%s}`, color.BlueString("[TelecommunicationsServer]"), color.GreenString("${time_custom}"), color.GreenString("${method}"), color.GreenString("${uri}"), color.GreenString("${status}"), color.GreenString("${latency_human}"), color.GreenString("${bytes_in}"), color.GreenString("${bytes_out}")) + "\n",
			CustomTimeFormat: "2006-01-02 15:04:05",
		}),
		"requestCounter": RequestCounterMiddleware,
	}	
	userMiddlewaresName := ""
	for name, middleware := range useMiddlewares {
		s.globalMiddlewares[name] = middleware
		userMiddlewaresName += name + ","
		s.Echo.Use(middleware)
	}
	userMiddlewaresName = userMiddlewaresName[:len(userMiddlewaresName)-1]
	s.globalMiddlewaresName = userMiddlewaresName
}

type TelecommunicationsContext struct {
	echo.Context
	DBInstance *db.DB
}

type HttpHandler func(ctx *TelecommunicationsContext, request any) error

func (s *Server) GET(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodGet, path, handler, middlewares...)
}

func (s *Server) POST(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodPost, path, handler, middlewares...)
}

func (s *Server) PUT(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodPut, path, handler, middlewares...)
}

func (s *Server) PATCH(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodPatch, path, handler, middlewares...)
}

func (s *Server) OPTIONS(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodOptions, path, handler, middlewares...)
}

func (s *Server) HEAD(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodHead, path, handler, middlewares...)
}

func (s *Server) DELETE(path string, handler any, middlewares ...echo.MiddlewareFunc) {
	s.Add(http.MethodDelete, path, handler, middlewares...)
}

// 终止如果不想等
// condition 是否终止
func (s *Server) Terminate(condition bool, message string) {
	if condition {
		panic(message)
	}
}

func (s *Server) RequestValidator(ctx echo.Context, requestType any, request reflect.Type) error {
	if s.Validator != nil {
		if err := ctx.Validate(requestType); err != nil {
			if validationErrors, ok := err.(*ValidationErrors); ok {
				firstError := validationErrors.Errors[0]
				fieldType := firstError.StructField()
				actualTag := firstError.ActualTag()

				validationFieldTag, validationRule := getValidationFieldTag(request, fieldType, actualTag)

				err = fmt.Errorf("参数 %s 无法通过 %s 规则的验证", validationFieldTag, validationRule)
				response.NewResponse(ctx).SetStatus(pkg.GetTeleCommunicationErrorCode(pkg.ErrParamError)).
					SetMessage(pkg.ErrParamError.Error()).
					Error(err)
				return err
			}

			response.NewResponse(ctx).
				SetStatus(pkg.GetTeleCommunicationErrorCode(pkg.ErrParamError)).
				SetMessage(pkg.ErrParamError.Error()).Error(err)
			return err
		}
	}
	return nil
}

func (s *Server) ResoverHandler(ctx echo.Context, handlerValue reflect.Value, requests ...any) error {
	context := &TelecommunicationsContext{
		Context:    ctx,
		DBInstance: s.db,
	}

	in := []reflect.Value{
		reflect.ValueOf(context),
	}
	if len(requests) > 0 {
		in = append(in, reflect.ValueOf(requests[0]))
	}
	result := handlerValue.Call(in)[0]
	if result.IsNil() {
		return nil
	}
	respError := result.Interface().(error)
	if respError != nil {
		slog.Error("API返回数据异常", "error", respError)
	}

	return nil
}

func (s *Server) Add(method string, path string, handler any, middlewares ...echo.MiddlewareFunc) {
	handlerValue := reflect.ValueOf(handler)
	s.Terminate(handlerValue.Kind() != reflect.Func, "处理函数必须是一个函数")

	handlerType := handlerValue.Type()

	s.Terminate(handlerType.NumIn() < 1, "处理函数必须至少有一个参数")
	s.Terminate(handlerType.In(0) != reflect.TypeOf(&TelecommunicationsContext{}), "处理函数第一个参数必须是TelecommunicationsContext")
	s.Terminate(handlerType.NumOut() != 1 || handlerType.Out(0) != reflect.TypeOf((*error)(nil)).Elem(), "处理函数必须返回一个error")

	inputNumber := handlerType.NumIn()
	userMiddlewaresName := s.globalMiddlewaresName

	if len(middlewares) > 0 {
		userMiddlewaresName = userMiddlewaresName + ","
		for _, middleware := range middlewares {
			funcName := getFuncName(middleware)
			userMiddlewaresName += funcName + ","
		}

		userMiddlewaresName = userMiddlewaresName[:len(userMiddlewaresName)-1]
	}

	handlerName := handlerType.String()
	addTerminalTable(s.config.App.Port, method, path,
		handlerName, userMiddlewaresName)

	s.Echo.Add(method, path, func(ctx echo.Context) error {
		if inputNumber == 1 {
			return s.ResoverHandler(ctx, handlerValue)
		}
		request := handlerType.In(1)
		if request.Kind() == reflect.Ptr {
			request = request.Elem()
		}

		requestType := reflect.New(request).Interface()

		if err := ctx.Bind(requestType); err != nil {
			err := response.NewResponse(ctx).
				SetStatus(pkg.GetTeleCommunicationErrorCode(pkg.ErrParamError)).
				SetMessage(pkg.ErrParamError.Error()).
				Error(err)

			return err
		}

		if err := s.RequestValidator(ctx, requestType, request); err != nil {
			return err
		}

		return s.ResoverHandler(ctx, handlerValue, requestType)
	}, middlewares...)
}

func getValidationFieldTag(structType reflect.Type, defaultTag string, actualTag string) (string, string) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	structField, _ := structType.FieldByName(defaultTag)

	validationFieldTag := ""
	ok := false
	validationFieldTag, ok = structField.Tag.Lookup("query")
	if !ok {
		validationFieldTag, ok = structField.Tag.Lookup("json")
	}
	if !ok {
		validationFieldTag, ok = structField.Tag.Lookup("path")
	}
	if !ok {
		validationFieldTag, ok = structField.Tag.Lookup("form")
	}
	if !ok {
		validationFieldTag, ok = structField.Tag.Lookup("header")
	}
	if !ok {
		validationFieldTag, ok = structField.Tag.Lookup("param")
	}

	if !ok {
		validationFieldTag = defaultTag
	}

	// 获取验证规则
	validationRule := structField.Tag.Get("validate")
	// 获取validationRule中的 actualTag
	validationRules := strings.Split(validationRule, ",")
	for _, rule := range validationRules {
		if strings.HasPrefix(rule, actualTag) {
			validationRule = rule
			break
		}
	}

	return validationFieldTag, validationRule
}

func getFuncName(middleware echo.MiddlewareFunc) string {
	funcName := runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name()
	funcNames := strings.Split(funcName, ".")
	funcName = funcNames[len(funcNames)-1]
	if funcName == "main" {
		funcName = "echo.MiddlewareFunc"
	}
	return funcName
}

func addTerminalTable(port string, method string, path string, handler string, middleware string) {
	userMiddlewaresName := middleware
	tableRouter.AppendRow(table.Row{
		port,
		method,
		path,
		handler,
		userMiddlewaresName,
	})
}

var tableRouter = table.NewWriter()

func initTerminalTable() {
	tableRouter.SetOutputMirror(os.Stdout)

	tableRouter.SetStyle(table.StyleDefault)

	tableRouter.Style().Options.SeparateRows = true
	tableRouter.Style().Options.SeparateColumns = true

	tableRouter.SetTitle(color.CyanString("Api 接口路由表"))
	tableRouter.AppendHeader(table.Row{
		color.BlueString("端口"),
		color.BlueString("方法"),
		color.BlueString("路径"),
		color.BlueString("处理函数"),
		color.BlueString("中间件"),
	})
	tableRouter.AppendSeparator()
	tableRouter.SetCaption("Telecommunications Server Routes")
}

func (s *Server) Start(host string, port string) error {
	initTerminalTable()
	tableRouter.Render()
	fmt.Println()

	address := net.JoinHostPort(host, port)

	err := s.Echo.Start(address)
	if err != nil {
		return err
	}
	return nil
}
