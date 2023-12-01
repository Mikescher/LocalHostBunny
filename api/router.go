package api

import (
	"fmt"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	bunny "locbunny"
	"locbunny/api/handler"
	"locbunny/logic"
	"locbunny/swagger"
)

type Router struct {
	app *logic.Application

	commonHandler handler.CommonHandler
	webHandler    handler.WebHandler
}

func NewRouter(app *logic.Application) *Router {
	return &Router{
		app: app,

		commonHandler: handler.NewCommonHandler(app),
		webHandler:    handler.NewWebHandler(app),
	}
}

// Init swaggerdocs
//
//	@title		LocalHostBunny
//	@version	1.0
//	@host		localhost
//
//	@BasePath	/api/v1/
func (r *Router) Init(e *ginext.GinWrapper) {

	api := e.Routes().Group("/api").Group(fmt.Sprintf("/v%d", bunny.APILevel))

	// ================ General ================

	api.Any("/ping").Handle(r.commonHandler.Ping)
	api.GET("/health").Handle(r.commonHandler.Health)
	api.POST("/sleep/:secs").Handle(r.commonHandler.Sleep)

	// ================ Swagger ================

	docs := e.Routes().Group("/documentation")
	{
		docs.GET("/swagger").Handle(ginext.RedirectTemporary("/documentation/swagger/"))
		docs.GET("/swagger/*sub").Handle(swagger.Handle)
	}

	// ================ API ================

	api.GET("/server").Handle(r.webHandler.ListServer)

	// ================  ================

	if r.app.Config.Custom404 {
		e.NoRoute(r.commonHandler.NoRoute)
	}

}
