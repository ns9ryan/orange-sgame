package player

import (
	"encoding/json"
	"github.com/duanhf2012/origin/v2/log"
	"strconv"
	"strings"
	"time"
	"videogame/common/global/variable"
	models "videogame/common/model"
	redisId "videogame/common/redis"
	"videogame/common/systemdata"
)

//const MaxAreaNum = 300 //全部的投注区域

// 投注区域结构体，主要用于浮点数
type BetAreaF struct {
	AreaId  uint32  //投注区域Id
	LowBet  float64 //低注
	HighBet float64 //高注
}

// 主要用于整数用
type BetAreaU struct {
	AreaId  uint32 //投注区域Id
	LowBet  uint32 //低注
	HighBet uint32 //高注
	Digits  uint32
}

// 打赏的数据
type TipData struct {
	MinTip    uint //打赏的最小金额
	MaxTip    uint //打赏的最大金额
	TipDigits uint //打赏金额的小数位

	TipItems []int32 //打赏金额的列表
}

const (
	SingleWallet = 1 //单钱包
	MultWallet   = 0 //多钱包
)

// UserInfo 用户信息
type UserInfo struct {
	Id       string //vg_users表字段：id
	UserName string

	Password string //密码

	Email string
	Phone string

	UserId string
	Avatar string

	ParentId     string //上级商户的ID
	MerchantId   string //商户的ID
	MerchantCode string //商户的Code
	RootKey      uint   //顶级商户的Key

	Paused  bool //是否暂停
	Blocked bool //是否被锁

	PlayerType uint //玩家的类型 0: 现金网帐户 1: 代理玩家
	MoneyType  uint //0 正式玩家  1:试玩帐号

	WalletType uint //钱包内型 2:多钱包  1：单钱包

	MaxWin uint //最大盈额

	MinTip    uint //打赏的最小金额
	MaxTip    uint //打赏的最大金额
	TipDigits uint //打赏金额的小数位

	Level uint //玩家级别

	LoginIpFirst string //首次进入的IP
	LoginIp      string //最近登陆的IP

	//LoginTimeFirst   time.Time    //首次进入的时间

	ChipsList          string        //此玩家的抽码列表
	BetLimitTemplateId string        //投注区域的限红模板ID
	TemplatePlanId     string        //限红模板ID对应的
	MerchantParent     *MerchantNode //它归属的商户,不是最顶层的商户
}

func (user *UserInfo) GetId() string {
	return user.Id
}

func (user *UserInfo) GetUserID() string {
	return user.UserId
}

func (user *UserInfo) GetMerchantId() string {
	return user.MerchantId
}
func (user *UserInfo) GetMerchantCode() string {
	return user.MerchantCode
}
func (user *UserInfo) GetRootKey() uint {
	return user.RootKey
}

func (user *UserInfo) GetParentId() string {
	return user.ParentId
}

// 检测玩家是否被禁用了
func (user *UserInfo) IsBlocked() bool {

	return user.Blocked
}

// 更新玩家信息到redis
func (user *UserInfo) UpdateUserInfoToRedis() {

	var usInfo redisId.RedisUserInfo

	usInfo.UserId = user.UserId
	usInfo.UserName = user.UserName
	usInfo.Password = user.Password
	usInfo.UserType = user.PlayerType
	usInfo.MoneyType = user.MoneyType
	usInfo.MaxWin = user.MaxWin
	usInfo.Level = user.Level

	usInfo.WalletType = user.WalletType

	usInfo.MerchantId = user.GetParentId()
	usInfo.TopMerchantId = user.GetMerchantId()
	usInfo.MerchantCode = user.GetMerchantCode()
	usInfo.RootKey = user.GetRootKey()

	usInfo.Blocked = user.Blocked
	usInfo.Paused = user.Paused

	usInfo.LoginIpFirst = user.LoginIpFirst
	usInfo.LoginIP = user.LoginIp
	usInfo.BetLimitTemplateId = user.BetLimitTemplateId

	usInfo.ChipsList = user.ChipsList
	usInfo.UpdateTime = time.Now().Unix()

	// 将用户结构体转换为 Redis 存储格式
	userData, err2 := json.Marshal(&usInfo)

	if err2 != nil {
		return
	}

	redisUserKey := "user:" + user.UserId

	//存到redis
	err2 = variable.RedisStorage.HSet(redisId.RedisHashId_PlayersKey, redisUserKey, string(userData))

}

