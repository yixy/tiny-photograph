package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yixy/tiny-photograph/common/env"
	"github.com/yixy/tiny-photograph/common/resp"
	"github.com/yixy/tiny-photograph/internal/log"
	"github.com/yixy/tiny-photograph/server/handle"
	"go.uber.org/zap"
)

const start_banner string = `
 _   _
| |_(_)_ __  _   _
| __| | '_ \| | | |
| |_| | | | | |_| |
 \__|_|_| |_|\__, |
             |___/
       _           _                              _
 _ __ | |__   ___ | |_ ___   __ _ _ __ __ _ _ __ | |__
| '_ \| '_ \ / _ \| __/ _ \ / _| | '__/ _| | '_ \| '_ \
| |_) | | | | (_) | || (_) | (_| | | | (_| | |_) | | | |
| .__/|_| |_|\___/ \__\___/ \__, |_|  \__,_| .__/|_| |_|
|_|                         |___/          |_|
`

func Start(port string, logFIle *string, isDebug *bool) {
	e := echo.New()
	log.InitLogger(*logFIle)

	e.HideBanner = true
	host, _ := os.Hostname()
	_, _ = fmt.Fprint(log.W, start_banner, fmt.Sprintf("Version %s has started at %s%s\n", env.Version, host, port))

	//print request log for debug mode
	if *isDebug {
		e.Use(middleware.Logger())
		log.Logger.Info("debug mod enable.")
	}
	e.Logger.SetOutput(log.W)

	e.Server.ReadHeaderTimeout = time.Second * 10
	e.Server.ReadTimeout = time.Second * 20
	e.Server.WriteTimeout = time.Second * 20

	e.HTTPErrorHandler = customHTTPErrorHandler
	log.Logger.Info("router init start.")

	handle.RegistRoute(e)

	g := e.Group("/api")
	//handle.RegistAuthRoute(g)
	//handle.RegistTestRoute(g)

	log.Logger.Info("router init end.")

	//wait stop signal to graceful shutdown.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		s := <-sig
		log.Logger.Info("receive a signal", zap.String("signal content", s.String()))
		err := e.Shutdown(context.Background())
		if err != nil {
			log.Logger.Error("echo.Shutdown return err", zap.Error(err))
		}
	}()

	log.Logger.Error("service shutdown", zap.Error(e.Start(port)))
}

func customHTTPErrorHandler(err error, c echo.Context) {
	//print err stack
	_, _ = fmt.Fprintf(log.W, "X-Request-ID: %s\n%+v\n", c.Response().Header().Get(echo.HeaderXRequestID), err)
	code := c.Get(resp.RespCode)
	msg := c.Get(resp.RespMsg)
	var httpStatus int
	var codei int
	var msgs string
	var ok bool
	codei, ok = code.(int)
	if !ok { //echo & Middleware err
		httpErr, ok := err.(*echo.HTTPError)
		if !ok { //unknow err
			httpStatus = http.StatusInternalServerError
			codei = resp.InternalErr
		} else { //Middleware err
			httpStatus = httpErr.Code
			msgs, _ = httpErr.Message.(string)
			if httpStatus == http.StatusForbidden || httpStatus == http.StatusBadRequest {
				//middleware.CSRFWithConfig
				codei = resp.RequestCheckErr
			} else {
				codei = resp.InternalErr
			}
		}
	} else { // handle err
		msgs, _ = msg.(string)
		httpStatus = resp.GetHttpStatus(codei)
	}
	if err := c.JSON(httpStatus, resp.RespWriter(codei, msgs)); err != nil {
		log.Logger.Error("customHTTPErrorHandler set JSON output err", zap.Error(err))
	}
}
