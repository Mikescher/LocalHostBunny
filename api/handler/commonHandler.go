package handler

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	"gogs.mikescher.com/BlackForestBytes/goext/timeext"
	"locbunny/logic"
	"net/http"
	"time"
)

type CommonHandler struct {
	app *logic.Application
}

func NewCommonHandler(app *logic.Application) CommonHandler {
	return CommonHandler{
		app: app,
	}
}

type pingResponse struct {
	Message string           `json:"message"`
	Info    pingResponseInfo `json:"info"`
}
type pingResponseInfo struct {
	Method  string              `json:"method"`
	Request string              `json:"request"`
	Headers map[string][]string `json:"headers"`
	URI     string              `json:"uri"`
	Address string              `json:"addr"`
}

// Ping swaggerdoc
//
//	@Summary	Simple endpoint to test connection (any http method)
//	@Tags		Common
//
//	@Success	200	{object}	pingResponse
//	@Failure	500	{object}	models.APIError
//
//	@Router		/api/ping [get]
//	@Router		/api/ping [post]
//	@Router		/api/ping [put]
//	@Router		/api/ping [delete]
//	@Router		/api/ping [patch]
func (h CommonHandler) Ping(pctx ginext.PreContext) ginext.HTTPResponse {
	ctx, g, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(g.Request.Body)
	resuestBody := buf.String()

	return ginext.JSON(http.StatusOK, pingResponse{
		Message: "Pong",
		Info: pingResponseInfo{
			Method:  g.Request.Method,
			Request: resuestBody,
			Headers: g.Request.Header,
			URI:     g.Request.RequestURI,
			Address: g.Request.RemoteAddr,
		},
	})
}

// Health swaggerdoc
//
//	@Summary	Server Health-checks
//	@Tags		Common
//
//	@Success	200	{object}	handler.Health.response
//	@Failure	500	{object}	models.APIError
//
//	@Router		/api/health [get]
func (h CommonHandler) Health(pctx ginext.PreContext) ginext.HTTPResponse {
	type response struct {
		Status string `json:"status"`
	}

	ctx, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	return ginext.JSON(http.StatusOK, response{Status: "ok"})
}

// Sleep swaggerdoc
//
//	@Summary	Return 200 after x seconds
//	@Tags		Common
//
//	@Param		secs	path		number	true	"sleep delay (in seconds)"
//
//	@Success	200		{object}	handler.Sleep.response
//	@Failure	400		{object}	models.APIError
//	@Failure	500		{object}	models.APIError
//
//	@Router		/api/sleep/:secs [post]
func (h CommonHandler) Sleep(pctx ginext.PreContext) ginext.HTTPResponse {
	type uri struct {
		Seconds float64 `uri:"secs"`
	}
	type response struct {
		Start    string  `json:"start"`
		End      string  `json:"end"`
		Duration float64 `json:"duration"`
	}

	var u uri
	ctx, _, errResp := pctx.URI(&u).Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	t0 := time.Now().Format(time.RFC3339Nano)

	time.Sleep(timeext.FromSeconds(u.Seconds))

	t1 := time.Now().Format(time.RFC3339Nano)

	return ginext.JSON(http.StatusOK, response{
		Start:    t0,
		End:      t1,
		Duration: u.Seconds,
	})
}

func (h CommonHandler) NoRoute(pctx ginext.PreContext) ginext.HTTPResponse {

	ctx, g, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	return ginext.JSON(http.StatusNotFound, gin.H{
		"":           "================ ROUTE NOT FOUND ================",
		"FullPath":   g.FullPath(),
		"Method":     g.Request.Method,
		"URL":        g.Request.URL.String(),
		"RequestURI": g.Request.RequestURI,
		"Proto":      g.Request.Proto,
		"Header":     g.Request.Header,
		"~":          "================ ROUTE NOT FOUND ================",
	})
}
