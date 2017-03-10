package routers

import (
	"coludRenderDiscovery/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/upfile", &controllers.UploadController{})

	beego.Router("/manager", &controllers.ManagerController{})
	beego.Router("/machine", &controllers.MachineController{})
	beego.Router("/setup", &controllers.SetupController{})

	beego.Router("/ws", &controllers.Wsc{}, "get:ClientServer")
	beego.Router("/ws/web", &controllers.Wsc{}, "get:WebServer")
}
