package config

import (
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Name                    string
	onConfigChangeCallbacks []func(e fsnotify.Event)
}

// 初始化配置
func Init(cfgPath string) (*Config, error) {
	c := Config{
		Name: cfgPath,
	}

	// 初始化配置
	if err := c.initConfig(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) initConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name)
	} else {
		// 没有传入配置地址的话使用默认路径
		viper.AddConfigPath("config")
		viper.SetConfigName("application")
	}

	viper.SetConfigType("yaml") // 配置文件格式为yaml
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

// 初始化日志配置
func (c *Config) InitLog() {
	ljack := &lumberjack.Logger{
		Filename:   viper.GetString("log.logger_file"),
		MaxSize:    viper.GetInt("log.log_rotate_size"), // megabytes
		MaxBackups: viper.GetInt("log.log_backup_count"),
		MaxAge:     viper.GetInt("log.log_rotate_date"), //days
		Compress:   false,                               // disabled by default
	}
	mWriter := io.MultiWriter(os.Stdout, ljack)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(mWriter)
	logrus.SetLevel(logrus.DebugLevel)
}

func (c *Config) AddConfigWatch(fn func(e fsnotify.Event)) {
	c.onConfigChangeCallbacks = append(c.onConfigChangeCallbacks, fn)
}

// 热更新配置文件
func (c *Config) WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Infof("Config file changed: %s", e.Name)
		if err := viper.ReadInConfig(); err != nil {
			logrus.Error("watchConfig ReadInConfig", err)
		} else {
			for _, fn := range c.onConfigChangeCallbacks {
				fn(e)
			}
		}
	})
}
