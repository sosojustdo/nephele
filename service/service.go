package service

import (
	quit "context"
	"github.com/ctripcorp/nephele/context"
	"github.com/ctripcorp/nephele/service/handler"
	"github.com/ctripcorp/nephele/service/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"time"
)

// Service Configuration.
type Config struct {
	Address        string `toml:"address"`
	BufferSize     int    `toml:"buffer-size"`
	MaxConcurrency int    `toml:"max-concurrency"`
	RequestTimeout int    `toml:"request-timeout"`
	QuitTimeout    int    `toml:"quit-timeout"`
	MiddlewarePath string `toml:"middleware-config-path"`
	Middleware     interface{}
}

// Represents http service to handle image request.
type Service struct {
	conf     Config
	image    *ImageService
	router   *gin.Engine
	internal *http.Server
	factory  *handler.HandlerFactory
}

// Returns service with given context and config.
func New(ctx *context.Context, conf Config) *Service {
	s := &Service{
		conf:    conf,
		router:  gin.New(),
		factory: handler.NewFactory(ctx),
	}
	s.image = &ImageService{internal: s}
	s.internal = &http.Server{
		Handler: s.router,
		Addr:    conf.Address,
	}
	return s
}

// Return an instance of Config with reasonable defaults.
func DefaultConfig() (Config, error) {
	return Config{
		Address:        ":8080",
		BufferSize:     200,
		RequestTimeout: 3000,
		QuitTimeout:    5000,
		MaxConcurrency: runtime.NumCPU(),
	}, nil
}

// Register http GET handler.
func (s *Service) GET(relativePath string, handlers ...handler.HandlerFunc) {
	s.router.GET(relativePath, s.factory.BuildMany(handlers...)...)
}

// Register http POST handler.
func (s *Service) POST(relativePath string, handlers ...handler.HandlerFunc) {
	s.router.POST(relativePath, s.factory.BuildMany(handlers...)...)
}

// Register http DELETE handler.
func (s *Service) DELETE(relativePath string, handlers ...handler.HandlerFunc) {
	s.router.DELETE(relativePath, s.factory.BuildMany(handlers...)...)
}

// Register http PUT handler.
func (s *Service) PUT(relativePath string, handlers ...handler.HandlerFunc) {
	s.router.PUT(relativePath, s.factory.BuildMany(handlers...)...)
}

// Register http OPTIONS handler.
func (s *Service) OPTIONS(relativePath string, handlers ...handler.HandlerFunc) {
	s.router.OPTIONS(relativePath, s.factory.BuildMany(handlers...)...)
}

// Register http HEAD handler.
func (s *Service) HEAD(relativePath string, handlers ...handler.HandlerFunc) {
	s.router.HEAD(relativePath, s.factory.BuildMany(handlers...)...)
}

// Return image service.
func (s *Service) Image() *ImageService {
	return s.image
}

// Open image http service.
func (s *Service) Open() error {
	s.router.Use(s.factory.Create())
	s.useMiddlewares()
	s.image.init().registerAll()
	return s.internal.ListenAndServe()
}

// Quit image http service gracefully.
func (s *Service) Quit() error {
	ctx, cancel := quit.WithTimeout(quit.Background(), time.Duration(s.conf.QuitTimeout)*time.Millisecond)
	defer cancel()
	return s.internal.Shutdown(ctx)
}

// Use middlewares
func (s *Service) useMiddlewares() {
	for _, m := range middleware.Build(s.conf.Middleware) {
		s.router.Use(m)
	}
}
