package handle

import (
	stderr "errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/yixy/go-web-demo/common/env"
	"github.com/yixy/go-web-demo/common/resp"
	"github.com/yixy/golang-util/str"
)

const (
	sessionCookieName   = "session"
	sessValidFlag       = "isValid"
	sessUserName        = "user"
	csrfTokenCookieName = "_csrf"
	csrfTokenHeaderName = "X-XSRF-TOKEN"
	sessionCtxName      = "sess"
)

func RegistRoute(e *echo.Echo) {

	//set uuid into http header:  "X-Request-ID"(echo.HeaderXRequestID)
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "no-referrer-when-downgrade",
	}))

	//req header is X-XSRF-TOKEN
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    fmt.Sprintf("header:%s", csrfTokenHeaderName),
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieName:     csrfTokenCookieName,
	}))

	sessConfig := session.Config{
		// Skipper defines a function to skip middleware.
		Skipper: sessionSkipper,
		// default maxAge is 30 dyas.
		Store: sessions.NewFilesystemStore("", env.Secret),
		//don't use NewCookieStore, not security.
		//Store: sessions.NewCookieStore(env.Secret),
	}
	e.Use(session.MiddlewareWithConfig(sessConfig))

	//custom fuction: check and set session to context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if sessionCheckSkipper(c) {
				return next(c)
			}
			sess, err := session.Get(sessionCookieName, c)
			if err != nil {
				return errors.Wrap(err, resp.RespSet(&c, resp.InternalErr, "init session err"))
			}
			if isVlalid, ok := sess.Values[sessValidFlag].(bool); !ok || !isVlalid {
				return errors.Wrap(stderr.New("the session is invalid"), resp.RespSet(&c, resp.InternalErr, "the session is invalid"))
			}
			c.Set(sessionCtxName, sess)
			return next(c)
		}
	})
}

//signup,login not need session
func sessionSkipper(c echo.Context) bool {
	return str.Match(c.Path(), "/api/v1/users")
}

//signup,signin not need session check
func sessionCheckSkipper(c echo.Context) bool {
	return str.Match(c.Path(), "/api/v1/sessions", "/api/v1/users")
}

func getSession(c echo.Context) (*sessions.Session, error) {
	sess, ok := c.Get(sessionCtxName).(*sessions.Session)
	if !ok {
		return nil, errors.New("cannot found session in context")
	}
	return sess, nil
}
