package httpgateservice

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
	"strconv"
	"strings"
	"time"
	"videogame/common/collect"
	"videogame/common/global/variable"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	redisId "videogame/common/redis"
	"videogame/common/util"
	"videogame/service/games/player"
)

// login 登录接口
func (h *HttpGateModule) login(c *gin.Context) {

	var req msg.HttpLoginReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		log.SError("绑定请求数据失败:", err.Error())
		util.Fail(c, util.ParamError)
		return
	}

	// 检查ip，中国和香港地区不能登录！
	//clientIP := c.ClientIP()
	//if util.CheckIpValid(clientIP) == false {
	//	log.Infof("IP:%s Login from China or Hong Kong is not allowed!", clientIP)
	//	msgStr := fmt.Sprintf("IP:%s Login from China or Hong Kong is not allowed!", clientIP)
	//	util.Fail(c, msgStr)
	//	return
	//}

	userId := strings.ToLower(req.UserName)

	//先到redis检测此存号
	if h.CheckUserValid(userId, req.PassWord) == false {

		log.SError("Account Error ")

		util.Fail(c, "Account Error ")

		return
	}

	token, err := util.GenerateUserToken(userId)

	if err != nil {
		log.SError("生成用户token失败, err:%v", err)
		util.Fail(c, util.SystemError)
		return
	}

	var enterLobby msg_struct.LoginEnterLobby

	enterLobby.Token = token //5e2809c4c978c57beb1460201bb2bedc
	enterLobby.UserId = req.UserName
	enterLobby.WsAddr = h.GetBestGateUrl()
	enterLobby.Language = "en"
	enterLobby.DeviceType = 1

	enterLobby.DirectIntoGame = false
	enterLobby.GroupId = ""
	enterLobby.GameId = ""
	enterLobby.DeskNo = 0
	enterLobby.BackUrl = ""

	c.JSON(http.StatusOK, gin.H{
		"Code": msg.ErrCode_Ok,
		"Msg":  "ok",
		"Data": enterLobby,
	})

	log.Infof(" ***** UserId:%s TO client :%s", req.UserName, enterLobby.WsAddr)

	return
}

// 获取玩家下注订单列表
func (h *HttpGateModule) betRecords(c *gin.Context) {
	var req msg.GetBetRecordListReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		log.SError("绑定请求数据失败:", err.Error())
		util.Fail(c, util.ParamError)
		return
	}

	var startTimeStr, endTimeStr string
	var startTime, endTime time.Time

	if req.StartTime != "" {
		var err1 error
		startTime, err1 = time.ParseInLocation(time.DateTime, req.StartTime, time.Local)
		if err1 != nil {
			log.SError("StartTime 时间格式错误:", err1)
			util.Fail(c, "StartTime Invalid time format, must be YYYY-MM-DD HH:MM:SS")
			return
		}
		startTimeStr = startTime.Format(time.DateTime)
	}

	if req.EndTime != "" {
		var err2 error
		endTime, err2 = time.ParseInLocation(time.DateTime, req.EndTime, time.Local)
		if err2 != nil {
			log.SError("EndTime 时间格式错误:", err2)
			util.Fail(c, "EndTime Invalid time format, must be YYYY-MM-DD HH:MM:SS")
			return
		}
		endTimeStr = endTime.Format(time.DateTime)
	}

	// 时间范围校验
	if !endTime.IsZero() {
		if endTime.Before(startTime) {
			util.Fail(c, "EndTime must be after StartTime")
			return
		}
		if endTime.Sub(startTime).Hours() > 24*30 {
			util.Fail(c, "Time range cannot exceed 30 days")
			return
		}
	}

	//解码token
	username, err := util.ValidateUserToken(req.Token)
	if err != nil {
		log.SError("Token is Invalid!:", err.Error())
		util.Fail(c, "Token is Invalid !")
		return
	}

	hashKey := "user:" + username

	//先从redis读取，不存在则提示请先登录
	_, err = variable.RedisStorage.HGet(redisId.RedisHashId_PlayersKey, hashKey)
	if err != nil {
		util.Fail(c, "please login first !")
		return

	}

	betRecord := &models.SBetRecord{}
	list, total, page, pageSize, err := betRecord.GetBetRecordList(username, req.DeskNo, int(req.Page), int(req.PageSize), startTimeStr, endTimeStr)
	if err != nil {
		log.SError("查询投注记录失败:", err.Error())
		util.Fail(c, util.ParamError)
		return
	}

	//后面修改成动态的
	defaultLang := "en"

	//转成proto
	betRecordList := []*msg.BetRecord{}
	for _, l := range list {
		areaName := util.GetAreaNameByLangAndAreaId(defaultLang, l.BetAreaId)
		betRecordList = append(betRecordList, l.ToBetRecordProto(areaName))
	}

	// 计算总页数，注意用 realPageSize
	totalPage := int32((total + int64(pageSize) - 1) / int64(pageSize)) // 向上取整

	// 整理返回数据
	ret := msg.GetBetRecordListRet{
		BetRecordList: betRecordList,           //投注明细列表
		Total:         gconv.String(total),     //总条数
		TotalPage:     gconv.String(totalPage), //总页数
		Page:          gconv.String(page),      //当前页
		PageSize:      gconv.String(pageSize),  //每页条数
	}

	util.Success(c, &ret)

}

