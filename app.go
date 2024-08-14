package pairec

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/middleware/prometheus"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

var (
	PairecApp *App
)

func init() {
	PairecApp = NewApp()
}

type App struct {
	Handlers *ControllerRegister
	Server   *http.Server
}

func NewApp() *App {
	cr := NewControllerRegister()
	app := &App{Handlers: cr, Server: &http.Server{}}

	return app
}

func (app *App) Run() {
	mode := os.Getenv("RUN_MODE")
	if mode == "COMMAND" {
		return
	}

	addr := fmt.Sprintf("%s:%d", recconf.Config.ListenConf.HttpAddr, recconf.Config.ListenConf.HttpPort)

	if recconf.Config.PrometheusConfig.Enable {
		config := recconf.Config.PrometheusConfig

		var options []prometheus.PrometheusOption

		if config.Subsystem != "" {
			options = append(options, prometheus.WithSubsystem(config.Subsystem))
		}
		if len(config.ReqDurBuckets) > 0 {
			options = append(options, prometheus.WithReqDurBuckets(config.ReqDurBuckets))
		}
		if len(config.ReqSizeBuckets) > 0 {
			options = append(options, prometheus.WithReqSzBuckets(config.ReqSizeBuckets))
		}
		if len(config.RespSizeBuckets) > 0 {
			options = append(options, prometheus.WithResSzBuckets(config.RespSizeBuckets))
		}

		p := prometheus.NewPrometheus(options...)

		if config.PushGatewayURL != "" && config.PushIntervalSecs > 0 {
			if config.Job == "" {
				env := recconf.Config.RunMode
				if os.Getenv("PAIREC_ENVIRONMENT") != "" {
					env = os.Getenv("PAIREC_ENVIRONMENT")
				}

				config.Job = env
			}
			p.Push(config.PushGatewayURL, config.PushIntervalSecs, config.Job)
		}

		app.Use(p.HandlerFunc)
	}
	app.Handlers.ApplyMiddlewares()

	app.Server.Handler = app.Handlers
	app.Server.Addr = addr
	app.Server.ReadTimeout = 30 * time.Second
	app.Server.WriteTimeout = 30 * time.Second
	app.Server.MaxHeaderBytes = 1 << 20

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("server stop, err=%v", err))
		}
	}()

	fmt.Println("server start")
	wg.Wait()
	log.Flush()
}

func (app *App) Use(middleware ...MiddlewareFunc) {
	app.Handlers.Middlewares = append(app.Handlers.Middlewares, middleware...)
}
