package web

import (
	"embed"
	"text/template"
	"time"
)

//go:embed static/index.html
var indexHtml []byte

// Embed the entire directory.
//
//go:embed templates
var htmlTemplates embed.FS
var photosTmpl = template.Must(template.ParseFS(htmlTemplates, "templates/photos.html.tmpl"))

type PhotosData struct {
	Title string
	Rows  []PhotoRow
}

type PhotoRow struct {
	Timestamp time.Time
	URL       string
}
