package handle

import (
	"context"
	stderr "errors"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/yixy/go-web-demo/common/resp"
	"github.com/yixy/go-web-demo/server/model"
)

func RegistAuthRoute(g *echo.Group) {

	//create user
	g.POST("/v1/users", func(c echo.Context) error {
		userName := c.FormValue("userName")
		hashPwd := c.FormValue("hashPwd")
		mail := c.FormValue("mail")
		localStr := c.FormValue("local_time")
		local, err := time.Parse("2006-01-02T15:04:05.999", localStr)
		if err != nil {
			return errors.Wrap(err, resp.RespSet(&c, resp.ParamCheckErr, "local time parse err"))
		}
		localTime := local.UTC().UnixMilli()
		sysTime := time.Now().UTC().UnixMilli()
		u := &model.User{
			Name:        userName,
			HashPwd:     hashPwd,
			Mail:        mail,
			Createtime:  sysTime,
			Updatetime:  sysTime,
			CreateLtime: localTime,
			UpdateLtime: localTime,
		}
		err = u.CreateUser(context.Background())
		if err != nil {
			return errors.Wrap(err, resp.RespSet(&c, resp.ServerHandleErr, "create user err"))
		}
		return c.JSON(http.StatusOK, resp.Success)
	})

	//delete user
	g.POST("/v1/users/:name", func(c echo.Context) error {
		//TODO
		return c.JSON(http.StatusOK, resp.Success)
	})

	//create session
	g.POST("/v1/sessions", func(c echo.Context) error {
		userName := c.FormValue("username")
		u := &model.User{Name: userName}
		err := u.GetUser(context.Background())
		if err != nil {
			return errors.Wrap(err, resp.RespSet(&c, resp.ServerHandleErr, "user is not exist"))
		}
		password := c.FormValue("password")

		// Throws unauthorized error
		if password != u.HashPwd {
			return errors.Wrap(stderr.New("pwd check err"), resp.RespSet(&c, resp.AuthenticationErr, "password not matched"))
		}

		sess, err := session.Get(sessionCookieName, c)
		if err != nil {
			return errors.Wrap(err, resp.RespSet(&c, resp.ServerHandleErr, "create session failed"))
		}
		sess.Values[sessUserName] = u.Name
		sess.Values[sessValidFlag] = true
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return errors.Wrap(err, resp.RespSet(&c, resp.InternalErr, "set session err"))
		}
		return c.JSON(http.StatusOK, resp.Success)
	})

	//delete session
	g.DELETE("/v1/sessions/:name", func(c echo.Context) error {
		sess, err := getSession(c)
		if err != nil {
			errors.Wrap(err, resp.RespSet(&c, resp.InternalErr, "get session err"))
		}
		if sess.Values[sessUserName] != c.Param("name") {
			return errors.Wrap(stderr.New("username not mathed"), resp.RespSet(&c, resp.AuthorizationErr, "username not mathed"))
		}
		sess.Values[sessValidFlag] = false
		sess.Options.MaxAge = -1
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			errors.Wrap(err, resp.RespSet(&c, resp.InternalErr, "set session err"))
		}
		return c.JSON(http.StatusOK, resp.Success)
	})
}
