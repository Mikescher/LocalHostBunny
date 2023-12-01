package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	bunny "locbunny"
	"locbunny/api"
	"locbunny/logic"
	"locbunny/webassets"
)

func main() {
	conf := bunny.Conf

	bunny.Init(conf)

	log.Info().Msg(fmt.Sprintf("Starting with config-namespace <%s>", conf.Namespace))

	assets := webassets.NewAssets()

	app := logic.NewApp(assets)

	ginengine := ginext.NewEngine(conf.Cors, conf.GinDebug, true, conf.RequestTimeout)

	router := api.NewRouter(app)

	appjobs := make([]logic.Job, 0)

	app.Init(conf, ginengine, appjobs)

	router.Init(ginengine)

	app.Run()
}
