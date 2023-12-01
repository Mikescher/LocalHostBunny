package handler

import (
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"locbunny/logic"
	"locbunny/models"
	"net/http"
)

type WebHandler struct {
	app *logic.Application
}

func NewWebHandler(app *logic.Application) WebHandler {
	return WebHandler{
		app: app,
	}
}

// ListServer swaggerdoc
//
//	@Summary	List running server
//
//	@Success	200	{object}	handler.ListServer.response
//	@Failure	400	{object}	models.APIError
//	@Failure	500	{object}	models.APIError
//
//	@Router		/server [GET]
func (h WebHandler) ListServer(pctx ginext.PreContext) ginext.HTTPResponse {
	type response struct {
		Server []models.Server `json:"server"`
	}

	ctx, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	srvs, err := h.app.ListServer(ctx)
	if err != nil {
		return ginext.Error(err)
	}

	langext.SortBy(srvs, func(v models.Server) int { return v.Port })

	return ginext.JSON(http.StatusOK, response{Server: srvs})
}
