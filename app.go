package pairec

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/middleware/prometheus"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
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
			if config.PushGatewayUseAliyunCredential {
				if cred, err := newAliyunPushGatewayCredential(); err != nil {
					log.Error(fmt.Sprintf("create prometheus push gateway credential error, err=%v", err))
				} else {
					p.SetCredential(cred)
				}
			}
			if config.PushGatewayUsername != "" {
				p.SetBasicAuth(config.PushGatewayUsername, config.PushGatewayPassword)
			}
			p.Push(config.PushGatewayURL, config.PushGatewayToken, config.PushIntervalSecs, config.Job)
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

// aliyunPushGatewayCredential adapts an Alibaba Cloud credential to the
// prometheus.BasicAuthCredential interface. The username is the AccessKeyId and
// the password is the AccessKeySecret. When the credential is an STS temporary
// credential (SecurityToken is not empty), the password follows the Alibaba Cloud
// Prometheus format: {AccessKeySecret}${SecurityToken}.
type aliyunPushGatewayCredential struct {
	cred credentials.Credential
}

func (a *aliyunPushGatewayCredential) GetBasicAuth() (username, password string, err error) {
	cm, err := a.cred.GetCredential()
	if err != nil {
		return "", "", err
	}
	// A provider may return a nil model without an error, so guard it explicitly
	// to avoid a nil pointer dereference in the push goroutine.
	if cm == nil {
		return "", "", errors.New("empty aliyun credential model")
	}

	username = tea.StringValue(cm.AccessKeyId)
	password = tea.StringValue(cm.AccessKeySecret)
	if securityToken := tea.StringValue(cm.SecurityToken); securityToken != "" {
		password = password + "$" + securityToken
	}
	return username, password, nil
}

// newAliyunPushGatewayCredential builds a push gateway credential provider backed
// by the Alibaba Cloud default credential chain (env / ECS RAM role / OIDC, etc.),
// which supports STS temporary credentials with automatic refresh.
func newAliyunPushGatewayCredential() (prometheus.BasicAuthCredential, error) {
	cred, err := credentials.NewCredential(nil)
	if err != nil {
		return nil, err
	}
	return &aliyunPushGatewayCredential{cred: cred}, nil
}
