package models

import "time"

// 渲染机
type RenderMachine struct {
	Id          int
	Ip          string `orm:"size(15);unique"`
	Online      bool   // 是否在线
	GroupId     string `orm:"size(36)"` // 组编号
	Name        string `orm:"size(50)"` // 名称
	OnlineTime  int64  // 上线时间
	OfflineTime int64  // 下线时间
}

// 渲染任务列表
type RenderTask struct {
	Id          int
	Level       int       // 优先级
	WorkStatus  int       // 状态
	Order       string    `orm:"size(36);unique"`             // 任务编号
	Date        time.Time `orm:"auto_now_add;type(datetime)"` // 提交时间
	FilePath    string    `orm:"size(250)"`                   // 上传文件保存路径
	Xml         string    `orm:"type(text)"`                  // 具体参数
	Ip          string    `orm:"size(15)"`                    // 上传机IP
	Del         bool      // 是否删除
	GroupNum    int16     // 分组数
	IsGroupItem bool      // 是否组子项
	Name        string    // 任务总名称
	Cores       int16     // 渲染核心数

}

// 渲染资源利用详细表
type RenderTaskUsage struct {
	Id          int
	WorkStatus  int    // 状态
	TaskId      int    // 任务ID
	StartTime   int    // 渲染启动时间
	EndTime     int    // 渲染结束数据
	DesignateIp string `orm:"size(15)"` // 渲染机IP
}

// 系统设置
type RenderSetup struct {
	Id               int
	MaxRenderTimeout int8 // 渲染超时最大值
}
