package controllers

import (
	"coludRenderDiscovery/discovery"
	"coludRenderDiscovery/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"time"
)

type MachineController struct {
	beego.Controller
}

func (c *MachineController) Prepare() {
	c.Data["PROTOCOL_TYPE_REG_RENDER"] = discovery.PROTOCOL_TYPE_REG_RENDER
	c.Data["PROTOCOL_TYPE_NOTIFY_WEB"] = discovery.PROTOCOL_TYPE_NOTIFY_WEB
}

func (c *MachineController) Get() {
	action := c.GetString("action")
	if action == "renders" {
		c.Data["json"] = c.renders()
		c.ServeJSON()

	} else if action == "" {
		c.Layout = "manage/layout.html"
		c.LayoutSections = map[string]string{
			"Scripts": "manage/machine_scripts.html",
		}
		c.TplName = "manage/machine.html"
	}
}

func (c *MachineController) renders() []map[string]interface{} {
	o := orm.NewOrm()
	renders := make([]models.RenderMachine, 0)
	qs := o.QueryTable("RenderMachine")
	qs.OrderBy("-Id").All(&renders)
	lst := make([]map[string]interface{}, 0)
	for _, v := range renders {
		lst = append(lst, map[string]interface{}{
			"id":           v.Id,
			"name":         v.Name,
			"ip":           v.Ip,
			"group_id":     v.GroupId,
			"online":       v.Online,
			"online_time":  timeFormat(v.OnlineTime),
			"offline_time": timeFormat(v.OfflineTime),
		})
	}
	return lst
}

func timeFormat(t int64) string {
	if t == 0 {
		return ""
	}

	return time.Unix(t, 0).Format("2006-01-02 03:04:05")

}
