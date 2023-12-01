package logic

import (
	"context"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	"gogs.mikescher.com/BlackForestBytes/goext/syncext"
	bunny "locbunny"
	"locbunny/models"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	Config bunny.Config

	stopChan  chan bool
	Port      string
	IsRunning *syncext.AtomicBool

	Gin  *ginext.GinWrapper
	Jobs []Job
}

func NewApp() *Application {
	return &Application{
		stopChan:  make(chan bool),
		IsRunning: syncext.NewAtomicBool(false),
	}
}

func (app *Application) Init(cfg bunny.Config, g *ginext.GinWrapper, jobs []Job) {
	app.Config = cfg
	app.Gin = g
	app.Jobs = jobs
}

func (app *Application) Stop() {
	syncext.WriteNonBlocking(app.stopChan, true)
}

func (app *Application) Run() {

	addr := net.JoinHostPort(app.Config.ServerIP, app.Config.ServerPort)

	errChan, httpserver := app.Gin.ListenAndServeHTTP(addr, func(port string) {
		app.Port = port
		app.IsRunning.Set(true)
	})

	sigstop := make(chan os.Signal, 1)
	signal.Notify(sigstop, os.Interrupt, syscall.SIGTERM)

	for _, job := range app.Jobs {
		err := job.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start job")
		}
	}

	select {
	case <-sigstop:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info().Msg("Stopping HTTP-Server")

		err := httpserver.Shutdown(ctx)

		if err != nil {
			log.Info().Err(err).Msg("Error while stopping the http-server")
		} else {
			log.Info().Msg("Stopped HTTP-Server")
		}

	case err := <-errChan:
		log.Error().Err(err).Msg("HTTP-Server failed")

	case _ = <-app.stopChan:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info().Msg("Manually stopping HTTP-Server")

		err := httpserver.Shutdown(ctx)

		if err != nil {
			log.Info().Err(err).Msg("Error while stopping the http-server")
		} else {
			log.Info().Msg("Manually stopped HTTP-Server")
		}
	}

	for _, job := range app.Jobs {
		job.Stop()
	}

	app.IsRunning.Set(false)
}

func (app *Application) ListServer(ctx *ginext.AppContext) ([]models.Server, error) {

	socks, err := netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}

	res := make([]models.Server, 0)

	for _, sock := range socks {

		res = append(res, models.Server{Port: int(sock.LocalAddr.Port)})

	}

	return res, nil
}
