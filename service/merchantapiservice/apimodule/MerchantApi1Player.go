//*****************************************************************
// OBLiveApi-1-Player.go里，放的是创建账号、进入游戏、快速进入游戏等3个接口。
//*****************************************************************

package apimodule

import (
	//"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
	"strconv"
	//"videogame/service/merchantapiservice/apimodule/utils"
	//"videogame/common/crm_api_struct"
	"videogame/common/global/variable"
	"encoding/json"
	"encoding/base64"
	"videogame/common/proto/rpc"
	"net/http"
	"videogame/common/util"
	redisId "videogame/common/redis"
	"videogame/common/model"
	"videogame/service/games/player"
	"videogame/common/msg_struct"
)

const (
	CrmNodeId = "crm"
	RedisHashId_MerchantPlayers = "Or_MerchantEnterPlayers:"
)

type ApiResponse struct {
	Code    	int32         	`json:"code"`
	Msg 		string      	`json:"message"`
	Data    	interface{} 	`json:"data"`
}

type ApiResponseWithReq struct {
	Code    	int32         	`json:"code"`
	Msg 		string      	`json:"message"`
	Data    	interface{} 	`json:"data"`
	Request    	interface{} 	`json:"request"`
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

	c.Abort()
}

func ReturnHttpApiWithReq(c *gin.Context, code int32, msg string, data interface{}, req interface{}) {
	if data == nil {
		data = ""
	}

	c.JSON(http.StatusOK, ApiResponseWithReq{
		Code:    code,     // 成功的状态码
		Msg: 	 msg,      // 消息提示
		Data:    data,     // 实际业务数据	
		Request: req,	   // 原始请求
	})

	c.Abort()
}

// 创建游戏账号，来自商户的原始请求结构,MCreatePlayerApiReq有别于rpc.CreatePlayerApiReq
type MCreatePlayerApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
	Code			string  `json:"currency"`
	Nickname		string  `json:"nickname"`
	Email			string  `json:"email"`
	Phone			string  `json:"phone"`
	Password		string  `json:"password"`
}

// 创建游戏账号(返回操作状态码，不进入游戏，不返回游戏地址)
func (h *MerchantApiModule) CreatePlayer(c *gin.Context) {
	var merchantReq MCreatePlayerApiReq

	if err := c.ShouldBind(&merchantReq); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApiWithReq(c, 10001, "请求参数错误", "", merchantReq)
		return
	}

	var rpcReq rpc.CreatePlayerApiReq
	var rpcRet rpc.CreatePlayerApiRet

	rpcReq.MerchantCode = merchantReq.MerchantCode
	rpcReq.Username = merchantReq.Username
	rpcReq.Code = merchantReq.Code
	rpcReq.Nickname = merchantReq.Nickname
	rpcReq.Email = merchantReq.Email
	rpcReq.Phone = merchantReq.Phone
	rpcReq.Password = merchantReq.Password
	rpcReq.DeviceType = 0 //从商户后台发来的创建请求，用户的设备类型都记为桌面H5横屏类型

	err := h.Call("CrmRpcService.RPCCreatePlayer", &rpcReq, &rpcRet) 
	if err != nil {
		fmt.Printf("CrmRpcService.RPCCreatePlayer ************************************** Call error: %+v\n", err)

		ReturnHttpApiWithReq(c, 10001, "Call CRM RPC失败", "", merchantReq)
		return		
	}

	fmt.Printf("RPCAddPlayer rpcRet is %v\n", rpcRet)
	if rpcRet.Code != 200 {
		fmt.Printf("CrmService rpcRet.Msg=%s", rpcRet.Msg)
	}

	//创建成功后，因没有登录游戏，先不考虑将玩家信息缓存到Redis里，交给EnterGame()去存缓存
	//DO NOT UserSaveToRedis()

	ReturnHttpApiWithReq(c, rpcRet.Code, rpcRet.Msg, "", merchantReq)
	return
}

// 玩家登录游戏，来自玩家的原始请求结构,MPlayerEnterGameApiReq有别于rpc.CreatePlayerApiReq
type MPlayerEnterGameApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
	Code			string  `json:"currency"`
	Nickname		string  `json:"nickname"`
	Email			string  `json:"email"`
	Phone			string  `json:"phone"`
	Password		string  `json:"password"`
	DeviceType		int32  	`json:"device_type"`

	DirectIntoGame	bool  	`json:"direct_into_game"`
	GameTypeId		string  `json:"game_type_id"`
	GameTableId		int32  	`json:"game_table_id"`
	BackUrl			string  `json:"back_url"`

}

