package db

import (
	"log"
	"villcore.com/admin/config"

	"github.com/go-xorm/xorm"
)
import _ "github.com/go-sql-driver/mysql"

var DbEngine *xorm.Engine

func init() {
	driver := config.ServerConfig.DataSourceDriver
	url := config.ServerConfig.DataSourceUrl
	engine, err := xorm.NewEngine(driver, url)
	if err != nil {
		log.Fatalf("Create db %v engine failed %v \n", url, err)
	}
	engine.ShowSQL(true)
	engine.ShowExecTime(true)
	DbEngine = engine
}

func Close() {
	if DbEngine != nil {
		_ = DbEngine.Close()
	}
}
