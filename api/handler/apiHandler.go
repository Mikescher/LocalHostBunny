package handler

import (
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