// 进入游戏(如果账号没创建则创建账号并返回游戏地址，如果有账号了则验证并返回游戏地址)
func (h *MerchantApiModule) EnterGame(c *gin.Context) {
	fmt.Printf("11111111 EnterGame\n")

	var playerReq MPlayerEnterGameApiReq

	if err := c.ShouldBind(&playerReq); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	//先检查redis这个商户账号是否存在，注意这个key和redisId.RedisHashId_PlayersKey不一样
	username := strings.ToLower(fmt.Sprintf("%s_%s", playerReq.MerchantCode, playerReq.Username))

	_, err := variable.RedisStorage.HGet(RedisHashId_MerchantPlayers, username)
	if err == nil {

		fmt.Printf("Redis已存在Username，不用再创建 %s\n", username)

		//先到redis检测是否存在这个账号，存在则拿redis缓存账号密码校验，否则跟mysql校验密码
		if h.CheckUserValid(username, playerReq.Password) == false {
			fmt.Printf("密码错误 %s\n", username)
			ReturnHttpApi(c, 10001, "账号密码错误", "")
			return
		}

	} else {

		//请求CRM RPC创建这个账号，不用验证密码
		var rpcReq rpc.CreatePlayerApiReq
		var rpcRet rpc.CreatePlayerApiRet

		rpcReq.MerchantCode = playerReq.MerchantCode
		rpcReq.Username = playerReq.Username
		rpcReq.Code = playerReq.Code
		rpcReq.Nickname = playerReq.Nickname
		rpcReq.Email = playerReq.Email
		rpcReq.Phone = playerReq.Phone
		rpcReq.Password = playerReq.Password
		rpcReq.DeviceType = playerReq.DeviceType //0桌面H5横屏，1手机H5竖屏

		err := h.Call("CrmRpcService.RPCCreatePlayer", &rpcReq, &rpcRet) 

		if err != nil {
			fmt.Printf("CrmRpcService.RPCCreatePlayer ************************************** Call error: %+v\n", err)
			ReturnHttpApi(c, 10001, "Call CRM RPC失败", "")
			return
		}

		fmt.Printf("RPCAddPlayer rpcRet is %v\n", rpcRet)
		if rpcRet.Code != 200 {
			fmt.Printf("CrmService rpcRet.Msg=%s", rpcRet.Msg)
		}

	}

	token, err := util.GenerateUserToken(username)
	if err != nil {
		fmt.Printf("生成token失败 %v\n", err)
		ReturnHttpApi(c, 10001, "Token生成失败", "")
		return
	}

	fmt.Printf("token=%s\n", token)

	var enterLobby msg_struct.LoginEnterLobby

	enterLobby.Token = token //5e2809c4c978c57beb1460201bb2bedc
	enterLobby.UserId = username
	enterLobby.WsAddr = h.getWsAddr()
	enterLobby.Language = "en"
	enterLobby.DeviceType = playerReq.DeviceType

	enterLobby.DirectIntoGame = playerReq.DirectIntoGame
	fmt.Printf("11111111 enterLobby.DirectIntoGame=%b\n", enterLobby.DirectIntoGame)

	if enterLobby.DirectIntoGame {

		gameTable, err := models.GetGameTable(playerReq.GameTypeId, playerReq.GameTableId)
		if err != nil {
			ReturnHttpApi(c, 10001, "获取游戏分类失败", "")
			return
		}

		enterLobby.GroupId = gameTable.GameCategoryCode
		enterLobby.GameId = gameTable.GameTableCode

		enterLobby.DeskNo = playerReq.GameTableId
		enterLobby.BackUrl = playerReq.BackUrl		

	}else{

		enterLobby.GroupId = ""
		enterLobby.GameId = ""
		enterLobby.DeskNo = 0
		enterLobby.BackUrl = ""

	}

	jsonData, err := json.Marshal(enterLobby) //先转换成 JSON
	if err != nil {
	    fmt.Println("JSON 编码失败:", err)
		ReturnHttpApi(c, 10001, "JSON编码失败", "")
		return
	}

	base64ParamEncode := base64.StdEncoding.EncodeToString(jsonData) // 再进行 Base64 编码

	//准备最终返回数据
	gameUrl := h.getHttpAddr()
	data := map[string]string{
		"redirect_url": fmt.Sprintf("%s?resource_id=%s", gameUrl, base64ParamEncode),
	}

	fmt.Printf("11111111 enterLobby.data=%v\n", data)

	//返回之前，将玩家enterLobby信息存入Redis，以便下次如果发现账号已存在不用再创建
	//注意在玩家状态改变、玩家登出、玩家从别的device_type/game_table_id进入时，删除Redis
	reqData, err := json.Marshal(&playerReq)
	if err != nil {
		fmt.Printf("json Marshal失败 %s\n", playerReq.Username)
		ReturnHttpApi(c, 10001, "JSON编码失败", "")
		return
	}

	err = variable.RedisStorage.HSet(RedisHashId_MerchantPlayers, username, string(reqData))
	if err != nil {
		fmt.Printf("Redis Hset失败 %s\n", playerReq.Username)
		ReturnHttpApi(c, 10001, "Redis更新失败", "")
		return
	}

	ReturnHttpApi(c, 200, "", data)
	return
}

