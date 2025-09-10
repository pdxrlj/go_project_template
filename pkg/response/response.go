package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	echo.Context `json:"-"`
	Status       int    `json:"status"`
	Message      string `json:"message"`
	Data         any    `json:"data"`
}

func NewResponse(ctx echo.Context) *Response {
	return &Response{
		Context: ctx,
		Status:  http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    nil,
	}
}

func (r *Response) SetStatus(status int) *Response {
	r.Status = status
	return r
}

func (r *Response) SetMessage(message string) *Response {
	r.Message = message
	return r
}

func (r *Response) SetData(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) Success(data any) error {
	r.Status = 0
	r.Message = http.StatusText(http.StatusOK)
	r.Data = data
	return r.Context.JSON(http.StatusOK, r)
}

func (r *Response) NoContent() error {
	r.Status = 0
	r.Message = http.StatusText(http.StatusOK)
	r.Data = nil
	return r.Context.JSON(http.StatusOK, r)
}

func (r *Response) Error(data error) error {
	if r.Status == 0 {
		r.Status = http.StatusInternalServerError
	}
	r.Data = data.Error()
	return r.Context.JSON(http.StatusOK, r)
}
