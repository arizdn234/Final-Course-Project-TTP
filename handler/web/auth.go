package web

import (
	"a21hc3NpZ25tZW50/client"
	"a21hc3NpZ25tZW50/service"
	"embed"
	"net/http"
	"path"
	"text/template"

	"github.com/gin-gonic/gin"
)

type AuthWeb interface {
	Login(c *gin.Context)
	LoginProcess(c *gin.Context)
	Register(c *gin.Context)
	RegisterProcess(c *gin.Context)
	Logout(c *gin.Context)
}

type authWeb struct {
	userClient     client.UserClient
	sessionService service.SessionService
	embed          embed.FS
}

func NewAuthWeb(userClient client.UserClient, sessionService service.SessionService, embed embed.FS) *authWeb {
	return &authWeb{userClient, sessionService, embed}
}

func (a *authWeb) Login(c *gin.Context) {
	var filepath = path.Join("views", "auth", "login.html")
	var header = path.Join("views", "general", "header.html")

	var tmpl, err = template.ParseFS(a.embed, filepath, header)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
	}
}

func (a *authWeb) LoginProcess(c *gin.Context) {
	email := c.Request.FormValue("email")
	password := c.Request.FormValue("password")

	status, err := a.userClient.Login(email, password)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	session, err := a.sessionService.GetSessionByEmail(email)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	if status == 200 {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:   "session_token",
			Value:  session.Token,
			Path:   "/",
			MaxAge: 31536000,
			Domain: "",
		})

		c.Redirect(http.StatusSeeOther, "/client/dashboard")
	} else {
		c.Redirect(http.StatusSeeOther, "/client/login")
	}
}

func (a *authWeb) Register(c *gin.Context) {
	var header = path.Join("views", "general", "header.html")
	var filepath = path.Join("views", "auth", "register.html")

	var tmpl, err = template.ParseFS(a.embed, filepath, header)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
	}
}

func (a *authWeb) RegisterProcess(c *gin.Context) {
	fullname := c.Request.FormValue("fullname")
	email := c.Request.FormValue("email")
	password := c.Request.FormValue("password")

	status, err := a.userClient.Register(fullname, email, password)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	if status == 201 {
		c.Redirect(http.StatusSeeOther, "/client/login")
	} else {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message=Register Failed!")
	}
}

func (a *authWeb) Logout(c *gin.Context) {
	c.SetCookie("session_token", "", -1, "/", "", false, false)
	c.Redirect(http.StatusSeeOther, "/client/dashboard")
}