func (h *MerchantApiModule) getWsAddr() string {
	//wsArr := "ws://192.168.0.112:9444"

	wsProto := "ws"	//默认值
	ip := "192.168.0.112" //默认值
	port := "9444" //默认值

	certFile := h.serviceConfig["CertFile"].(string)
	if certFile != "" {
		wsProto = "wss"
	}

	//从or_service_node表读取gate节点的IP和Port
	serviceNode, err := models.GetServiceNode()
	if err == nil {
		ip = serviceNode.Ip
		port = strconv.Itoa(int(serviceNode.Port))
	}


	wsAddr := fmt.Sprintf("%s://%s:%s", wsProto, ip, port)
	fmt.Printf("HttpGateModule getWsAddr=%s\n", wsAddr)
	return wsAddr
}


func (h *MerchantApiModule) getHttpAddr() string {
	//gameUrl := "http://192.168.0.116:8803/login"

	httpProto := "http"	//默认值
	ip := "t2.ds16688.com" //默认值
	port := "8803" //默认值

	certFile := h.serviceConfig["CertFile"].(string)
	if certFile != "" {
		httpProto = "https"
	}

	//从or_service_node表读取gate节点的IP和Port
	serviceNode, err := models.GetServiceNode()
	if err == nil {
		ip = serviceNode.Ip
	}


	httpAddr := fmt.Sprintf("%s://%s:%s/login", httpProto, ip, port)
	fmt.Printf("MerchantApiModule getHttpAddr=%s\n", httpAddr)
	return httpAddr
}

// 检测帐号密码是否合法，注意这个函数取自HttpGateService()，如果两边有改变记得保持一致!
func (h *MerchantApiModule) CheckUserValid(userId string, password string) bool {

	hashKey := "user:" + userId

	//先从redis读取，看是否存在
	bytes, err := variable.RedisStorage.HGet(redisId.RedisHashId_PlayersKey, hashKey)

	var usInfo redisId.RedisUserInfo

	if err == nil {
		//若有存在，则解码
		fmt.Printf("redis存在用户,解码redis的值\n")
		err = json.Unmarshal([]byte(bytes), &usInfo)

		if err != nil {
			return false
		}

		//检测帐号密码是正确
		if util.BcryptCheck(password, usInfo.Password) == false {
			return false
		}
		return true
	}

	//若不存在，则从数据库 读取
	playerInfo, err1 := models.GetPlayerByUid(userId)

	if err1 != nil {

		log.Errorf(" error:%s " + err1.Error())

		return false
	}

	//读取商户的数据，以后需要优化从redis读取
	//merchatNode, err2 := models.GetMerchantGroup(playerInfo.MerchantId)

	merchatNode, ok := player.GetMerchantNode(playerInfo.MerchantId)

	if ok == false {
		return false
	}

	//一定要检查，不然出大事
	usInfo.UserId = playerInfo.Username
	usInfo.UserName = playerInfo.Nickname
	usInfo.Password = playerInfo.Password
	usInfo.UserType = playerInfo.PlayerType
	usInfo.MoneyType = playerInfo.MoneyType
	usInfo.MaxWin = playerInfo.MaxWin
	usInfo.LoginIpFirst = playerInfo.LoginIpFirst

	usInfo.ChipsList = playerInfo.Chips

	usInfo.MerchantId = playerInfo.MerchantId
	usInfo.TopMerchantId = merchatNode.TopMerchantId
	usInfo.MerchantCode = merchatNode.MerchantCode
	usInfo.RootKey = merchatNode.GetRootKey()

	usInfo.BetLimitTemplateId = playerInfo.TemplateId

	if playerInfo.Status == 1 {
		usInfo.Blocked = false
	} else {
		usInfo.Blocked = true
	}
	if playerInfo.BetPaused == 1 {
		usInfo.Paused = true
	} else {
		usInfo.Paused = false
	}

	//检测玩家是否第一次登陆
	if playerInfo.LoginIpFirst == "" {
		//是否需要在这边修改第一次登陆的IP
	}

	//检测帐号密码是正确
	if util.BcryptCheck(password, usInfo.Password) == false {
		return false
	}

	// 将用户结构体转换为 Redis 存储格式
	userData, err2 := json.Marshal(&usInfo)

	if err2 != nil {
		return false
	}

	//帐号信息存到redis
	err2 = variable.RedisStorage.HSet(redisId.RedisHashId_PlayersKey, hashKey, string(userData))

	if err2 != nil {
		return false
	}

	return true
}

