package dataconfig

import (
	"sync"

	cutils "github.com/beijian01/xgame/framework/extend/utils"
	cfacade "github.com/beijian01/xgame/framework/facade"
	clog "github.com/beijian01/xgame/framework/logger"
	cprofile "github.com/beijian01/xgame/framework/profile"
)

const (
	Name = "data_config_component"
)

type Component struct {
	sync.RWMutex
	cfacade.Component
	dataSource IDataSource
	parser     IDataParser
	configs    []IConfig
}

func New() *Component {
	return &Component{}
}

// Name unique components name
func (d *Component) Name() string {
	return Name
}

func (d *Component) Init() {
	// read data_config node in profile-{env}.json
	dataConfig := cprofile.GetConfig("data_config")
	if dataConfig.LastError() != nil {
		logrus.Fatalf("`data_config` node in `%s` file not found.", cprofile.Name())
	}

	// get data source
	sourceName := dataConfig.GetString("data_source")
	d.dataSource = GetDataSource(sourceName)
	if d.dataSource == nil {
		logrus.Fatalf("[sourceName = %s] data source not found.", sourceName)
	}

	// get parser
	parserName := dataConfig.GetString("parser")
	d.parser = GetParser(parserName)
	if d.parser == nil {
		logrus.Fatalf("[parserName = %s] parser not found.", parserName)
	}

	cutils.Try(func() {
		d.dataSource.Init(d)

		// init
		for _, cfg := range d.configs {
			cfg.Init()
		}

		// read register IConfig
		for _, cfg := range d.configs {
			data, found := d.GetBytes(cfg.Name())
			if !found {
				logrus.Warnf("[config = %s] load data fail.", cfg.Name())
				continue
			}

			cutils.Try(func() {
				d.onLoadConfig(cfg, data, false)
			}, func(errString string) {
				logrus.Errorf("[config = %s] init config error. [error = %s]", cfg.Name(), errString)
			})
		}

		// on after load
		for _, cfg := range d.configs {
			cfg.OnAfterLoad(false)
		}

		// on change process
		d.dataSource.OnChange(func(configName string, data []byte) {
			iConfig := d.GetIConfig(configName)
			if iConfig != nil {
				d.onLoadConfig(iConfig, data, true)
				iConfig.OnAfterLoad(true)
			}
		})

	}, func(errString string) {
		logrus.Error(errString)
	})
}

func (d *Component) onLoadConfig(cfg IConfig, data []byte, reload bool) {
	var parseObject interface{}
	err := d.parser.Unmarshal(data, &parseObject)
	if err != nil {
		logrus.Warnf("[config = %s] unmarshal error = %v", err, cfg.Name())
		return
	}

	d.Lock()
	defer d.Unlock()

	// load data
	size, err := cfg.OnLoad(parseObject, reload)
	if err != nil {
		logrus.Warnf("[config = %s] execute Load() error = %s", cfg.Name(), err)
		return
	}

	logrus.Infof("[config = %s] loaded. [size = %d]", cfg.Name(), size)
}

func (d *Component) OnStop() {
	if d.dataSource != nil {
		d.dataSource.Stop()
	}
}

func (d *Component) Register(configs ...IConfig) {
	if len(configs) < 1 {
		logrus.Warnf("IConfig size is less than 1.")
		return
	}

	for _, cfg := range configs {
		if cfg != nil {
			d.configs = append(d.configs, cfg)
		}
	}
}

func (d *Component) GetIConfig(name string) IConfig {
	for _, cfg := range d.configs {
		if cfg.Name() == name {
			return cfg
		}
	}
	return nil
}

func (d *Component) GetBytes(configName string) (data []byte, found bool) {
	data, err := d.dataSource.ReadBytes(configName)
	if err != nil {
		logrus.Warn(err)
		return nil, false
	}

	return data, true
}

func (d *Component) GetParser() IDataParser {
	return d.parser
}

func (d *Component) GetDataSource() IDataSource {
	return d.dataSource
}
