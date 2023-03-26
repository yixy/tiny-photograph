package resp

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yixy/tiny-photograph/common/env"
)

type Resp struct {
	Code    int
	AppName string
	Msg     string
}

const (
	RespCode = "RespCode"
	RespMsg  = "RespMsg"

	SuccessCode       = 0
	ParamCheckErr     = 10001
	RequestCheckErr   = 10002
	InternalErr       = 20001
	ServerHandleErr   = 20002
	AuthenticationErr = 20003
	AuthorizationErr  = 20004
)

var Success Resp

func init() {
	Success = Resp{
		Code:    SuccessCode,
		AppName: env.AppName,
		Msg:     "success",
	}
}

func RespSet(c *echo.Context, respCode int, respMsg string) string {
	(*c).Set(RespCode, respCode)
	(*c).Set(RespMsg, respMsg)
	return respMsg
}

func GetHttpStatus(code int) int {
	if code == SuccessCode {
		return http.StatusOK
	} else if code >= 10000 && code < 20000 {
		return http.StatusBadRequest
	} else if code >= 20000 && code < 30000 {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func RespWriter(code int, msg string) (r Resp) {
	r.AppName = env.AppName
	r.Code = code
	r.Msg = msg
	return r
}
