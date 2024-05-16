package templates

import (
	"embed"
	"text/template"
)

type TemplateHandler struct {
	TmplHome   *template.Template
	TmplBlocks *template.Template
}

//go:embed *
var templates embed.FS

func NewTemplateHandler() *TemplateHandler {
	mainTemplate := []string{"index.html", "header.html"}
	homeTemplate := "content.html"
	blocksTemplate := []string{"blocks.html"}

	return &TemplateHandler{
		TmplHome:   template.Must(template.ParseFS(templates, append(mainTemplate, homeTemplate)...)),
		TmplBlocks: template.Must(template.ParseFS(templates, blocksTemplate...)),
	}
}
