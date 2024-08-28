package web

import (
	"embed"
	"net/http"
	"path"
	"text/template"

	"github.com/gin-gonic/gin"
)

type ModalWeb interface {
	Modal(c *gin.Context)
}

type modalWeb struct {
	embed embed.FS
}

func NewModalWeb(embed embed.FS) *modalWeb {
	return &modalWeb{embed}
}

func (m *modalWeb) Modal(c *gin.Context) {
	status := c.Query("status")
	message := c.Query("message")

	var header = path.Join("views", "general", "header.html")
	var filepath = path.Join("views", "modals", "modals.html")

	var tmpl, err = template.ParseFS(m.embed, filepath, header)
	if err != nil {
		c.JSON(http.StatusSeeOther, err.Error())
		return
	}

	var dataTemplate = map[string]interface{}{
		"status":  status,
		"message": message,
	}

	err = tmpl.Execute(c.Writer, dataTemplate)
	if err != nil {
		c.JSON(http.StatusSeeOther, err.Error())
	}
}
