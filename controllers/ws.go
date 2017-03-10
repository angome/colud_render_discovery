package controllers

import (
	"coludRenderDiscovery/discovery"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/gorilla/websocket"
	"strconv"
	"time"
)

type Wsc struct {
	beego.Controller
	ip string
}

func (c *Wsc) Prepare() {
	c.ip = c.Ctx.Input.IP()
	if c.ip == "" {
		c.Abort("403")
	}
}

func (c *Wsc) ClientServer() { // CMD连接
	if err := c.conn(discovery.CLIENT_CMD); err == nil {
		go c.handle(discovery.CLIENT_CMD)
		c.readMsg(discovery.CLIENT_CMD)
	}

	c._defer(discovery.CLIENT_CMD)
}

func (c *Wsc) WebServer() { // WEB连接
	if err := c.conn(discovery.CLIENT_WEB); err == nil {
		go c.handle(discovery.CLIENT_WEB)
		c.readMsg(discovery.CLIENT_WEB)
	}

	c._defer(discovery.CLIENT_WEB)
}

func (c *Wsc) _defer(t int) {
	m := c.wsConnType(t)
	render := new(discovery.Render)
	render.Ip = c.ip
	render.Offline() // 修改渲染机为离线状态

	if _, ok := m[c.ip]; ok {
		m[c.ip].Conn.Close()
		discovery.Lock.Lock()
		delete(m, c.ip)
		discovery.Lock.Unlock()
	}

	c.Ctx.Output.Body([]byte(""))
}

func (c *Wsc) wsConnType(t int) (m map[string]*discovery.WsConn) {
	switch t {
	case discovery.CLIENT_CMD:
		m = discovery.RenderMachineConn
	case discovery.CLIENT_WEB:
		m = discovery.WebConn
	}

	return m
}

func (c *Wsc) conn(t int) error {
	m := c.wsConnType(t)
	if _, ok := m[c.ip]; ok {
		return errors.New(discovery.ERR_IP_EXISTS)
	}

	ws, err := websocket.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request,
		nil, 0x400, 0x400)
	if err != nil {
		return err
	}

	wsConn := new(discovery.WsConn)
	wsConn.Ip = c.ip
	wsConn.Conn = ws
	wsConn.InnerChan = make(chan discovery.WsMsg, 10)
	wsConn.ConnType = t
	discovery.Lock.Lock()
	m[c.ip] = wsConn
	discovery.Lock.Unlock()
	return err
}

func (c *Wsc) handle(t int) {
	m := c.wsConnType(t)
	clientConn := m[c.ip]
	for {
		msg, ok := <-clientConn.InnerChan
		switch msg.Type {
		case discovery.PROTOCOL_TYPE_REG_RENDER:
			render := new(discovery.Render)
			render.Ip = c.ip
			render.Register()

		case discovery.PROTOCOL_TYPE_HEART_BEAT:
			c.heartBeat(clientConn.Conn)

		case discovery.CONTROL_START:
			c.designateIp(clientConn.Conn, msg.Id)

		case discovery.CONTROL_STOP:
		case discovery.CONTROL_RESTART:
		case discovery.CONTROL_SCREEN:

		default:

		}

		if !ok {
			return
		}
	}
}

// 查找空闲渲染机
func (c *Wsc) designateIp(clientConn *websocket.Conn, id int) {
	render := discovery.Render{}
	ips := render.FindFreeDesignateIp()
	responseMsg := discovery.WsMsg{Id: id, Type: discovery.DESIGNATEIP}
	if len(ips) == 0 {
		responseMsg.Err = discovery.ERR_NO_FREE_MACHINE
		wsWriteJson(clientConn, responseMsg)
		return
	}

	ip := ips[0]
	machineMsg, err := c.postMachine(ip, map[string]string{
		"action": strconv.Itoa(discovery.DESIGNATEIP),
	})
	if err != nil {
		responseMsg.Err = discovery.ERR_MAX_FAILURE
		wsWriteJson(clientConn, responseMsg)
		return
	}

	if !machineMsg.Status {
		responseMsg.Err = machineMsg.Err
		wsWriteJson(clientConn, responseMsg)
		return
	}

	if err := render.RenderTaskUsageInstall(id, ip); err != nil {
		responseMsg.Err = discovery.ERR_DESIGNATE_MACHINE
		wsWriteJson(clientConn, responseMsg)
		return
	}

	wsWriteJson(clientConn, responseMsg)
}

func (c *Wsc) postMachine(ip string, params map[string]string) (discovery.MachineMsg, error) {
	msg := discovery.MachineMsg{}
	req := httplib.Post(fmt.Sprintf("http://%s:%d/post", ip, 8081))
	if beego.AppConfig.String("default::RunMode") == "dev" {
		req = req.Debug(true)
	}

	req = req.SetTimeout(time.Second*60, time.Second*60)
	st := fmt.Sprintf("%d", time.Now().Add(time.Second*0x378).Unix())
	params["QM_INPUT_CHEKC"] = fmd5(st)
	req = req.Header("QM_INPUT_CHEKC", st)
	for key, val := range params {
		req = req.Param(key, val)
	}

	response, err := req.Response()
	if err != nil {
		return msg, err
	}

	if response.StatusCode != 200 {
		return msg, errors.New(discovery.ERR_MAX_FAILURE)
	}

	body, err := req.Bytes()
	if err != nil {
		return msg, err
	}

	if err := json.Unmarshal(body, &msg); err != nil {
		return msg, err
	}

	return msg, nil
}

func (c *Wsc) readMsg(t int) {
	m := c.wsConnType(t)
	conn := m[c.ip]
	ch := conn.InnerChan

	for {
		m := discovery.WsMsg{}
		if err := conn.Conn.ReadJSON(&m); err != nil {
			break
		}

		if m.Type > 0 {
			ch <- m
		}
	}

	close(ch)
}

func (c *Wsc) heartBeat(conn *websocket.Conn) {
	wsWriteJson(conn, discovery.WsMsg{Type: discovery.PROTOCOL_TYPE_HEART_BEAT})
}

func wsWriteJson(conn *websocket.Conn, msg discovery.WsMsg) {
	if err := conn.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		return
	}

	conn.SetWriteDeadline(time.Time{})
}

func fmd5(s string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(s))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}
