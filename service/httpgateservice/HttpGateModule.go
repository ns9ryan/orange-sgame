package httpgateservice

import (
	"errors"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

const defaultLisenAddr = "192.168.0.147:9444"

type HttpGateModule struct {
	service.Module

	gin    *gin.Engine
	router *gin.RouterGroup

	serviceConfig map[string]interface{}

	MapBalanceInfoMutex sync.Mutex
	MapBalanceInfo      map[string]*BalanceInfo // GateService nodeId为key
}

func (h *HttpGateModule) OnInit() error {

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
	//h.router = h.gin.Group("/check")
	//h.router.POST("/login", h.login)

	h.gin.POST("/login", h.login)                             //登录
	h.gin.POST("/bet_records", h.betRecords)                  //玩家下注订单列表
	h.gin.POST("/user_winlost_account", h.userWinLostAccount) //玩家输赢统计
	h.gin.POST("/chat_list", h.chatList)                      //获取最近10条的群聊消息(如果有商户公告,也一并返回)

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

	//初始化代理,已移动main.go初始化
	//player.InitMerchantGroups()

	//初始化gate各节点nodeid对应的负载map
	h.initMapBalance()

	return nil
}