// userWinLostAccount 玩家输赢统计
func (h *HttpGateModule) userWinLostAccount(c *gin.Context) {
	var req msg.UserWinLostAccountReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		log.SError("绑定请求数据失败:", err.Error())
		util.Fail(c, util.ParamError)
		return
	}

	//解码token
	username, err := util.ValidateUserToken(req.Token)
	if err != nil {
		log.SError("Token is Invalid!:", err.Error())
		util.Fail(c, "Token is Invalid !")
		return
	}

	hashKey := "user:" + username

	//先从redis读取，不存在则提示请先登录
	_, err = variable.RedisStorage.HGet(redisId.RedisHashId_PlayersKey, hashKey)
	if err != nil {
		util.Fail(c, "please login first !")
		return

	}

	var startTime, endTime time.Time

	if req.StartTime != "" {
		var err1 error
		startTime, err1 = time.ParseInLocation(time.DateTime, req.StartTime, time.Local)
		if err1 != nil {
			log.SError("StartTime 时间格式错误:", err1)
			util.Fail(c, "StartTime Invalid time format, must be YYYY-MM-DD HH:MM:SS")
			return
		}
	}

	if req.EndTime != "" {
		var err2 error
		endTime, err2 = time.ParseInLocation(time.DateTime, req.EndTime, time.Local)
		if err2 != nil {
			log.SError("EndTime 时间格式错误:", err2)
			util.Fail(c, "EndTime Invalid time format, must be YYYY-MM-DD HH:MM:SS")
			return
		}
	}

	// 时间范围校验
	if !endTime.IsZero() {
		if endTime.Before(startTime) {
			util.Fail(c, "EndTime must be after StartTime")
			return
		}
		if endTime.Sub(startTime).Hours() > 24*30 {
			util.Fail(c, "Time range cannot exceed 30 days")
			return
		}
	}
	// Time: 当前时间
	now := time.Now()
	// Time: 今天的零点 2025-04-09 00:00:00.000
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 只查询今天
	onlyToday := !startTime.Before(todayStart) && endTime.After(todayStart)
	// 只查询之前
	onlyBefore := endTime.Before(todayStart)
	// 查询今天和之前
	mixRange := startTime.Before(todayStart) && endTime.After(todayStart)

	switch {
	case onlyToday:
		h.queryToday(c, username, startTime, endTime)
	case onlyBefore:
		h.queryBefore(c, username, startTime, endTime)
	case mixRange:
		h.queryRange(c, username, startTime, endTime)
	default:
		util.Fail(c, "invalid time param")
	}
}

func (h *HttpGateModule) getWsAddr() string {
	//wsArr := "ws://192.168.0.112:9444"

	wsProto := "ws"       //默认值
	ip := "192.168.0.112" //默认值
	port := "9444"        //默认值

	certFile := h.serviceConfig["CertFile"].(string)
	if certFile != "" {
		wsProto = "wss"
	}

	//从or_service_node表读取gate节点的IP和Port，如果有值则替代默认值
	serviceNode, err := models.GetServiceNode()
	if err == nil {
		ip = serviceNode.Ip
		port = strconv.Itoa(int(serviceNode.Port))
	}

	wsAddr := fmt.Sprintf("%s://%s:%s", wsProto, ip, port)
	fmt.Printf("HttpGateModule getWsAddr=%s\n", wsAddr)
	return wsAddr
}

