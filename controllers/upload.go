package controllers

import (
	"coludRenderDiscovery/discovery"
	"coludRenderDiscovery/models"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/golibs/uuid"
	"github.com/gorilla/websocket"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type UploadController struct {
	beego.Controller
	responseMsg ResponseMsg
}

// 上传文件返回结构
type ResponseMsg struct {
	Code int16        // 状态码
	Err  string       // 错误描述
	Time time.Time    // 处理时间
	Data responseData // 返回数据
}

type responseData struct {
	Name     string
	TaskId   string
	SavePath string
	GroupNum int16
	Cores    int16
}

func (c *UploadController) wsConn() (*websocket.Conn, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:" + beego.AppConfig.String("default::httpport"),
		Path:   "/ws",
	}

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (c *UploadController) base64(s string) string {
	b4 := base64.NewEncoding(discovery.BASE64_TABLE)
	return b4.EncodeToString([]byte(s))
}

func (c *UploadController) Post() {
	c.responseMsg.Data = responseData{}
	c.responseMsg.Code = 500
	c.responseMsg.Time = time.Now()
	c.responseMsg.Data.TaskId = uuid.Rand().Hex()
	c.responseMsg.Data.GroupNum, _ = c.GetInt16("groupNum")
	c.responseMsg.Data.Cores, _ = c.GetInt16("cores")
	c.responseMsg.Data.Name = c.GetString("name")
	if c.responseMsg.Data.GroupNum < 1 {
		c.responseMsg.Data.GroupNum = 1
	}

	defer func() {
		c.Data["json"] = c.responseMsg
		c.ServeJSON()
	}()

	savePath, err := c.upfile()
	if err != nil {
		c.responseMsg.Err = err.Error()
		return
	}

	c.responseMsg.Data.SavePath = savePath
	if err := c.idb(); err != nil {
		c.responseMsg.Err = err.Error()
		return
	}

	err = c.notice()
	if err != nil {
		c.responseMsg.Err = err.Error()
		return
	}

	c.responseMsg.Code = 200
}

// 通知WEB端数据有变化
func (c *UploadController) notice() error {
	render := discovery.Render{}
	render.BroadcastRenderOnline()
	return nil
}

/**
处理上传附件
*/
func (c *UploadController) upfile() (string, error) {
	saveDir, err := c.mkDir()
	if err != nil {
		return "", err
	}

	f, h, err := c.GetFile("file")
	if err != nil {
		return "", err
	}

	f.Close()
	savePath := path.Join(saveDir, fmt.Sprintf("%s%s", c.responseMsg.Data.TaskId, path.Ext(h.Filename)))
	if err := c.SaveToFile("file", savePath); err != nil {
		return "", err
	}

	return savePath, nil
}

// 创建文件存储文件夹
func (c *UploadController) mkDir() (string, error) {
	file, _ := exec.LookPath(os.Args[0])
	runPath, _ := filepath.Abs(file)
	runPath = filepath.Dir(runPath)
	runPath = strings.Replace(runPath, `\`, "/", -1)
	now := time.Now()
	saveDir := path.Join(runPath,
		beego.AppConfig.String("upfile::SaveDir"),
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
	)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", err
	}

	return saveDir, nil
}

// 写入数据库
func (c *UploadController) idb() error {
	var id int64
	var err error
	var order = c.responseMsg.Data.TaskId
	var isGroupItem = false
	o := orm.NewOrm()
	if c.responseMsg.Data.GroupNum > 1 {
		if id, err = o.Insert(&models.RenderTask{
			Order:    order,
			GroupNum: c.responseMsg.Data.GroupNum,
			Name:     c.responseMsg.Data.Name,
		}); err != nil {
			return err
		}

		isGroupItem = true
		order = strconv.Itoa(int(id))
	}

	i := int16(1)
	renderTasks := make([]models.RenderTask, 0)
	for ; i <= c.responseMsg.Data.GroupNum; i++ {
		renderTasks = append(renderTasks, models.RenderTask{
			Order:       order,
			FilePath:    c.responseMsg.Data.SavePath,
			Xml:         c.base64(c.GetString(fmt.Sprintf("xml%d", i))),
			Ip:          c.Ctx.Input.IP(),
			WorkStatus:  discovery.RENDER_WAIT,
			GroupNum:    1,
			IsGroupItem: isGroupItem,
			Cores:       c.responseMsg.Data.Cores,
		})
	}

	renderTasksLen := len(renderTasks)
	if renderTasksLen > 0 {
		if successNum, err := o.InsertMulti(renderTasksLen, renderTasks); err != nil {
			return err
		} else if int(successNum) != renderTasksLen {
			return errors.New(discovery.ERR_SUCCESS_NUM)
		}
	}

	return nil
}
