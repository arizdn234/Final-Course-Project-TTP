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

type DashboardWeb interface {
	Dashboard(c *gin.Context)
}

type dashboardWeb struct {
	userClient     client.UserClient
	sessionService service.SessionService
	embed          embed.FS
}

func NewDashboardWeb(userClient client.UserClient, sessionService service.SessionService, embed embed.FS) *dashboardWeb {
	return &dashboardWeb{userClient, sessionService, embed}
}

func (d *dashboardWeb) Dashboard(c *gin.Context) {
	var email string
	if temp, ok := c.Get("email"); ok {
		if contextData, ok := temp.(string); ok {
			email = contextData
		}
	}

	session, err := d.sessionService.GetSessionByEmail(email)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	userTaskCategories, err := d.userClient.GetUserTaskCategory(session.Token)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	var dataTemplate = map[string]interface{}{
		"email":                email,
		"user_task_categories": userTaskCategories,
	}

	var funcMap = template.FuncMap{
		"exampleFunc": func() int {
			return 0
		},
	}

	var header = path.Join("views", "general", "header.html")
	var filepath = path.Join("views", "main", "dashboard.html")

	t, err := template.New("dashboard.html").Funcs(funcMap).ParseFS(d.embed, filepath, header)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}

	err = t.Execute(c.Writer, dataTemplate)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
	}
}
