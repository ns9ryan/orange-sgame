package apimodule

import (
	"errors"
	//"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	//"videogame/service/merchantapiservice/apimodule/utils"
	"strings"
	"net"
)

type MerchantApiModuleImitate struct {
	service.Module
	gin    *gin.Engine
	router *gin.RouterGroup
}

func (h *MerchantApiModuleImitate) OnInit() error {
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
	httpListen := cc["HttpListen"].(string)

	// 初始化Gin引擎
	h.gin = gin.New()
	h.gin.Use(gin.Recovery(), gin.Logger())

	// 设置跨域等中间件
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"} // 根据需要修改允许的源
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Token", "token", "MerchantCode", "pageIndex"}
	h.gin.Use(cors.New(corsConfig))

	// 设置入口路由
	h.router = h.gin.Group("/ns9/api/")

	//模拟作为单钱包的商户端，自测
	singleWalletGroup := h.router.Group("gvs/orange/")
	singleWalletGroup.GET("/healthcheck/", h.CheckSingleWalletOk)
	singleWalletGroup.POST("/game-round/settle/", h.RoundSettle)
	singleWalletGroup.POST("/single-wallet/get-balance/", h.GetBalanceSingleWallet)
	singleWalletGroup.POST("/single-wallet/debit/", h.DebitSingleWallet)
	singleWalletGroup.POST("/single-wallet/cancel-debit/", h.CancelDebitSingleWallet)
	singleWalletGroup.POST("/single-wallet/credit/", h.CreditSingleWallet)

	// 启动 HTTP 服务
	go func() {
		if err := h.gin.Run(httpListen); err != nil {
			log.Fatal("HTTP服务启动失败:", log.ErrorField("error", err))
		}
	}()

	return nil
}

// 模拟请求的安全预检查：token + 白名单IP

// 1、检查请求头Header里是否有商户merchantCode和token，并检查code是否合法，token是否有效
// 2、检查解析出来的商户是否有设置IP白名单，如果有再检查请求的IP地址是否在白名单里

func CheckHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过某些特定路由
		if c.FullPath() == "/merchant/api/v1/ok" {
			c.Next()
			return
		}

		// 先获取 token 和 IP
		token := c.GetHeader("OrToken")
		clientIP := getClientIP(c)

		// 如果为空直接返回
		if token == "" || clientIP == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing merchantCode, token, or client IP"})
			c.Abort()
			return
		}

		// 检查 IP 是否在白名单中
		if !isIPAllowed(clientIP) {
			c.JSON(http.StatusForbidden, gin.H{"error": "IP not allowed"})
			c.Abort()
			return
		}

		// 检查 token 是否合法
		if _, exists := validTokens[token]; !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 允许的商户代码集合，TODO待改进从数据库读取
var validMerchantCodes = map[string]bool{
	"2yNe": true,
	"XAqW": true,
}

// 允许的白名单 IP（支持 CIDR），TODO待改进从数据库读取
var whitelistCIDRs = []string{
	"192.168.0.0/24",
	"192.168.1.0/24",
}

// 允许的 Token，TODO待改进从数据库或Redis读取
var validTokens = map[string]bool{
	"token11111111": true,
	"token22222222": true,
}

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

// 判断 IP 是否在白名单中（支持 CIDR）
func isIPAllowed(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range whitelistCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
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

type ApiResponse struct {
	Code    	int32         	`json:"code"`
	Msg 		string      	`json:"msg"`
	Data    	interface{} 	`json:"data"`
}

func ReturnHttpApi(c *gin.Context, code int32, msg string, data interface{}) {
	if data == nil {
		data = ""
	}

	c.JSON(http.StatusOK, ApiResponse{
		Code:    code,     // 成功的状态码
		Msg: 	 msg,      // 消息提示
		Data:    data,     // 实际业务数据		
	})
}