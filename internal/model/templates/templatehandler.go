package templates

import (
	"embed"
	"text/template"
)

type TemplateHandler struct {
	TmplHome *template.Template
}

//go:embed *
var templates embed.FS

func NewTemplateHandler() *TemplateHandler {
	mainTemplate := []string{"index.html", "header.html"}
	homeTemplate := "content.html"

	return &TemplateHandler{
		TmplHome: template.Must(template.ParseFS(templates, append(mainTemplate, homeTemplate)...)),
	}
}
