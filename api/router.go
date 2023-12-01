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
	apiHandler    handler.APIHandler
	webHandler    handler.WebHandler
}

func NewRouter(app *logic.Application) *Router {
	return &Router{
		app: app,

		commonHandler: handler.NewCommonHandler(app),
		apiHandler:    handler.NewAPIHandler(app),
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

	// ================ Website ================

	e.Routes().GET("/").Handle(r.webHandler.ServeIndexHTML)
	e.Routes().GET("/index.html").Handle(r.webHandler.ServeIndexHTML)
	e.Routes().GET("/scripts/script.js").Handle(r.webHandler.ServeScriptJS)
	e.Routes().GET("/:fp1").Handle(r.webHandler.ServeAssets)
	e.Routes().GET("/:fp1/:fp2").Handle(r.webHandler.ServeAssets)
	e.Routes().GET("/:fp1/:fp2/:fp3").Handle(r.webHandler.ServeAssets)

	// ================ API ================

	api.GET("/server").Handle(r.apiHandler.ListServer)

	// ================  ================

	if r.app.Config.Custom404 {
		e.NoRoute(r.commonHandler.NoRoute)
	}

}
