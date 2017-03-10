package controllers

import (
	"coludRenderDiscovery/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type SetupController struct {
	beego.Controller
}

func (c *SetupController) Get() {
	o := orm.NewOrm()
	setup := models.RenderSetup{}
	err := o.QueryTable("RenderSetup").Limit(1).One(&setup)
	if err != nil {
		setup.MaxRenderTimeout = 6
		o.Insert(&setup)
	}

	c.Data["data"] = setup
	c.Layout = "manage/layout.html"
	c.LayoutSections = map[string]string{
		"Scripts": "manage/setup_scripts.html",
	}

	c.TplName = "manage/setup.html"
}

func (c *SetupController) Post() {
	defer c.ServeJSON()
	maxRenderTimeout, _ := c.GetInt8("maxRenderTimeout")
	if _, err := orm.NewOrm().QueryTable("RenderSetup").Update(orm.Params{
		"MaxRenderTimeout": maxRenderTimeout,
	}); err != nil {
		c.Data["json"] = false
		return
	}

	c.Data["json"] = true
}