// 检测帐号是否合法
func (h *HttpGateModule) CheckUserValid(userId string, password string) bool {

	hashKey := "user:" + userId

	//先从redis读取，看是否存在
	bytes, err := variable.RedisStorage.HGet(redisId.RedisHashId_PlayersKey, hashKey)

	var usInfo redisId.RedisUserInfo

	if err == nil {

		//若有存在，则解码
		err = json.Unmarshal([]byte(bytes), &usInfo)

		//若解码正确则检测密码
		if err == nil {

			nowTime := time.Now().Unix()

			//检测更新的时间是否小于2小时
			if nowTime-usInfo.UpdateTime < 3600*2 {

				//检测帐号密码是正确
				if util.BcryptCheck(password, usInfo.Password) == false {
					//写入登陆日志，密码错
					return false
				}
				return true
			}
		}
	}

	//若不存在，或者redis之前解码错误 ，则从数据库 读取
	playerInfo, err1 := models.GetPlayerByUid(userId)

	if err1 != nil {

		log.Errorf(" read db playerInfo  error:%s " + err1.Error())

		return false
	}

	//检测帐号是否禁用
	if playerInfo.Status == 0 {

		//写入陆录日志错误原因，帐号禁用

		return false
	}

	//读取商户的数据，以后需要优化从redis读取
	merchatNode, ok := player.GetMerchantNode(playerInfo.MerchantId)

	if ok == false {

		//重新去读取这个商户
		merchatNode, ok = player.BuildMerchantNode(playerInfo.MerchantId)

		if ok == false {

			log.Errorf("无法建立商户Node:%s", playerInfo.MerchantId)

			return false
		}
	}

	//检测商户是否被禁用
	if merchatNode.Blocked == true {
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
	usInfo.UpdateTime = time.Now().Unix()

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

	//寻找最小，最大打赏金额,2025.5.29 add
	topNode, ok := player.GetMerchantNode(usInfo.TopMerchantId)

	if ok == true && topNode != nil {

		for key := range topNode.MerchantCurrencys {

			currentNode, ok := topNode.MerchantCurrencys[key]

			if ok == true {
				usInfo.MinTip = currentNode.MinTip
				usInfo.MaxTip = currentNode.MaxTip
				usInfo.Digits = currentNode.Digits

				break
			}
		}
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

func (h *HttpGateModule) chatList(c *gin.Context) {
	var req msg.ChatlistReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		log.SError("绑定请求数据失败:", err.Error())
		util.Fail(c, util.ParamError)
		return
	}

	//解码token
	username, err := util.ValidateUserToken(req.Token)
	if err != nil {
		log.SError("Token is Invalid!:", err.Error())
		util.Fail(c, "Token is Invalid !")
		return
	}

	hashKey := "user:" + username

	//先从redis读取，不存在则提示请先登录
	usInfoBytes, err := variable.RedisStorage.HGet(redisId.RedisHashId_PlayersKey, hashKey)
	if err != nil {
		util.Fail(c, "please login first !")
		return
	}

	//用户信息
	var usInfo redisId.RedisUserInfo
	err = json.Unmarshal([]byte(usInfoBytes), &usInfo)
	if err != nil {
		util.Fail(c, util.SystemError)
		return
	}

	//主商户ID
	var merchantId string

	//获取商户信息
	merchant, err := models.GetMerchantGroup(usInfo.MerchantId)
	if err != nil {
		util.Fail(c, util.SystemError)
		return
	}

	if merchant.ParentId != "root" { //是子商户，则找出主商户ID
		merchantId = merchant.ParentId
	} else {
		merchantId = merchant.Username
	}

	//根据merchanId获取公告,可能有多条，一并返回给前端
	merchantNotices, err := models.GetMerchantNotice(merchantId)
	if err != nil {
		util.Fail(c, util.ParamError)
		return
	}

	chatlist := []*msg.TalkChat{}

	for _, n := range merchantNotices {

		noticeMsg := msg.TalkChat{
			Message: n.Content,
			MsgType: 1, // 消息类型 0：聊天内容 1：商户公告
		}

		chatlist = append(chatlist, &noticeMsg)
	}

	chat := &collect.TalkChat{}

	pageNum := int64(1)
	pageSize := int64(10) //默认取10条

	list, err := chat.GetChatListByGameIdAndDeskNo(req.GameId, req.DeskNo, pageNum, pageSize)

	for _, l := range list {

		unused1 := 0

		if l.PlayerId == username {
			unused1 = 1
		}

		bytes := []byte(l.PlayerId)
		size := len(l.PlayerId)

		//把玩家字符串尾数填”*“
		if size < 5 {
			for i := size - 1; i > size-2; i-- {
				bytes[i] = '*'
			}
		} else {
			for i := size - 1; i > size-5; i-- {
				bytes[i] = '*'
			}
		}

		newId := string(bytes)

		talkChatMsg := msg.TalkChat{
			GameId:   l.GameId,
			Deskno:   l.Deskno,
			Message:  l.Message,
			Unused1:  int32(unused1),
			PlayerId: newId,
			CreateAt: l.CreateAt,
		}

		chatlist = append(chatlist, &talkChatMsg)
	}

	ret := msg.ChatlistRet{
		List: chatlist,
	}

	util.Success(c, &ret)

}
