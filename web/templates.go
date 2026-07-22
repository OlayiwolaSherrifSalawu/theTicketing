package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
)

//go:embed templates
var templateFS embed.FS

// TemplateCache holds one fully-parsed template set per page, keyed by
// filename (e.g. "signup.tmpl"). Each set already includes the shared base
// layout and every partial, so rendering never touches the filesystem again
// after startup.
//
// Trade-off: because parsing happens once, editing a .tmpl file has no
// effect until the process restarts. Air's file watcher already restarts
// the binary on any source change, so in this project that restart is
// effectively free — but it's worth knowing this cache is not for hot
// reload without it.
type TemplateCache map[string]*template.Template

// NewTemplateCache walks templates/pages, and for each page parses it
// together with the base layout and every partial. A page template only
// needs to define "title" and "content"; partials (like "signup_form") are
// available to every page for free.
func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}

	pages, err := fs.Glob(templateFS, "templates/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := path.Base(page)

		ts, err := template.ParseFS(
			templateFS,
			"templates/base.tmpl",
			"templates/partials/*.tmpl",
			page,
		)
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", name, err)
		}

		cache[name] = ts
	}

	return cache, nil
}

// Render writes the full page (base layout + content block) for a normal
// GET request / full page load.
func (tc TemplateCache) Render(w http.ResponseWriter, status int, page string, data map[string]any) error {
	return tc.execute(w, status, page, "base", data)
}

// RenderPartial writes a single named block from a page's template set
// without the surrounding layout — this is what an htmx response swaps in.
func (tc TemplateCache) RenderPartial(w http.ResponseWriter, status int, page, block string, data map[string]any) error {
	return tc.execute(w, status, page, block, data)
}

func (tc TemplateCache) execute(w http.ResponseWriter, status int, page, block string, data map[string]any) error {
	ts, ok := tc[page]
	if !ok {
		return fmt.Errorf("template %q does not exist in cache", page)
	}

	// Render into a buffer first, not straight to w. If ExecuteTemplate
	// fails partway through (e.g. a nil map field), writing directly to w
	// would already have sent a 200 status and a half-built HTML page —
	// the client sees a broken UI with no error. Buffering lets us bail
	// out cleanly and let the caller send a proper 500 instead.
	buf := new(bytes.Buffer)
	if err := ts.ExecuteTemplate(buf, block, data); err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}
