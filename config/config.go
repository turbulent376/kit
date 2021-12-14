package config

import (
	"context"
	"git.jetbrains.space/orbi/fcsd/kit/config/configuro"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"os"
	"path/filepath"
	"reflect"
)

type Loader interface {
	WithConfigPath(path string) Loader
	WithContext(ctx context.Context) Loader
	Load(target interface{}) error
}

type loaderImpl struct {
	configPath string
	logger     log.CLoggerFunc
	ctx        context.Context
}

func (c *loaderImpl) l() log.CLogger {
	return c.logger()
}

func NewConfigLoader(logger log.CLoggerFunc) *loaderImpl {
	return &loaderImpl{
		logger: logger,
	}
}

func (c *loaderImpl) WithContext(ctx context.Context) Loader {
	c.ctx = ctx
	return c
}

func (c *loaderImpl) WithConfigPath(path string) Loader {
	c.configPath = path
	return c
}

func (c *loaderImpl) Load(target interface{}) error {
	var l log.CLogger

	if c.ctx != nil {
		l = c.l().C(c.ctx).Cmp("config-loader")
	}

	l = c.l().Cmp("config-loader")

	if reflect.ValueOf(target).Kind() != reflect.Ptr || reflect.TypeOf(target).Elem().Kind() != reflect.Struct {
		return ErrConfigTargetObjectInvalidType()
	}

	var path string
	if c.configPath != "" {
		path = c.configPath
	}

	if path == "" {
		return ErrConfigPaErrConfigPathIsEmpty()
	}

	absPath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return ErrConfigFileNotFound(absPath)
		}
		return ErrConfigFileOpen(err, absPath)
	}

	// build options
	opts := []configuro.ConfigOptions{configuro.WithLoadFromConfigFile(absPath, true)}

	// create a new config loader
	Loader, err := configuro.NewConfig(opts...)
	if err != nil {
		return ErrConfigInit(err)
	}

	// load config
	err = Loader.Load(&target)
	if err != nil {
		return ErrConfigLoad(err)
	}

	l.Dbg("config loaded")

	l.TrcObj("%v", &target)

	return nil
}