// 更新玩家信息，来自商户的原始请求结构,MUpdatePlayerApiReq有别于rpc.UpdatePlayerApiReq
type MUpdatePlayerApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`

	Nickname		*string  `json:"nickname"`
	Email			*string  `json:"email"`
	Phone			*string  `json:"phone"`
	Password		*string  `json:"password"`
	BetPaused		*int32   `json:"be_paused"`
}

// 创建游戏账号(返回操作状态码，不进入游戏，不返回游戏地址)
func (h *MerchantApiModule) UpdatePlayerInfo(c *gin.Context) {
	var req MUpdatePlayerApiReq

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	// 检查必传参数
	if req.MerchantCode == "" || req.Username == "" {
		ReturnHttpApi(c, 10001, "缺少商户码和用户名", "")
		return
	}

	// 判断前端是否传递了可选字段
	hasNickname := req.Nickname != nil
	hasEmail := req.Email != nil
	hasPhone := req.Phone != nil
	hasPassword := req.Password != nil
	hasBePaused := req.BetPaused != nil

	log.Infof("前端传递的字段空与否: nickname: %v, email: %v, phone: %v, password: %v, be_paused: %v",
		hasNickname, hasEmail, hasPhone, hasPassword, hasBePaused)

	var rpcReq rpc.UpdatePlayerApiReq
	var rpcRet rpc.UpdatePlayerApiRet

	rpcReq.MerchantCode = req.MerchantCode
	rpcReq.Username = req.Username
	rpcReq.UpdateKeys = make([]string, 0)

	// 检查并添加传递的字段
	if req.Nickname != nil {
	    rpcReq.Nickname = *req.Nickname
	    rpcReq.UpdateKeys = append(rpcReq.UpdateKeys, "nickname")
	}
	if req.Email != nil {
	    rpcReq.Email = *req.Email
	    rpcReq.UpdateKeys = append(rpcReq.UpdateKeys, "email")
	}
	if req.Phone != nil {
	    rpcReq.Phone = *req.Phone
	    rpcReq.UpdateKeys = append(rpcReq.UpdateKeys, "phone")
	}
	if req.Password != nil {
	    rpcReq.Password = *req.Password
	    rpcReq.UpdateKeys = append(rpcReq.UpdateKeys, "password")
	}
	if req.BetPaused != nil {
	    rpcReq.BetPaused = *req.BetPaused
	    rpcReq.UpdateKeys = append(rpcReq.UpdateKeys, "be_paused")
	}

	log.Infof("准备RPC请求: %+v\n", rpcReq)
	err := h.Call("CrmRpcService.RPCUpdatePlayer", &rpcReq, &rpcRet)
	if err != nil {
		fmt.Printf("CrmRpcService.RPCUpdatePlayer ************************************** Call error: %+v\n", err)
	}

	fmt.Printf("RPCUpdatePlayer ret is %v\n", rpcRet)

	if 	rpcRet.Code == 200 {//删除Redis里的原账号信息
		hashKey := fmt.Sprintf("user:%s_%s", strings.ToLower(req.MerchantCode), strings.ToLower(req.Username))

		fmt.Printf("111111111, key=%s, hashKey=%s\n", redisId.RedisHashId_PlayersKey, hashKey)
		_, err := variable.RedisStorage.HDel(redisId.RedisHashId_PlayersKey, hashKey)

		if err != nil {
			fmt.Printf("移除Redis用户缓存信息失败!\n")
		}
	}

	ReturnHttpApi(c, rpcRet.Code, rpcRet.Msg, "")
	return
}


// 定义检查函数
func validateCreatePlayerRequest(req *rpc.CreatePlayerApiReq, merchantPrefix string) error {
	// 1. 检查 LoginName（游戏账号）
	if err := validateLoginName(req.Username, merchantPrefix); err != nil {
		return err
	}

	// 2. 检查 LoginPassword（登录密码）
	if err := validateLoginPassword(req.Password); err != nil {
		return err
	}

	// 3. 检查 Nickname（昵称），如果没有传入 Nickname，自动生成
	if err := validateNickname(req, merchantPrefix); err != nil {
		return err
	}

	return nil
}

// 校验 LoginName（游戏账号）
func validateLoginName(loginName, merchantPrefix string) error {
	// 转换为小写
	loginName = strings.ToLower(loginName)

	// 检查长度
	if len(loginName) < 6 || len(loginName) > 50 {
		return fmt.Errorf("loginName 必须是 6 到 50 个字符")
	}

	// 检查是否包含商户前缀，都转成小写对比
	if !strings.HasPrefix(loginName, merchantPrefix) {
		return fmt.Errorf("loginName 必须包含商户前缀: %s", merchantPrefix)
	}

	// 检查特殊字符：只允许 _ @ # & *
	validLoginNamePattern := `^[a-z0-9_@#&*]+$`
	if matched, _ := regexp.MatchString(validLoginNamePattern, loginName); !matched {
		return fmt.Errorf("loginName 只允许包含小写字母、数字和特殊字符 [_@#&*]")
	}

	return nil
}

