package web

import (
	"embed"
	"net/http"
	"path"
	"text/template"

	"github.com/gin-gonic/gin"
)

type HomeWeb interface {
	Index(c *gin.Context)
}

type homeWeb struct {
	embed embed.FS
}

func NewHomeWeb(embed embed.FS) *homeWeb {
	return &homeWeb{embed}
}

func (h *homeWeb) Index(c *gin.Context) {
	var filepath = path.Join("views", "main", "index.html")
	var header = path.Join("views", "general", "header.html")

	var tmpl = template.Must(template.ParseFS(h.embed, filepath, header))

	err := tmpl.Execute(c.Writer, nil)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/client/modal?status=error&message="+err.Error())
		return
	}
}