// 从redis读取此玩家的数据,若redis没有, 则从数据库中读取
func (user *UserInfo) ReadUserInfo() bool {

	redisUserKey := "user:" + user.UserId

	//先从redis读取，看是否存在
	bytes, err := variable.RedisStorage.HGet(redisId.RedisHashId_PlayersKey, redisUserKey)

	var usInfo redisId.RedisUserInfo

	if err == nil {

		//若有存在，则解码
		err = json.Unmarshal([]byte(bytes), &usInfo)

		if err != nil {

			//删除redis 信息
			variable.RedisStorage.HDel(redisId.RedisHashId_PlayersKey, redisUserKey)

		} else {
			user.UserId = usInfo.UserId
			user.UserName = usInfo.UserName

			user.Paused = usInfo.Paused
			user.Blocked = usInfo.Blocked
			user.Avatar = usInfo.AvatarUrl
			user.Email = usInfo.Email
			user.Phone = usInfo.Phone
			user.ParentId = usInfo.MerchantId
			user.MerchantId = usInfo.TopMerchantId
			user.MerchantCode = usInfo.MerchantCode
			user.RootKey = usInfo.RootKey
			user.Password = usInfo.Password

			//2025.5.29 coolz add
			user.MinTip = usInfo.MinTip
			user.MaxTip = usInfo.MaxTip
			user.TipDigits = usInfo.Digits

			user.LoginIpFirst = usInfo.LoginIpFirst

			user.ChipsList = usInfo.ChipsList
			user.BetLimitTemplateId = usInfo.BetLimitTemplateId

			merchantNode, ok := GetMerchantNode(usInfo.MerchantId) //读取商户资料

			if ok == true {

				user.MerchantParent = merchantNode
				user.WalletType = merchantNode.WalletType

			} else {

				//重新到数据库读取这个商户的资料
				merchantNode, ok := BuildMerchantNode(usInfo.MerchantId)

				if ok == true {

					user.MerchantParent = merchantNode
					user.WalletType = merchantNode.WalletType

				} else {
					return false
				}

			}

			return true
		}
	}

	//若不存在，则从数据库 读取
	playerInfo, err1 := models.GetPlayerByUid(user.UserId)

	if err1 != nil {
		return false
	}

	merchantNode, ok := GetMerchantNode(playerInfo.MerchantId) //读取商户资聊

	if ok == false { //找不到它的商户，可能是新商户的帐号,重新到数据库再读取商户资料

		log.Errorf("player %s find parentId:%s not found!!", playerInfo.Username, playerInfo.MerchantId)

		merchantNode, ok := BuildMerchantNode(usInfo.MerchantId)

		if ok == true {
			user.MerchantParent = merchantNode //得到商户Node

		} else {
			//去数据库查找此商户
			return false
		}

	} else {
		user.MerchantParent = merchantNode //得到商户Node
	}

	usInfo.UserId = playerInfo.Username
	usInfo.UserName = playerInfo.Nickname
	usInfo.Password = playerInfo.Password
	usInfo.UserType = playerInfo.PlayerType
	usInfo.MoneyType = playerInfo.MoneyType
	usInfo.MaxWin = playerInfo.MaxWin

	usInfo.MerchantId = merchantNode.MerchantId //它的最近商户的ID
	usInfo.TopMerchantId = merchantNode.TopMerchantId
	usInfo.RootKey = playerInfo.RootKey

	usInfo.WalletType = merchantNode.WalletType //钱包类型

	usInfo.ChipsList = playerInfo.Chips
	usInfo.BetLimitTemplateId = playerInfo.TemplateId

	usInfo.MerchantCode = merchantNode.GetMerchantCode()
	usInfo.LoginIpFirst = playerInfo.LoginIpFirst
	usInfo.UpdateTime = time.Now().Unix()

	if playerInfo.Status == 1 {
		usInfo.Blocked = false
	} else {
		usInfo.Blocked = true
	}

	//是否暂时投注
	if playerInfo.BetPaused == 1 {
		usInfo.Paused = true
	} else {
		usInfo.Paused = false
	}

	//寻找最小，最大打赏金额,2025.5.29 add
	topNode, ok := GetMerchantNode(usInfo.TopMerchantId)

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

	// 将用户结构体转换为 Redis 存储格式
	userData, err2 := json.Marshal(&usInfo)

	if err2 != nil {
		return false
	}

	//存到redis
	err2 = variable.RedisStorage.HSet(redisId.RedisHashId_PlayersKey, redisUserKey, string(userData))

	if err2 != nil {
		return false
	}

	user.UserId = usInfo.UserId
	user.UserName = usInfo.UserName
	user.Paused = usInfo.Paused
	user.Blocked = usInfo.Blocked
	user.Paused = usInfo.Paused
	user.Password = usInfo.Password

	user.ParentId = usInfo.MerchantId
	user.MerchantId = usInfo.TopMerchantId
	user.MerchantCode = usInfo.MerchantCode
	user.RootKey = usInfo.RootKey
	user.WalletType = merchantNode.WalletType
	user.LoginIpFirst = usInfo.LoginIpFirst
	user.ChipsList = usInfo.ChipsList
	user.BetLimitTemplateId = usInfo.BetLimitTemplateId

	user.MinTip = usInfo.MinTip
	user.MaxTip = usInfo.MaxTip
	user.TipDigits = usInfo.Digits

	return true
}

