package main

import (
	"lottery_single/configs"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/task"
	"lottery_single/internal/service"
	"lottery_single/router"
)

func Init() {
	conf := configs.InitConfig()
	logConf := conf.LogConfig
	dbConf := conf.DbConfig
	cacheConf := conf.RedisConfig

	// 初始化日志
	log.Init(
		log.WithFileName(logConf.FileName),
		log.WithLogLevel(logConf.Level),
		log.WithLogPath(logConf.LogPath),
		log.WithMaxSize(logConf.MaxSize),
		log.WithMaxBackups(logConf.MaxBackups))

	// 初始化DB
	gormcli.Init(
		gormcli.WithAddr(dbConf.Addr),
		gormcli.WithUser(dbConf.User),
		gormcli.WithPassword(dbConf.Password),
		gormcli.WithDataBase(dbConf.DataBase),
		gormcli.WithMaxIdleConn(dbConf.MaxIdleConn),
		gormcli.WithMaxOpenConn(dbConf.MaxOpenConn),
		gormcli.WithMaxIdleTime(dbConf.MaxIdleTime))

	cache.Init(
		cache.WithAddr(cacheConf.Addr),
		cache.WithPassWord(cacheConf.PassWord),
		cache.WithDB(cacheConf.DB),
		cache.WithPoolSize(cacheConf.PoolSize))

	// 初始化各个service
	service.Init()
}

func DoTask() {
	task.DoResetIPLotteryNumsTask()
	task.DoResetUserLotteryNumsTask()
	task.DoPrizePlanTask()
}

func main() {
	Init()
	DoTask()
	router.InitRouterAndServe()
}
