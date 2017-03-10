package discovery

import (
	"github.com/gorilla/websocket"
	"sync"
)

// 定义客户端类型
const (
	CLIENT_CMD = 1 + iota // CMD连接方式
	CLIENT_WEB            // web连接方式
)

// 定义协议类型
const (
	PROTOCOL_TYPE_REG_RENDER = 1 + iota // 注册渲染机
	PROTOCOL_TYPE_NOTIFY_WEB            // 通知WEB端
	PROTOCOL_TYPE_HEART_BEAT            // 心跳包

	// 控制相关
	CONTROL_START   // 启动
	CONTROL_STOP    // 停止
	CONTROL_RESTART // 重提
	CONTROL_SCREEN  // 拍屏

	// 渲染状态
	RENDER_WAIT   // 等待渲染
	RENDER_WORK   // 渲染中
	RENDED_ERR    // 渲染错误或失败
	RENDER_FINISH // 渲染完成

	// 分配渲染IP
	DESIGNATEIP
)

const (
	ERR_IP_EXISTS         = "IP已存在"
	ERR_SUCCESS_NUM       = "写入数据长度不一致"
	ERR_NO_FREE_MACHINE   = "无空闲渲染机"
	ERR_DESIGNATE_MACHINE = "分配渲染机失败"
	ERR_MAX_FAILURE       = "渲染器通讯失败，检查渲染机是否死机状态"
	ERR_RENDER_FAILURE   = "渲染失败"
)

const (
	BASE64_TABLE = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var (
	Lock              sync.Mutex
	RenderMachineConn = make(map[string]*WsConn, 0)
	WebConn           = make(map[string]*WsConn, 0)
)

type WsConn struct {
	InnerChan chan WsMsg
	Conn      *websocket.Conn
	Ip        string
	ConnType  int
}

type WsMsg struct {
	Id   int
	Type int
	Err  string
}

type MachineMsg struct {
	Status bool
	Err    string
}
