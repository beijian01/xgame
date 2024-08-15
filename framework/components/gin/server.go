package cherryGin

import (
	"context"
	cfile "github.com/beijian01/xgame/framework/extend/file"
	cfacade "github.com/beijian01/xgame/framework/facade"
	clog "github.com/beijian01/xgame/framework/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type (
	OptionFunc func(opts *Options)

	// Options http server parameter
	Options struct {
		ReadTimeout       time.Duration
		ReadHeaderTimeout time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		MaxHeaderBytes    int
		CertFile          string
		KeyFile           string
	}

	HttpServer struct {
		cfacade.IApplication
		Options
		*gin.Engine
		server      *http.Server
		controllers []IController
	}
)

func init() {
	SetMode(gin.ReleaseMode)
}

func SetMode(value string) {
	gin.SetMode(value)
}

func NewHttpServer(address string, opts ...OptionFunc) *HttpServer {
	if address == "" {
		logrus.Error("listener address is empty.")
		return nil
	}

	httpServer := &HttpServer{
		Engine: gin.New(),
	}

	httpServer.server = &http.Server{
		Addr:    address,
		Handler: httpServer.Engine,
	}

	httpServer.server.Handler = http.AllowQuerySemicolons(httpServer.server.Handler)

	httpServer.Options = defaultOptions()
	for _, opt := range opts {
		opt(&httpServer.Options)
	}

	return httpServer
}

func defaultOptions() Options {
	return Options{
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		CertFile:          "",
		KeyFile:           "",
	}
}

func (p *HttpServer) Use(middleware ...GinHandlerFunc) {
	p.Engine.Use(BindHandlers(middleware)...)
}

func (p *HttpServer) SetIApplication(app cfacade.IApplication) {
	p.IApplication = app
}

func (p *HttpServer) Register(controllers ...IController) *HttpServer {
	p.controllers = append(p.controllers, controllers...)
	return p
}

func (p *HttpServer) Static(relativePath string, staticDir string) {
	dir, ok := cfile.JudgePath(staticDir)
	if !ok {
		logrus.Errorf("static dir path not found. staticDir = %s", staticDir)
		return
	}

	p.Engine.StaticFS(relativePath, http.Dir(dir))
}

func (p *HttpServer) StaticFile(relativePath, filepath string) {
	dir, ok := cfile.JudgeFile(filepath)
	if !ok {
		logrus.Errorf("static dir path not found. filePath = %s", filepath)
		return
	}

	p.Engine.StaticFile(relativePath, dir)
}

func (p *HttpServer) Run() {
	if p.server.Addr == "" {
		logrus.Warn("no set listener address.")
		return
	}

	if p.Options.ReadTimeout > 0 {
		p.server.ReadTimeout = p.Options.ReadTimeout
	}

	if p.Options.ReadHeaderTimeout > 0 {
		p.server.ReadHeaderTimeout = p.Options.ReadHeaderTimeout
	}

	if p.Options.WriteTimeout > 0 {
		p.server.WriteTimeout = p.Options.WriteTimeout
	}

	if p.Options.IdleTimeout > 0 {
		p.server.IdleTimeout = p.Options.IdleTimeout
	}

	if p.Options.MaxHeaderBytes > 0 {
		p.server.MaxHeaderBytes = p.Options.MaxHeaderBytes
	}

	for _, controller := range p.controllers {
		controller.PreInit(p.IApplication, p.Engine)
		controller.Init()
	}

	p.listener()
}

func (p *HttpServer) listener() {
	var err error
	if p.Options.CertFile != "" && p.Options.KeyFile != "" {
		logrus.Infof("https run. https://%s, certFile = %s, keyFile = %s",
			p.server.Addr, p.Options.CertFile, p.Options.KeyFile)
		err = p.server.ListenAndServeTLS(p.Options.CertFile, p.Options.KeyFile)
	} else {
		logrus.Infof("http run. http://%s", p.server.Addr)
		err = p.server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		logrus.Infof("run error = %s", err)
	}
}

func (p *HttpServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	for _, controller := range p.controllers {
		controller.Stop()
	}

	if err := p.server.Shutdown(ctx); err != nil {
		logrus.Info(err.Error())
	}

	logrus.Infof("shutdown http server on %s", p.server.Addr)
}

func WithReadTimeout(t time.Duration) OptionFunc {
	return func(opts *Options) {
		opts.ReadTimeout = t
	}
}

func WithReadHeaderTimeout(t time.Duration) OptionFunc {
	return func(opts *Options) {
		opts.ReadHeaderTimeout = t
	}
}

func WithIdleTimeout(t time.Duration) OptionFunc {
	return func(opts *Options) {
		opts.IdleTimeout = t
	}
}

func WithMaxHeaderBytes(val int) OptionFunc {
	return func(opts *Options) {
		opts.MaxHeaderBytes = val
	}
}

func WithCert(certFile, keyFile string) OptionFunc {
	return func(opts *Options) {
		if certFile == "" || keyFile == "" {
			return
		}
		opts.CertFile = certFile
		opts.KeyFile = keyFile
	}
}
