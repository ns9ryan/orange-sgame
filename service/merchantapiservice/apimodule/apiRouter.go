package apimodule

import (
	"errors"
	//"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	utils "videogame/service/merchantapiservice/apimodule/utils"
	"strings"
	"net"
	"fmt"
	models "videogame/common/model"
)

type MerchantApiModule struct {
	service.Module
	gin    *gin.Engine
	router *gin.RouterGroup

	serviceConfig map[string]interface{}
}

func (h *MerchantApiModule) OnInit() error {
	// 读取服务配置
	conf := h.GetService().GetServiceCfg()
	if conf == nil {
		log.Error("找不到服务配置")
		return errors.New("找不到服务配置")
	}
	cc, ok := conf.(map[string]interface{})
	if !ok {
		log.Error("服务配置格式错误")
		return errors.New("服务配置格式错误")
	}

	h.serviceConfig = cc
	httpListen := cc["HttpListen"].(string)

	// 初始化Gin引擎
	h.gin = gin.New()
	h.gin.Use(gin.Recovery(), gin.Logger())

	// 设置跨域等中间件
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"} // 根据需要修改允许的源
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Token", "Ortoken", "Merchantcode", "pageIndex"}
	
	corsConfig.AllowCredentials = false

	h.gin.Use(cors.New(corsConfig))

	// 设置入口路由
	h.router = h.gin.Group("/orange/api/v1/")
	merchantGroup := h.router.Group("merchant/").Use(CheckHeader())

	// 商户预先要初始化好，在CRM系统里预先完成以下操作：

	// 1、创建商户，生成merchantCode商户唯一识别码例如2YNE87WT，商户把他放在api请求Header里；
	// 2、生成这个商户的验证token，放在api请求Header里，token目前是1个月有效，也能重新生成；
	// 3、告诉商户我们api服务的访问地址，例如https://s6.gamining.com/orange/api/merchant/；
	// 4、商户编辑填入它的来源出口IP白名单(可以为空)，api服务将根据IP白名单判断是否合法请求；
	// 5、商户填入各个子系统的回调地址，比如单钱包子系统(必填)、防沉迷子系统(后续)、异常处理页子系统(必填)、登出游戏后访问的子系统地址(后续考虑)等；



	// 一、玩家账号相关，MerchantApiPlayer.go
	// 创建游戏账号(返回操作状态码，不进入游戏，不返回游戏地址)
	merchantGroup.POST("/player_create", h.CreatePlayer)

	// 玩家进入游戏(如果账号没创建则创建账号并返回游戏地址，如果有账号了则验证并返回游戏地址)
	merchantGroup.POST("/player_enter", h.EnterGame)

	// 商户方修改了玩家的个人信息之后调用
	merchantGroup.POST("/player_update_info", h.UpdatePlayerInfo)


	// 二、钱包余额相关，MerchantApiWallet.go
	// 多钱包模式，商户->请求->Orange Api服务，并要进行定期(比如每日)对账

	// 多钱包模式，商户查询玩家在我们游戏数据库里的实时余额
	merchantGroup.POST("/wallet_balance", h.GetBalance)

	// 多钱包模式，商户玩家往我们游戏数据库里存款上分
	merchantGroup.POST("/wallet_deposit", h.Deposit)

	// 多钱包模式，商户玩家从我们游戏数据库里取现下分
	merchantGroup.POST("/wallet_withdraw", h.Withdraw)

	// 多钱包模式，因网络超时异常未响应上下分动作时，商户方调此接口，核对上一个订单信息
	merchantGroup.POST("/wallet_validate_transfer", h.ValidateTransfer)

	// 多钱包模式，对账接口(按游戏/按桌台/按玩家等统计注单量，请求JSON里定义查询条件)
	merchantGroup.POST("/wallet_player_report", h.PlayerReport)


	// 单钱包模式，Orange Wallet服务->请求->商户，我们游戏钱包服的每笔下注/开奖/小费都需要访问商户单钱包子系统提供的接口

	// 单钱包，商户方要给我们钱包服提供哪些单一钱包APIs，约定文档重点参考"EVO API document-单一钱包.docx"，待结合EVO,OBLive,COG三者，整理我们的规范约定


	// 三、商户面板展示与服务状态维护相关，MerchantApiMaintenance.go

	// 校验API服务器接口是否正常，可用于定期心跳检测
	merchantGroup.GET("/ok", h.CheckApiOk)	

	// 获取游戏分类目录，可用于商户方的面板展示
	merchantGroup.GET("/get_game_catalog", h.GetGameCatalog)

	// 获取各个游戏下注玩法说明，可用于商户方的面板展示
	merchantGroup.GET("/get_game_rules", h.GetGameRules)

	// 获取游戏各桌子的游戏的当前维护状态，正常(开始下注停止下注等)/休场/维护中，可用于玩家进入游戏前的检查
	merchantGroup.GET("/get_game_tables", h.GetGameTables)

	// 查询游戏各桌子在线玩家状态列表统计，可用于商户方的面板展示
	merchantGroup.GET("/online_players", h.OnlinePlayers)

	// 查询游戏各桌子在线玩家状态列表统计，可用于商户方的面板展示
	merchantGroup.POST("/update_token", h.GetGameCatalog)//h.UpdateToken)


	// 四、游戏历史对局查询相关，MerchantApiGame.go

	// 各桌台/玩家的游戏对局查询(请求JSON里定义查询条件)
	merchantGroup.POST("/game_records_bet", h.GameRecordsBet)

	// 荷官打赏明细记录查询(请求JSON里定义查询条件)
	merchantGroup.POST("/game_records_tip", h.GameRecordsTip)

	// 启动 HTTP 服务
	go func() {
		if err := h.gin.Run(httpListen); err != nil {
			log.Fatal("HTTP服务启动失败:", log.ErrorField("error", err))
		}
	}()

	return nil
}

