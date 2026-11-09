package coreservice

import (
	"errors"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"videogame/common/eventid"
	"videogame/common/msg_struct_crm"
)

//const defaultLisenAddr = "192.168.0.147:9444"

type ServiceNode struct {
	SType int //游戏服类型  0 :游戏服务器  1:网关服 2:大厅服   3:钱包服 4: 投注服 5： 结算服

	WalletType uint //钱包服时用0：多钱包， 1：单钱包

	GameId      string //游戏ID
	ServiceName string //服务名
	DeskNo      string //桌号

	IpAddr string //IP地址
}

type HttpCoreModule struct {
	service.Module

	gin    *gin.Engine
	router *gin.RouterGroup

	serviceConfig map[string]interface{}
}

func (h *HttpCoreModule) OnInit() error {

	// 读取服务配置
	conf := h.GetService().GetServiceCfg()

	if conf == nil {
		log.SError("服务配置缺失")
		return errors.New("service config error")
	}

	cc, ok := conf.(map[string]interface{})
	if !ok {
		log.SError("服务配置格式错误")
		return errors.New("invalid service config format")
	}

	h.serviceConfig = cc
	httpListen := cc["HttpListen"].(string)

	// 初始化 Gin 引擎
	h.gin = gin.New()
	h.gin.Use(gin.Recovery(), gin.Logger())

	// 设置跨域中间件
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"} // 根据需要修改允许的源
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = false
	h.gin.Use(cors.New(corsConfig))

	h.gin.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "Welcome to Video game!!!")
	})

	// 设置路由
	h.router = h.gin.Group("/core")
	h.router.POST("/createmerchant", h.CreateMerchant)
	h.router.POST("/updatemerchant", h.UpdateMerchant)
	h.router.POST("/operateplanid", h.OperatePlandId)
	h.router.POST("/operategameconfig", h.OperateGameConfig)

	//	h.gin.POST("/login", h.login)                             //登录

	// 启动 HTTP 服务
	go func() {
		if err := h.gin.Run(httpListen); err != nil {
			log.SFatal("HTTP服务启动失败:", err.Error())
		}
	}()

	//以下要测试https可以开启
	//go func() {
	//	keyFile := "C:\\go_work\\orange-sgame\\config\\data\\pem\\key.pem"
	//	certFile := "C:\\go_work\\orange-sgame\\config\\data\\pem\\cert.pem"
	//
	//	if err := h.gin.RunTLS(httpListen, certFile, keyFile); err != nil {
	//		log.SFatal("HTTPS 服务启动失败:", err.Error())
	//	}
	//}()

	return nil
}

// 新建商户
func (h *HttpCoreModule) CreateMerchant(c *gin.Context) {

	var merchantItem msg_struct_crm.MerchantItem

	if err := c.ShouldBindBodyWith(&merchantItem, binding.JSON); err != nil {

		log.SError("绑定请求数据失败:", err.Error())

		ResponseFail(c, 0, merchantItem.RequestId, msg_struct_crm.ErrParam)

		return
	}

	log.Infof("CreateMerchant:%v", merchantItem)

	if merchantItem.MsgId != msg_struct_crm.CrmMsgId_CreateMerchant {
		ResponseFail(c, merchantItem.MsgId, merchantItem.RequestId, msg_struct_crm.ErrIdError)
		return
	}

	//回复收到成功
	ResponseSuccess(c, merchantItem.MsgId, merchantItem.RequestId)

	//通知服务事件
	h.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Core_CreateMerChant,
		Data: &merchantItem,
	})

}

// 更新商户数据
func (h *HttpCoreModule) UpdateMerchant(c *gin.Context) {

	var merchantItem msg_struct_crm.MerchantItem

	if err := c.ShouldBindBodyWith(&merchantItem, binding.JSON); err != nil {

		log.SError("绑定请求数据失败:", err.Error())

		ResponseFail(c, 0, merchantItem.RequestId, msg_struct_crm.ErrParam)

		return
	}

	log.Infof("UpdateMerchant:%v", merchantItem)

	if merchantItem.MsgId != msg_struct_crm.CrmMsgId_UpdateMerchant {
		ResponseFail(c, merchantItem.MsgId, merchantItem.RequestId, msg_struct_crm.ErrIdError)
		return
	}

	//通知成功了，回复
	ResponseSuccess(c, merchantItem.MsgId, merchantItem.RequestId)

	//通知服务事件
	h.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Core_UpdateMerChant,
		Data: &merchantItem,
	})

}

// 更新商户数据
func (h *HttpCoreModule) OperatePlandId(c *gin.Context) {

	var PlanIdItem msg_struct_crm.PlanIdItem

	if err := c.ShouldBindBodyWith(&PlanIdItem, binding.JSON); err != nil {

		log.SError("绑定请求数据失败:", err.Error())
		ResponseFail(c, 0, PlanIdItem.RequestId, msg_struct_crm.ErrParam)
		return
	}

	log.Infof("OperatePlandId:%v", PlanIdItem)

	if PlanIdItem.MsgId != msg_struct_crm.CrmMsgId_OperatePlanId {
		ResponseFail(c, PlanIdItem.MsgId, PlanIdItem.RequestId, msg_struct_crm.ErrIdError)
		return
	}

	//处理成功了，回复
	ResponseSuccess(c, PlanIdItem.MsgId, PlanIdItem.RequestId)

	//通知服务事件
	h.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Core_UpdateTemplatePlanId,
		Data: &PlanIdItem,
	})

}

// 操作游戏配置
func (h *HttpCoreModule) OperateGameConfig(c *gin.Context) {

	var setItem msg_struct_crm.GameConfigItem

	if err := c.ShouldBindBodyWith(&setItem, binding.JSON); err != nil {

		log.SError("绑定请求数据失败:", err.Error())

		ResponseFail(c, 0, setItem.RequestId, msg_struct_crm.ErrParam)

		return
	}

	log.Infof("OperateGameConfig:%v", setItem)

	if setItem.MsgId != msg_struct_crm.CrmMsgId_OperateGameConfig {
		ResponseFail(c, setItem.MsgId, setItem.RequestId, msg_struct_crm.ErrIdError)
		return
	}

	//处理成功了，回复
	ResponseSuccess(c, setItem.MsgId, setItem.RequestId)

	h.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Core_OperateGameConfig,
		Data: &setItem,
	})

}