// 校验 LoginPassword（登录密码）
func validateLoginPassword(loginPassword string) error {
	// 检查长度
	if len(loginPassword) < 6 || len(loginPassword) > 32 {
		return fmt.Errorf("loginPassword 必须是 6 到 32 个字符")
	}

	// 检查不允许的特殊符号
	invalidPasswordPattern := `['"\\\[\]./‘”“’$]`
	if matched, _ := regexp.MatchString(invalidPasswordPattern, loginPassword); matched {
		return fmt.Errorf("loginPassword 包含不允许的符号")
	}

	return nil
}

// 校验 Nickname（昵称）
func validateNickname(req *rpc.CreatePlayerApiReq, merchantPrefix string) error {
	// 如果没有传入 Nickname，自动生成
	if req.Nickname == "" {
		// 截取 LoginName 中去除商户前缀后的前 12 个字符作为昵称
		withoutPrefix := strings.TrimPrefix(req.Username, merchantPrefix)
		if len(withoutPrefix) > 12 {
			withoutPrefix = withoutPrefix[:12]
		}
		req.Nickname = strings.ToLower(withoutPrefix) // 转小写
	}

	// 检查昵称长度
	fmt.Printf("req.Nickname=%s, len=%d\n", req.Nickname, len(req.Nickname))
	if len(req.Nickname) > 12 {
		return fmt.Errorf("nickName 最多只能包含 12 个字符")
	}

	// 检查昵称只允许字母、数字、下划线和 @ 符号
	validNicknamePattern := `^[a-z0-9_@]+$`
	if matched, _ := regexp.MatchString(validNicknamePattern, req.Nickname); !matched {
		return fmt.Errorf("nickName 只允许包含小写字母、数字、下划线和 @ 符号")
	}

	// 确保昵称是小写
	if req.Nickname != strings.ToLower(req.Nickname) {
		return fmt.Errorf("nickName 必须是小写字符")
	}

	return nil
}

func getRetUrlBygameTypeId(gameTypeId int) (string, error) {
	//不传入gameTypeId参数对于首次登陆用户会自动进入大厅
	//...

	//不传入gameTypeId参数对于老用户会自动进入大厅
	//...

	//传入gameTypeId参数对于首次登陆用户会自动进入该游戏下的随机桌台(游戏逻辑决定进入哪个桌台)
	//...

	//传入gameTypeId参数对于老用户会自动进入该游戏下的随机桌台(游戏逻辑决定进入哪个桌台)
	//...

	//对于断线重连场景，自动重连后会员将会被直接拉入最后一次的桌台，直到5局未下注后自动剔除
	//...

	//对于直接关闭浏览器未能正常手动退出的场景，将会直接拉入最后一次的桌台，直到5局未下注后自动剔除
	//...

	demoUrl := "http://xxooh5.com/?token=eyJhb42" //先返回demo值，待游戏架构完成后，返回真实的Url地址
	return demoUrl, nil
}