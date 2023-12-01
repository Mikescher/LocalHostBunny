package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"gogs.mikescher.com/BlackForestBytes/goext/syncext"
	"io"
	bunny "locbunny"
	"locbunny/models"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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

	addr := net.JoinHostPort(app.Config.ServerIP, strconv.Itoa(app.Config.ServerPort))

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

	socks4, err := netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}

	socks6, err := netstat.TCP6Socks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}

	sockCount := len(socks4) + len(socks6)

	wg := sync.WaitGroup{}

	echan := make(chan error, sockCount*3)
	rchan := make(chan models.Server, sockCount*3)

	for _i := range socks4 {
		i := _i
		wg.Add(1)
		go func() {
			defer wg.Done()

			con1, err := app.verifyHTTPConn(socks4[i], "HTTP", "v4")
			if err == nil {
				rchan <- con1
				return
			} else {
				echan <- err
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			con2, err := app.verifyHTTPConn(socks4[i], "HTTPS", "v4")
			if err == nil {
				rchan <- con2
				return
			} else {
				echan <- err
			}
		}()
	}

	for _i := range socks6 {
		i := _i
		wg.Add(1)
		go func() {
			defer wg.Done()

			con1, err := app.verifyHTTPConn(socks6[i], "HTTP", "v6")
			if err == nil {
				rchan <- con1
				return
			} else {
				echan <- err
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			con2, err := app.verifyHTTPConn(socks6[i], "HTTPS", "v6")
			if err == nil {
				rchan <- con2
				return
			} else {
				echan <- err
			}
		}()
	}

	wg.Wait()
	close(echan)
	close(rchan)

	duplicates := make(map[int]bool, sockCount*3)
	res := make([]models.Server, 0, sockCount*3)
	for v := range rchan {

		if _, ok := duplicates[v.Port]; !ok {
			res = append(res, v)
			duplicates[v.Port] = true
		}
	}

	return res, nil
}

func (app *Application) verifyHTTPConn(sock netstat.SockTabEntry, proto string, ipversion string) (models.Server, error) {

	ctx, cancel := context.WithTimeout(context.Background(), bunny.Conf.VerifyConnTimeout)
	defer cancel()

	if sock.State != netstat.Listen {
		log.Debug().Msg(fmt.Sprintf("Failed to verify socket [%s|%s] invalid state: %s", ipversion, strings.ToUpper(proto), sock.State.String()))
		return models.Server{}, errors.New("invalid sock-state")
	}

	if int(sock.LocalAddr.Port) == bunny.Conf.ServerPort && sock.Process != nil && sock.Process.Pid == bunny.SelfProcessID {
		log.Debug().Msg(fmt.Sprintf("Skip socket [%s|%s] (this is our own server)", ipversion, strings.ToUpper(proto)))
		return models.Server{}, errors.New("skip self")
	}

	c := http.Client{}
	url := fmt.Sprintf("%s://localhost:%d", strings.ToLower(proto), sock.LocalAddr.Port)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to create [%s|%s] request to %d", ipversion, strings.ToUpper(proto), sock.LocalAddr.Port))
		return models.Server{}, err
	}

	resp1, err := c.Do(req)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to send [%s|%s] request to %s", ipversion, strings.ToUpper(proto), url))
		return models.Server{}, err
	}

	defer func() { _ = resp1.Body.Close() }()

	resbody, err := io.ReadAll(resp1.Body)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to read [%s|%s] response from %s", ipversion, strings.ToUpper(proto), url))
		return models.Server{}, err
	}

	ct := resp1.Header.Get("Content-Type")
	if ct != "" {

		var pnm *string = nil
		var pid *int = nil
		if sock.Process != nil {
			pnm = langext.Ptr(sock.Process.Name)
			pid = langext.Ptr(sock.Process.Pid)
		}

		return models.Server{
			Port:        int(sock.LocalAddr.Port),
			IP:          sock.LocalAddr.IP.String(),
			Protocol:    proto,
			StatusCode:  resp1.StatusCode,
			Response:    string(resbody),
			ContentType: ct,
			Process:     pnm,
			PID:         pid,
			UID:         sock.UID,
			SockState:   sock.State.String(),
		}, nil
	}

	log.Debug().Msg(fmt.Sprintf("Failed to categorize [%s|%s] response from %s (Content-Type: '%s')", ipversion, strings.ToUpper(proto), url, ct))
	return models.Server{}, errors.New("invalid response-type")
}