// 请求的安全预检查：merchantCode + token + 白名单IP

// 1、检查请求头Header里是否有商户merchantCode和token，并检查code是否合法，token是否有效
// 2、检查解析出来的商户是否有设置IP白名单，如果有再检查请求的IP地址是否在白名单里

func CheckHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过某些特定路由
		if c.FullPath() == "/merchant/api/merchant/ok" {
			c.Next()
			return
		}

		// 先获取 merchantCode、token 和 IP
		merchantCode := c.GetHeader("Merchantcode")
		token := c.GetHeader("Ortoken")
		clientIP := getClientIP(c)

		fmt.Printf("111111111 Header, merchantCode=%s, token=%s, clientIP=%s\n", merchantCode, token, clientIP)

		// 如果为空直接返回
		if merchantCode == "" || token == "" {
			ReturnHttpApi(c, 10001, "没有merchantCode, token", "")
			return
		}

		// 如果为空直接返回
		if clientIP == "" {
			ReturnHttpApi(c, 10001, "来源IP未知", "")
			return
		}

		merchant, err := models.GetMerchantByCode(merchantCode)
		if err != nil {
			ReturnHttpApi(c, 10001, "没有找到商户信息", "")
			return
		}

		// 检查 IP 是否在白名单中
		if !checkIPInWhiteList(merchant.WhiteList, clientIP) {
			ReturnHttpApi(c, 10001, "来源IP不在白名单中", "")
			return
		}

		newToken, _ := utils.GenerateMerchantToken(merchantCode, []byte(merchant.OrTokenKey))

		fmt.Printf("===newToken=%s\n", newToken)

		// 检查 merchantCode 是否合法
		tokenParseMerchantCode, err := utils.ValidateMerchantToken(token, []byte(merchant.OrTokenKey))

		if err != nil {
			fmt.Printf("11111 ValidateMerchantToken err=%v\n", err)
			ReturnHttpApi(c, 10001, "ValidateMerchantToken失败", "")
			return
		}

		fmt.Printf("11111 tokenParseMerchantCode=%s, merchantCode=%s\n", tokenParseMerchantCode, merchantCode)

		c.Next()
	}
}

// 获取请求来源IP
func getClientIP(c *gin.Context) string {
	// 优先使用 X-Real-IP
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// 其次使用 X-Forwarded-For（可能包含多个 IP，取第一个）
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 默认获取客户端 IP
	return c.ClientIP()
}

// checkIPInWhiteList 判断 clientIP 是否在白名单中
func checkIPInWhiteList(whiteListStr, clientIP string) bool {
	if whiteListStr == "" {
		fmt.Printf("商户未设置白名单IP列表\n")
		return true
	}

	clientIP = strings.TrimSpace(clientIP)
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	whiteList := strings.Split(whiteListStr, ",")
	for _, item := range whiteList {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// 精确 IP 匹配
		if item == clientIP {
			return true
		}

		// CIDR 匹配
		if strings.Contains(item, "/") {
			_, ipNet, err := net.ParseCIDR(item)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		}
	}

	fmt.Printf("===商户白名单IP列表:%s, 来源IP:%s\n", whiteListStr, clientIP)
	return true
}

// ApiRequest 定义请求结构体
type ApiRequest struct {
	MerchantCode string `json:"merchantCode"`
	Signature    string `json:"signature"`
	Params       string `json:"params"`
}

// ApiResponseOk 定义响应结构体
type ApiResponseOk struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Request interface{} `json:"request"`
}

type ApiResponseOkSimple struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ApiResponseError 定义响应结构体
type ApiResponseError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Request interface{} `json:"request"`
}

// ReturnHttpError 返回错误响应
func ReturnHttpError(c *gin.Context, code int, message string, originJson interface{}) {
	// 如果 code 为空，则使用默认值 9000
	if code == 0 {
		code = 9000
	}

	if originJson == nil {
		originJson = ""
	}

	c.JSON(http.StatusOK, ApiResponseError{
		Code:    code,       // 错误代码示例
		Message: message,    // 消息提示
		Request: originJson, // 请求的业务参数，原样返回
	})
}

// ReturnHttpOK 用于返回成功响应
func ReturnHttpOK(c *gin.Context, data interface{}, originJson interface{}) {
	if data == nil {
		data = ""
	}

	if originJson == nil {
		originJson = ""
	}

	c.JSON(http.StatusOK, ApiResponseOk{
		Code:    200,        // 成功的状态码
		Message: "成功",      // 消息提示
		Data:    data,       // 实际业务数据
		Request: originJson, // 请求的业务参数，原样返回
	})
}

// ReturnHttpOKSimple ReturnHttpOK 用于返回成功响应
func ReturnHttpOKSimple(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, ApiResponseOkSimple{
		Code:    code,
		Message: msg,
	})
}

// ReturnHttp 用于返回code+msg+data成功响应
func ReturnHttp(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}