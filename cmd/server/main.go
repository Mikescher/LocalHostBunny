package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	bunny "locbunny"
	"locbunny/api"
	"locbunny/logic"
)

func main() {
	conf := bunny.Conf

	bunny.Init(conf)

	log.Info().Msg(fmt.Sprintf("Starting with config-namespace <%s>", conf.Namespace))

	app := logic.NewApp()

	ginengine := ginext.NewEngine(conf.Cors, conf.GinDebug, true, conf.RequestTimeout)

	router := api.NewRouter(app)

	appjobs := make([]logic.Job, 0)

	app.Init(conf, ginengine, appjobs)

	router.Init(ginengine)

	app.Run()
}
