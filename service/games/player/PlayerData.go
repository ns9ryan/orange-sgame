package player

import (
	"strconv"
	"strings"
	"time"
	"videogame/common/model"
)

type PoolObj struct {
	ref bool
}

// DataInfo 用于存于不需要持久化的数据，需要序列化的数据首字母大写
type DataInfo struct {
	FromGateId string //关联网关nodeId
	ClientId   string //clientId

	isOnline bool //是否在线状态

	IsRoomUser bool //是否在房间内
	IsInLobby  bool //是否在大厅

	NoBetRound int //连续没有下注的局数

	LoginTime   time.Time //登陆的时间
	LastBetTime time.Time //最后投注时间

	WalletData  PlayerWallet  //玩家钱包
	AccountData PlayerAccount //玩家帐目

	SettingData UserSetting //玩家的配置

}

// =============================================================================================
func (po *PoolObj) IsRef() bool {
	return po.ref
}

func (po *PoolObj) Ref() {
	po.ref = true
}

func (po *PoolObj) UnRef() {
	po.ref = false
}

func (po *PoolObj) Reset() {
	po.ref = false
}

// ==============================================================================================
func (df *DataInfo) GetClientId() string {
	return df.ClientId
}

func (df *DataInfo) GetGateNodeId() string {
	return df.FromGateId
}

func (df *DataInfo) SetOnline(onLine bool) {
	df.isOnline = onLine
}

func (df *DataInfo) GetOnline() bool {
	return df.isOnline
}

// 更新登陆时间
func (df *DataInfo) UpdateLoginTime() {
	df.LoginTime = time.Now()
}

func (df *DataInfo) AttachConn(fromGateNodeId string, clientId string, sessionId string, ip string, isroom bool) {

	df.FromGateId = fromGateNodeId

	ipStr := strings.Split(ip, ":")

	if len(ipStr) > 0 {
		ip = ipStr[0]
	}

	df.ClientId = clientId
	if clientId == "" {
		df.isOnline = false
	}

	df.IsRoomUser = isroom
	df.SettingData.SuperVisible = true

	//没有下注的局数置0
	df.ResetNoBetRound()
}

// 设定在房间内
func (df *DataInfo) SetRoomIn(roomIn bool) {

	df.IsRoomUser = roomIn

}

// 检测是否房间内
func (df *DataInfo) IsRoomIn() bool {

	return df.IsRoomUser
}

// 设定在房间内
func (df *DataInfo) SetLobbyIn(lobbyIn bool) {

	df.IsInLobby = lobbyIn

}

// 检测是否房间内
func (df *DataInfo) IsLobbyIn() bool {

	return df.IsInLobby
}

// 是否启用SuperSix状态
func (user *DataInfo) IsSuperSixOpt() bool {
	return user.SettingData.SuperSixOpt
}

// 设置是否可以操作SuperSix
func (user *DataInfo) SetSuperSixOpt(opt bool) bool {

	user.SettingData.SuperSixOpt = opt

	return user.SettingData.SuperSixOpt
}

// 读取supersix按钮是否可见
func (user *DataInfo) GetSuperVisible() bool {

	return user.SettingData.SuperVisible
}

// 读取supersix状态
func (user *DataInfo) GetSuperSix() int {

	return user.SettingData.SuperSix
}

// 设置supersix 状态
func (user *DataInfo) SetSuperSix(status int) int {

	user.SettingData.SuperSix = status

	return user.SettingData.SuperSix
}

// 读取限红模板配置信息
func (user *DataInfo) GetBetLimit(templateId string, gameId string) ([]*BetAreaU, error) {

	var templateTable models.SMerchantTemplateBetlimitArea

	items, err := templateTable.GetMerchantTemplateData(templateId, gameId)

	if err != nil {
		return nil, err
	}

	size := len(items)
	digits := 1

	betAreas := make([]*BetAreaU, size)

	minbet := 99999999
	maxbet := 0

	for i := 0; i < size; i++ {

		digits = int(items[i].Digits)

		if digits == 0 {
			digits = 1
		}

		betid, _ := strconv.Atoi(items[i].AreaId)
		lowbet := float64(items[i].MinBetAmount) / float64(digits)
		highbet := float64(items[i].MaxBetAmount) / float64(digits)

		betAreas[i] = &BetAreaU{
			AreaId:  uint32(betid),
			LowBet:  uint32(items[i].MinBetAmount),
			HighBet: uint32(items[i].MaxBetAmount),
		}

		user.SettingData.BetSet.BetAreas[betid] = BetAreaF{
			AreaId:  uint32(betid),
			LowBet:  lowbet,
			HighBet: highbet,
		}

		//找出最小下注额
		if minbet > int(items[i].MinBetAmount) {
			minbet = int(items[i].MinBetAmount)
		}
		//找出最大下注额
		if maxbet < int(items[i].MaxBetAmount) {
			maxbet = int(items[i].MaxBetAmount)
		}
	}

	user.SettingData.BetSet.Digits = digits
	user.SettingData.BetSet.LowBet = minbet
	user.SettingData.BetSet.HighBet = maxbet

	return betAreas, nil

}

// 没有下注的局数清0
func (user *DataInfo) ResetNoBetRound() {
	user.NoBetRound = 0
}

// 没有下注的局数加1
func (user *DataInfo) NoBetRoundInc() {
	user.NoBetRound++
}

func (user *DataInfo) IsOverNoBetRound(maxround int) bool {

	//0表示不限
	if maxround == 0 {
		return false
	}

	if user.NoBetRound >= maxround {
		return true
	}

	return false

}
