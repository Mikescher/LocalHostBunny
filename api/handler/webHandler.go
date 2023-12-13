package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	templhtml "html/template"
	bunny "locbunny"
	"locbunny/logic"
	"locbunny/models"
	"locbunny/webassets"
	"net/http"
	"path/filepath"
	templtext "text/template"
	"time"
)

type WebHandler struct {
	app *logic.Application
}

func NewWebHandler(app *logic.Application) WebHandler {
	return WebHandler{
		app: app,
	}
}

// ServeIndexHTML swaggerdoc
//
//	@Summary	(Website)
//
//	@Router		/ [GET]
//	@Router		/index.html [GET]
func (h WebHandler) ServeIndexHTML(pctx ginext.PreContext) ginext.HTTPResponse {
	ctx, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	templ, err := h.app.Assets.Template("index.html", h.buildIndexHTMLTemplate)
	if err != nil {
		return ginext.Error(err)
	}

	data := map[string]any{}

	bin := bytes.Buffer{}
	err = templ.Execute(&bin, data)
	if err != nil {
		return ginext.Error(err)
	}

	return ginext.Data(http.StatusOK, "text/html", bin.Bytes())
}

// ServeScriptJS swaggerdoc
//
//	@Summary	(Website)
//
//	@Router		/scripts.script.js [GET]
func (h WebHandler) ServeScriptJS(pctx ginext.PreContext) ginext.HTTPResponse {
	ctx, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	templ, err := h.app.Assets.Template("scripts/script.js", h.buildScriptJSTemplate)
	if err != nil {
		return ginext.Error(err)
	}

	data := map[string]any{}

	bin := bytes.Buffer{}
	err = templ.Execute(&bin, data)
	if err != nil {
		return ginext.Error(err)
	}

	return ginext.Data(http.StatusOK, "text/javascript", bin.Bytes())
}

func (h WebHandler) buildIndexHTMLTemplate(content []byte) (webassets.ITemplate, error) {
	t := templhtml.New("index.html")

	t.Funcs(h.templateFuncMap())

	_, err := t.Parse(string(content))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (h WebHandler) ServeAssets(pctx ginext.PreContext) ginext.HTTPResponse {
	type uri struct {
		FP1 *string `uri:"fp1"`
		FP2 *string `uri:"fp2"`
		FP3 *string `uri:"fp3"`
	}

	var u uri
	ctx, _, errResp := pctx.URI(&u).Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	assetpath := ""
	if u.FP1 == nil && u.FP2 == nil && u.FP3 == nil {
		assetpath = filepath.Join()
	} else if u.FP2 == nil && u.FP3 == nil {
		assetpath = filepath.Join(*u.FP1)
	} else if u.FP3 == nil {
		assetpath = filepath.Join(*u.FP1, *u.FP2)
	} else {
		assetpath = filepath.Join(*u.FP1, *u.FP2, *u.FP3)
	}

	data, err := h.app.Assets.Read(assetpath)
	if err != nil {
		return ginext.JSON(http.StatusNotFound, gin.H{"error": "AssetNotFound", "assetpath": assetpath})
	}

	mime := bunny.FilenameToMime(assetpath, "text/plain")

	return ginext.Data(http.StatusOK, mime, data)
}

func (h WebHandler) buildScriptJSTemplate(content []byte) (webassets.ITemplate, error) {
	t := templtext.New("scripts/script.js")

	t.Funcs(h.templateFuncMap())

	_, err := t.Parse(string(content))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (h WebHandler) templateFuncMap() map[string]any {
	return map[string]any{
		"listServers": func() []models.Server {
			ctx, cancel := context.WithTimeout(context.Background(), bunny.Conf.VerifyConnTimeoutHTML+5*time.Second)
			defer cancel()
			v, err := h.app.ListServer(ctx, bunny.Conf.VerifyConnTimeoutHTML)
			if err != nil {
				panic(err)
			}
			return v
		},
		"safe_html": func(s string) templhtml.HTML { return templhtml.HTML(s) }, //nolint:gosec
		"safe_js":   func(s string) templhtml.JS { return templhtml.JS(s) },     //nolint:gosec
		"json": func(obj any) string {
			v, err := json.Marshal(obj)
			if err != nil {
				panic(err)
			}
			return string(v)
		},
		"json_indent": func(obj any) string {
			v, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				panic(err)
			}
			return string(v)
		},
		"mkarr": func(ln int) []int { return make([]int, ln) },
	}
}
