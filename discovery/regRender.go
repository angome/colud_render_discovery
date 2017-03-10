package discovery

import (
	"coludRenderDiscovery/models"
	"github.com/astaxie/beego/orm"
	"log"
	"time"
)

type Render struct {
	Ip string
}

func (c *Render) Register() {
	o := orm.NewOrm()
	qs := o.QueryTable("RenderMachine").Filter("Ip", c.Ip)
	n, err := qs.Count()
	if err != nil {
		return
	}

	if n == 0 {
		render := models.RenderMachine{
			Ip:         c.Ip,
			Online:     true,
			OnlineTime: time.Now().Unix(),
		}
		o.Insert(&render)
	} else {
		c.updateStatus()
	}

	c.BroadcastRenderOnline()
}

func (c *Render) Offline() {
	o := orm.NewOrm()
	o.QueryTable("RenderMachine").Filter("Ip", c.Ip).Update(orm.Params{
		"Online":      false,
		"OnlineTime":  0,
		"OfflineTime": time.Now().Unix(),
	})

	c.BroadcastRenderOnline()
}

func (c *Render) updateStatus() {
	o := orm.NewOrm()
	qs := o.QueryTable("RenderMachine").Filter("Ip", c.Ip)
	params := orm.Params{"Online": true, "OfflineTime": 0}
	render := models.RenderMachine{}
	if err := qs.Limit(1).One(&render, "OnlineTime"); err == nil {
		if render.OnlineTime == 0 {
			params["OnlineTime"] = time.Now().Unix()
		}
	}

	qs.Update(params)
}

func (c *Render) BroadcastRenderOnline() {
	for _, conn := range WebConn {
		conn.Conn.WriteJSON(WsMsg{Type: PROTOCOL_TYPE_NOTIFY_WEB})
	}
}

// 查找空闲渲染机
func (c *Render) FindFreeDesignateIp() []string {
	ips := make([]string, 0)
	sql := `
	SELECT A.ip AS Ip FROM render_machine AS A
	LEFT JOIN render_task_usage AS B
	ON A.ip=B.designate_ip
	WHERE ISNULL(B.id) OR (B.id > 0 AND (B.work_status=? OR B.work_status=?))
	`
	type machine struct{ Ip string }
	var machines = make([]orm.Params, 0)
	o := orm.NewOrm()
	_, err := o.Raw(sql, RENDED_ERR, RENDER_FINISH).Values(&machines)
	if err != nil {
		return ips
	}

	for _, v := range machines {
		ip, ok := v["Ip"].(string)
		if !ok {
			continue
		}

		if ip == "" {
			continue
		}

		ips = append(ips, ip)
	}

	log.Println(ips)
	return ips
}

func (c *Render) RenderTaskUsageInstall(id int, ip string) error {
	o := orm.NewOrm()
	_, err := o.Insert(&models.RenderTaskUsage{
		WorkStatus:  RENDER_WORK,
		TaskId:      id,
		DesignateIp: ip,
		StartTime:   int(time.Now().Unix()),
	})

	return err
}
