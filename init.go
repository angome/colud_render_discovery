package main

import (
	"coludRenderDiscovery/models"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func init() {
	os.MkdirAll(beego.AppConfig.String("upfile::SaveDir"), 0755)
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s",
		beego.AppConfig.String("mysql::User"),
		beego.AppConfig.String("mysql::Pwd"),
		beego.AppConfig.String("mysql::Host"),
		beego.AppConfig.String("mysql::Port"),
		beego.AppConfig.String("mysql::DbName"),
		"Asia%2FShanghai",
	)

	orm.RegisterDataBase("default", "mysql", connStr, 3, 35)
	orm.RegisterModel(
		new(models.RenderTask),
		new(models.RenderMachine),
		new(models.RenderTaskUsage),
		new(models.RenderSetup),
	)

	if beego.AppConfig.String("default::RunMode") == "dev" {
		//orm.Debug = true
	}

	orm.RunSyncdb("default", false, true)

	beego.SetStaticPath("/assets", "static/assets")
}


