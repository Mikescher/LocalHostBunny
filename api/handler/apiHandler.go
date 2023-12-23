package handler

import (
	"gogs.mikescher.com/BlackForestBytes/goext/exerr"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	bunny "locbunny"
	"locbunny/logic"
	"locbunny/models"
	"net/http"
)

type APIHandler struct {
	app *logic.Application
}

func NewAPIHandler(app *logic.Application) APIHandler {
	return APIHandler{
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
func (h APIHandler) ListServer(pctx ginext.PreContext) ginext.HTTPResponse {
	type response struct {
		Servers []models.Server `json:"servers"`
	}

	ctx, _, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	srvs, err := h.app.ListServer(ctx, bunny.Conf.VerifyConnTimeoutAPI)
	if err != nil {
		return ginext.Error(err)
	}

	return ginext.JSON(http.StatusOK, response{Servers: srvs})
}

// GetIcon swaggerdoc
//
//	@Summary	Get Icon
//
//	@Param		cs	path	number	true	"Icon Checksum"
//
//	@Router		/icon/:cs [GET]
func (h APIHandler) GetIcon(pctx ginext.PreContext) ginext.HTTPResponse {
	type uri struct {
		Checksum string `uri:"cs"`
	}

	var u uri
	ctx, _, errResp := pctx.URI(&u).Start()
	if errResp != nil {
		return *errResp
	}
	defer ctx.Cancel()

	icn := h.app.GetIcon(ctx, u.Checksum)
	if icn == nil {
		return ginext.Error(exerr.New(bunny.ErrEntityNotFound, "Icon not found").Str("cs", u.Checksum).WithStatuscode(404).Build())
	}

	return ginext.Data(200, icn.ContentType, icn.Data).
		WithHeader("X-BUNNY-ICONID", icn.IconID.String()).
		WithHeader("X-BUNNY-CHECKSUM", icn.Checksum).
		WithHeader("X-BUNNY-ICONDATE", icn.Time.String())
}