// 清除Redis的数据
func (user *UserInfo) RemoveRedisUserInfo() {

	redisUserKey := "user:" + user.UserId

	//先从redis读取，看是否存在
	variable.RedisStorage.HDel(redisId.RedisHashId_PlayersKey, redisUserKey)

}

// 读取抽码的列表
func (user *UserInfo) GetChipsShow(digits uint) []int32 {

	slices := strings.Split(user.ChipsList, ",")

	chipshow := make([]int32, 0)

	numcount := 0

	for _, valstr := range slices {

		chipKey, err := strconv.Atoi(valstr)

		if err != nil {
			continue
		}

		//从系统抽码列表中读取对应的抽码
		item, ok := systemdata.SystemChipsGetChip(uint(chipKey))

		if !ok {
			continue
		}

		if item.ChipDigits == 0 {
			item.ChipDigits = 1
		}

		chipval := item.ChipValue * digits / item.ChipDigits

		chipshow = append(chipshow, int32(chipval))

		numcount++

		//控在6个抽码，后面再改动
		if numcount >= 6 {
			break
		}

	}

	return chipshow
}

func (user *UserInfo) GetTipData() (*TipData, bool) {

	//寻找最小，最大打赏金额,2025.5.29 add
	if user.MerchantParent == nil {
		return nil, false
	}

	//读取最上级商户的开表
	topNode, ok := GetMerchantNode(user.MerchantParent.TopMerchantId)

	if ok == false {
		return nil, false
	}

	//读取打赏金额的列表
	for key := range topNode.MerchantCurrencys {

		//现在只有一个币种，因此只读取第一个
		currentNode, ok := topNode.MerchantCurrencys[key]

		if ok == true {

			tipData := &TipData{}

			tipData.MinTip = currentNode.MinTip
			tipData.MaxTip = currentNode.MaxTip
			tipData.TipDigits = currentNode.Digits
			tipData.TipItems = make([]int32, 0)

			//读取列表
			for _, item := range currentNode.TipItems {
				tipData.TipItems = append(tipData.TipItems, int32(item))
			}

			return tipData, true
		}
	}

	return nil, false
}

// 读取需要调用的钱包模式
func (user *UserInfo) GetWallType() uint {

	if user.MerchantParent == nil {
		return ClusterWallet
	}

	return user.WalletType

}

// 检测是否API远程调用钱包模式
func (user *UserInfo) IsWalletAPIMode() bool {

	return user.GetWallType() == APIWallet
}

// 检测玩家是否首次登陆，若是的话，需要写IP
func (user *UserInfo) FirstLoginWite(ipaddr string) bool {

	if user.LoginIpFirst != "" {
		return false
	}

	user.LoginIpFirst = ipaddr
	user.LoginIp = ipaddr

	var userTable models.SUserPlayer

	err := userTable.WriteLoginIpFirst(user.Id, ipaddr)

	if err != nil {
		return false
	}

	return true
}

// 写玩家IP到数据库
func (user *UserInfo) WriteLoginIp(ipaddr string) bool {

	user.LoginIp = ipaddr

	var userTable models.SUserPlayer

	err := userTable.WriteLoginIp(user.Id, ipaddr)

	if err != nil {
		return false
	}

	return true
}

// 把登陆写入日志
func (user *UserInfo) WriteLoginLog(ipaddr string, isOk uint, reason string) bool {

	var loginTable models.SLoginLog

	nowTime := time.Now()

	ipstr := strings.Split(ipaddr, ":")

	if len(ipstr) == 2 {
		ipaddr = ipstr[0]
	}

	loginTable.LoginIp = ipaddr
	loginTable.LoginTime = &nowTime
	loginTable.Username = user.UserId
	loginTable.UserType = 1 //玩家类型
	loginTable.Reason = reason
	loginTable.IsOk = isOk

	loginTable.MerchantId = user.ParentId
	loginTable.RootKey = user.RootKey

	err := loginTable.Create()

	return err == nil

}

// 写入玩家在线
func (user *UserInfo) WriteOnlineFlag(online int) bool {

	var userTable models.SUserPlayer

	err := userTable.WriteOnline(user.Id, online)

	if err != nil {
		return false
	}

	return true
}
