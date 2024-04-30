package configs

import (
	"github.com/spf13/viper"
	"sync"
)

const defaultTimeFormat = "2006-01-02 15:04:05"

// AppConf 服务配置
type AppConf struct {
	AppName string `yaml:"app_name" mapstructure:"app_name"` // 业务名
	Version string `yaml:"version" mapstructure:"version"`   // 版本
	Port    int    `yaml:"port" mapstructure:"port"`         // 端口
	RunMode string `yaml:"run_mode" mapstructure:"run_mode"` // 运行模式
}

type LogConf struct {
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	FileName   string `mapstructure:"file_name"`
	Level      string `mapstructure:"level"`
	LogPath    string `mapstructure:"log_path"`
}

// DbConf db配置结构
type DbConf struct {
	Addr        string `yaml:"addr" mapstructure:"addr"`                   // db地址
	User        string `yaml:"user" mapstructure:"user"`                   // 用户名
	Password    string `yaml:"password" mapstructure:"password"`           // 密码
	DataBase    string `yaml:"database" mapstructure:"database"`           // db名
	MaxIdleConn int    `yaml:"max_idle_conn" mapstructure:"max_idle_conn"` // 最大空闲连接数
	MaxOpenConn int    `yaml:"max_open_conn" mapstructure:"max_open_conn"` // 最大打开的连接数
	MaxIdleTime int64  `yaml:"max_idle_time" mapstructure:"max_idle_time"` // 连接最大空闲时间
}

// RedisConf 配置
type RedisConf struct {
	Addr     string `yaml:"addr" mapstructure:"addr"`
	PassWord string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
	PoolSize int    `yaml:"poolsize" mapstructure:"poolsize"`
}

// GlobalConfig 业务配置结构体
type GlobalConfig struct {
	AppConfig   AppConf   `yaml:"app" mapstructure:"app"`
	DbConfig    DbConf    `yaml:"db" mapstructure:"db"`       // db配置
	RedisConfig RedisConf `yaml:"redis" mapstructure:"redis"` // redis配置
	LogConfig   LogConf   `yaml:"log" mapstructure:"log"`     //
}

var (
	config GlobalConfig // 全局业务配置文件
	once   sync.Once
)

// GetGlobalConfig 获取全局配置文件
func GetGlobalConfig() *GlobalConfig {
	once.Do(readConf)
	return &config
}

func readConf() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("../../../../configs")
	err := viper.ReadInConfig()
	if err != nil {
		panic("read config file err:" + err.Error())
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic("config file unmarshal err:" + err.Error())
	}
}

// InitConfig 配置初始化
func InitConfig() *GlobalConfig {
	return GetGlobalConfig()
}
