package templates

import (
	"embed"
	"text/template"
)

type TemplateHandler struct {
	TmplHome   *template.Template
	TmplPlayer *template.Template
}

//go:embed *
var templates embed.FS

func NewTemplateHandler() *TemplateHandler {
	mainTemplate := []string{"index.html", "header.html"}
	homeTemplate := "content.html"
	playerTemplate := []string{"player.html"}

	return &TemplateHandler{
		TmplHome:   template.Must(template.ParseFS(templates, append(mainTemplate, homeTemplate)...)),
		TmplPlayer: template.Must(template.ParseFS(templates, playerTemplate...)),
	}
}
