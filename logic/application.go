package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/adampresley/gofavigrab/parser"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/process"
	"gogs.mikescher.com/BlackForestBytes/goext/cryptext"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"gogs.mikescher.com/BlackForestBytes/goext/rext"
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
	"gogs.mikescher.com/BlackForestBytes/goext/syncext"
	"io"
	bunny "locbunny"
	"locbunny/icons"
	"locbunny/models"
	"locbunny/webassets"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var regexTitle = rext.W(regexp.MustCompile(`(?i)<title>(?P<v>[^>]+)</title>`))

type Application struct {
	Config bunny.Config

	stopChan  chan bool
	Port      string
	IsRunning *syncext.AtomicBool

	Gin    *ginext.GinWrapper
	Assets *webassets.Assets
	Jobs   []Job

	cacheLock        sync.Mutex
	serverCacheValue []models.Server
	serverCacheTime  *time.Time

	iconCache     map[string]models.Icon
	iconCacheLock sync.Mutex
}

func NewApp(ass *webassets.Assets) *Application {
	//nolint:exhaustruct
	return &Application{
		Assets:    ass,
		stopChan:  make(chan bool),
		IsRunning: syncext.NewAtomicBool(false),
		iconCache: make(map[string]models.Icon, 1024),
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

func (app *Application) ListServer(ctx context.Context, timeout time.Duration) ([]models.Server, error) {

	app.cacheLock.Lock()
	if app.serverCacheTime != nil && app.serverCacheTime.After(time.Now().Add(-bunny.Conf.CacheDuration)) {
		v := langext.ArrCopy(app.serverCacheValue)
		log.Debug().Msg(fmt.Sprintf("Return cache values (from %s)", app.serverCacheTime.Format(time.RFC3339Nano)))
		app.cacheLock.Unlock()
		return v, nil
	}
	app.cacheLock.Unlock()

	socks4, err := netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	socks6, err := netstat.TCP6Socks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
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

			con1, err := app.verifyHTTPConn(socks4[i], "HTTP", "v4", timeout)
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

			con2, err := app.verifyHTTPConn(socks4[i], "HTTPS", "v4", timeout)
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

			con1, err := app.verifyHTTPConn(socks6[i], "HTTP", "v6", timeout)
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

			con2, err := app.verifyHTTPConn(socks6[i], "HTTPS", "v6", timeout)
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

	langext.SortBy(res, func(v models.Server) int { return v.Port })

	app.cacheLock.Lock()
	app.serverCacheValue = langext.ArrCopy(res)
	app.serverCacheTime = langext.Ptr(time.Now())
	app.cacheLock.Unlock()

	return res, nil
}

func (app *Application) verifyHTTPConn(sock netstat.SockTabEntry, proto string, ipversion string, timeout time.Duration) (models.Server, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	port := int(sock.LocalAddr.Port)

	if sock.State != netstat.Listen && sock.State != netstat.Established && sock.State != netstat.TimeWait {
		log.Debug().Msg(fmt.Sprintf("Failed to verify socket [%s|%s|%d] invalid state: %s", strings.ToUpper(proto), ipversion, port, sock.State.String()))
		return models.Server{}, errors.New("invalid sock-state")
	}

	if port == bunny.Conf.ServerPort && sock.Process != nil && sock.Process.Pid == bunny.SelfProcessID {
		log.Debug().Msg(fmt.Sprintf("Skip socket [%s|%s|%d] (this is our own server)", strings.ToUpper(proto), ipversion, port))
		return models.Server{}, errors.New("skip self")
	}

	resbody, header, statuscode, err := app.doRequest(ctx, proto, port, "")
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to [%s|%s|%d] request to %d (-> %s)", strings.ToUpper(proto), ipversion, port, port, err.Error()))
		return models.Server{}, err
	}

	ct := header.Get("Content-Type")
	if ct != "" {

		var pnm *string = nil
		var pid *int = nil
		if sock.Process != nil {
			pnm = langext.Ptr(sock.Process.Name)
			pid = langext.Ptr(sock.Process.Pid)
		}

		name := app.DetectName(sock, ct, string(resbody))

		var iconRef *string = nil
		iconData, iconCT := app.DetectIcon(sock, proto, port, name, string(resbody))
		if iconData != nil && iconCT != "" {
			cs := cryptext.StrSha256(cryptext.BytesSha256(iconData) + iconCT)
			_, _ = app.StoreIcon(cs, iconData, iconCT)
			iconRef = &cs
		}

		return models.Server{
			Port:        port,
			IP:          sock.LocalAddr.IP.String(),
			Name:        name,
			Icon:        iconRef,
			Protocol:    proto,
			StatusCode:  statuscode,
			Response:    string(resbody),
			ContentType: ct,
			Process:     pnm,
			PID:         pid,
			UID:         sock.UID,
			SockState:   sock.State.String(),
		}, nil
	}

	log.Debug().Msg(fmt.Sprintf("Failed to categorize [%s|%s|%d] response (Content-Type: '%s')", strings.ToUpper(proto), ipversion, port, ct))

	return models.Server{}, errors.New("invalid response-type")
}

func (app *Application) doRequest(ctx context.Context, proto string, port int, path string) ([]byte, http.Header, int, error) {
	c := http.Client{}
	url := fmt.Sprintf("%s://localhost:%d"+path, strings.ToLower(proto), port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	resp1, err := c.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}

	defer func() { _ = resp1.Body.Close() }()

	resbody, err := io.ReadAll(resp1.Body)
	if err != nil {
		return nil, nil, 0, err
	}
	return resbody, resp1.Header, resp1.StatusCode, nil
}

func (app *Application) DetectName(sock netstat.SockTabEntry, ct string, body string) string {

	if strings.Contains(strings.ToLower(ct), "html") {
		if m, ok := regexTitle.MatchFirst(body); ok {
			title := m.GroupByName("v").Value()
			if !app.isInvalidHTMLTitle(title) {
				return title
			}
		}
	}

	if strings.Contains(strings.ToLower(body), "it looks like you are trying to access mongodb over http on the native driver port.") {
		return "MongoDB"
	}

	if sock.Process != nil {

		if sock.Process.Name == "java" {

			proc, err := process.NewProcess(int32(sock.Process.Pid))
			if err == nil {
				cmdl, err := proc.CmdlineSlice()

				if err == nil {
					if v, ok := app.extractNameFromJava(cmdl); ok {
						return v
					}
				}
			}

		}

		if len(sock.Process.Name) > 0 {
			return sock.Process.Name
		}

	}

	return "unknown"
}

func (app *Application) DetectIcon(sock netstat.SockTabEntry, proto string, port int, name string, body string) ([]byte, string) {

	if strings.Contains(strings.ToLower(body), "it looks like you are trying to access mongodb over http on the native driver port.") {
		return icons.MongoDB, "image/svg+xml"
	}

	if sock.Process != nil {
		pname := strings.ToLower(strings.TrimSpace(sock.Process.Name))
		if pname == "vlc" {
			return icons.VLC, "image/svg+xml"
		}
		if pname == "cupsd" {
			return icons.CUPS, "image/svg+xml"
		}
		if pname == "containerd" {
			return icons.Docker, "image/svg+xml"
		}
	}

	name = strings.ToLower(name)

	if strings.HasPrefix(name, "goland") {
		return icons.GoLand, "image/svg+xml"
	}
	if strings.HasPrefix(name, "phpstorm") {
		return icons.PHPStorm, "image/svg+xml"
	}
	if strings.HasPrefix(name, "pycharm") {
		return icons.PyCharm, "image/svg+xml"
	}
	if strings.HasPrefix(name, "webstorm") {
		return icons.WebStorm, "image/svg+xml"
	}
	if strings.HasPrefix(name, "intellijidea") {
		return icons.IntellijIDEA, "image/svg+xml"
	}
	if strings.HasPrefix(name, "rider") {
		return icons.Rider, "image/svg+xml"
	}
	if strings.HasPrefix(name, "androidstudio") {
		return icons.AndroidStudio, "image/svg+xml"
	}

	if favurlAbs, err := parser.NewHTMLParser(body).GetFaviconURL(); err == nil {
		if parsedURL, err := url.Parse(favurlAbs); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			resbody, hdr, sc, err := app.doRequest(ctx, proto, port, parsedURL.EscapedPath())
			if err == nil && sc >= 200 && sc < 300 && hdr.Get("Content-Type") != "" {
				return resbody, hdr.Get("Content-Type")
			}
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		resbody, _, sc, err := app.doRequest(ctx, proto, port, "/favicon.ico")
		if err == nil && sc >= 200 && sc < 300 {
			return resbody, "image/vnd.microsoft.icon"
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		resbody, _, sc, err := app.doRequest(ctx, proto, port, "/favicon.png")
		if err == nil && sc >= 200 && sc < 300 {
			return resbody, "image/png"
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		resbody, _, sc, err := app.doRequest(ctx, proto, port, "/favicon.jpeg")
		if err == nil && sc >= 200 && sc < 300 {
			return resbody, "image/jpeg"
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		resbody, _, sc, err := app.doRequest(ctx, proto, port, "/favicon.jpg")
		if err == nil && sc >= 200 && sc < 300 {
			return resbody, "image/jpeg"
		}
	}

	if sock.Process != nil && sock.Process.Name == "java" {
		return icons.Java, "image/svg+xml"
	}

	return nil, ""
}

func (app *Application) isInvalidHTMLTitle(title string) bool {
	title = strings.ToLower(title)
	title = strings.TrimSpace(title)
	title = strings.Trim(title, ".,\r\n\t ;")

	arr := []string{
		"404",
		"Not found",
		"404 Not Found",
		"404 - Not Found",
		"Page Not Found",
		"File Not Found",
		"Not Found",
		"Site Not Found",
		"ISAPI or CGI restriction",
		"MIME type restriction",
		"No handler configured",
		"Denied by request filtering configuration",
		"Verb denied",
		"File extension denied",
		"Hidden namespace",
		"File attribute hidden",
		"Request header too long",
		"Request contains double escape sequence",
		"Request contains high-bit characters",
		"Content length too large",
		"Request URL too long",
		"Query string too long",
		"DAV request sent to the static file handler",
		"Dynamic content mapped to the static file handler via a wildcard MIME mapping",
		"Query string sequence denied",
		"Denied by filtering rule",
		"Too Many URL Segments",
	}

	for _, v := range arr {
		if title == strings.ToLower(v) {
			return true
		}
	}

	return false
}

func (app *Application) extractNameFromJava(cmdl []string) (string, bool) {

	for i, v := range cmdl {
		if strings.ToLower(v) == "-jar" && i+1 < len(cmdl) {
			return cmdl[i+1], true
		}
	}

	for _, v := range cmdl {
		if strings.HasPrefix(strings.ToLower(v), "-didea.paths.selector=") {
			return v[len("-Didea.paths.selector="):], true
		}
	}

	for _, v := range cmdl {
		if strings.HasPrefix(strings.ToLower(v), "-didea.platform.prefix") {
			return v[len("-Didea.platform.prefix"):], true
		}
	}

	if len(cmdl) > 0 {
		return cmdl[len(cmdl)-1], true
	}

	return "", false
}

func (app *Application) StoreIcon(cs string, data []byte, ct string) (models.Icon, bool) {
	app.iconCacheLock.Lock()
	defer app.iconCacheLock.Unlock()

	if v, ok := app.iconCache[cs]; ok {
		return v, false
	}

	v := models.Icon{
		IconID:      models.NewIconID(),
		Checksum:    cs,
		Data:        data,
		ContentType: ct,
		Time:        rfctime.NowRFC3339Nano(),
	}

	app.iconCache[cs] = v

	return v, true
}

func (app *Application) GetIcon(ctx *ginext.AppContext, cs string) *models.Icon {
	app.iconCacheLock.Lock()
	defer app.iconCacheLock.Unlock()

	if v, ok := app.iconCache[cs]; ok {
		return langext.Ptr(v)
	}

	return nil
}
