package controllers

import (
	"coludRenderDiscovery/discovery"
	"coludRenderDiscovery/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"strconv"
)

type ManagerController struct {
	beego.Controller
}

func (c *ManagerController) Prepare() {
	c.Data["PROTOCOL_TYPE_REG_RENDER"] = discovery.PROTOCOL_TYPE_REG_RENDER
	c.Data["PROTOCOL_TYPE_NOTIFY_WEB"] = discovery.PROTOCOL_TYPE_NOTIFY_WEB

	c.Data["RENDER_WAIT"] = discovery.RENDER_WAIT
	c.Data["RENDER_WORK"] = discovery.RENDER_WORK
	c.Data["RENDED_ERR"] = discovery.RENDED_ERR
	c.Data["RENDER_FINISH"] = discovery.RENDER_FINISH

	c.Data["CONTROL_START"] = discovery.CONTROL_START
	c.Data["CONTROL_STOP"] = discovery.CONTROL_STOP
	c.Data["CONTROL_RESTART"] = discovery.CONTROL_RESTART
	c.Data["CONTROL_SCREEN"] = discovery.CONTROL_SCREEN

	c.Data["DESIGNATEIP"] = discovery.DESIGNATEIP
}

func (c *ManagerController) Get() {
	action := c.GetString("action")
	if action == "tasks" {
		c.Data["json"] = c.tasks()
		c.ServeJSON()

	} else if action == "deleteTask" {
		c.Data["json"] = c.deleteTask()
		c.ServeJSON()

	} else if action == "" {
		c.Layout = "manage/layout.html"
		c.LayoutSections = map[string]string{
			"Scripts": "manage/manage_scripts.html",
		}
		c.TplName = "manage/manage.html"
	}
}

func (c *ManagerController) tasks() []map[string]interface{} {
	o := orm.NewOrm()
	tasks := make([]models.RenderTask, 0)
	qs := o.QueryTable("RenderTask").Filter("Del", false)
	qs.OrderBy("Id").All(&tasks)
	lst := make([]map[string]interface{}, 0)
	num := 0
	for _, v := range tasks {
		lst = append(lst, map[string]interface{}{
			"num": func() int {
				if !v.IsGroupItem {
					num += 1
				}
				return num
			}(),
			"order":         v.Order,
			"id":            strconv.Itoa(v.Id),
			"ip":            v.Ip,
			"date":          v.Date.Format("2006/01/02 15:04:05"),
			"level":         v.Level,
			"group_num":     v.GroupNum,
			"is_group_item": v.IsGroupItem,
			"name":          v.Name,
			"work_status":   c.workStatus(v.Id),
		})
	}

	return lst
}

func (c *ManagerController) workStatus(id int) int {
	renderTaskUsage := models.RenderTaskUsage{}
	qs := orm.NewOrm().QueryTable("RenderTaskUsage").Filter("TaskId", id)
	err := qs.Limit(1).One(&renderTaskUsage, "WorkStatus")
	if err != nil {
		return discovery.RENDER_WAIT
	}

	return renderTaskUsage.WorkStatus
}

func (c *ManagerController) deleteTask() map[string]interface{} {
	id, _ := c.GetInt("id")
	response := map[string]interface{}{
		"status":   false,
		"id":       id,
		"group_id": 0,
	}

	if id < 1 {
		return response
	}

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return response
	}

	qs := o.QueryTable("RenderTask")
	renderTask := models.RenderTask{}
	currTask := qs.Filter("Id", id)
	if err := currTask.Limit(1).One(&renderTask, "Order"); err != nil {
		return response
	}

	if _, err := currTask.Delete(); err != nil {
		return response
	}

	n, err := qs.Filter("Order", renderTask.Order).Count()
	if err != nil {
		o.Rollback()
		return response
	}

	if n == 0 {
		if _, err := qs.Filter("Id", renderTask.Order).Delete(); err != nil {
			o.Rollback()
			return response
		}

		response["group_id"], _ = strconv.Atoi(renderTask.Order)
	}

	o.Commit()
	response["status"] = true
	return response
}
