//*****************************************************************
// OBLiveApi-3-Maintenance.go里，放的是API文档里描述的基础接口(目前一共10个)，包括：
// 创建账号、快速开始游戏、上分下分、查询余额等供商户调用的API接口。
//*****************************************************************

package apimodule

import (
	//"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	models "videogame/common/model"
	"strconv"
)

// 校验API服务器接口是否正常，可用于定期心跳检测
func (h *MerchantApiModule) CheckApiOk(c *gin.Context) {
	//判断api服务或者游戏服务处于维护状态
	//ReturnHttpApi(c, 10001, "服务维护中", "")

	ReturnHttpApi(c, 200, "成功", "")
}

type GameInfo struct {
	GameTypeId 		    string  `json:"game_type_id"`
	GameTypeName 		string  `json:"game_type_name"`
}

// 获取游戏分类目录，可用于商户方的面板展示
func (h *MerchantApiModule) GetGameCatalog(c *gin.Context) {
	var gameList = make([]GameInfo, 0)

	games, err := models.GetGameCatalog()
	if err != nil {
		ReturnHttpApi(c, 10001, "获取游戏分类失败", "")
		return
	}

	for _, game := range games {
		var gameInfo GameInfo

		gameInfo.GameTypeId = game.SubCode
		gameInfo.GameTypeName = game.Description

		gameList = append(gameList, gameInfo)
	}

	ReturnHttpApi(c, 200, "成功", gameList)
	return
}

type GameRuleInfo struct {
	BetAreaCode 		string  `json:"bet_area_code"`
	BetAreaName 		string  `json:"bet_area_name"`
	BetAreaComment 		string  `json:"bet_area_comment"`
}

// 获取游戏分类目录，可用于商户方的面板展示
func (h *MerchantApiModule) GetGameRules(c *gin.Context) {
	gameSubCode := c.DefaultQuery("game_type_id", "")
	if gameSubCode == "" {
		ReturnHttpApi(c, 10001, "未传game_type_id", "")
		return
	}

	var gameRuleList = make([]GameRuleInfo, 0)
	game, err := models.GetGameCatalogById(gameSubCode)
	if err != nil {
		ReturnHttpApi(c, 10001, "获取游戏分类失败", "")
		return
	}	

	gameRules, err := models.GetGameRules(game.GameCategoryCode)
	if err != nil {
		ReturnHttpApi(c, 10001, "获取游戏玩法列表失败", "")
		return
	}

	for _, gameRule := range gameRules {
		var gameRuleInfo GameRuleInfo

		gameRuleInfo.BetAreaCode = gameRule.AreaKey
		gameRuleInfo.BetAreaName = gameRule.AreaName
		gameRuleInfo.BetAreaComment = gameRule.Description

		gameRuleList = append(gameRuleList, gameRuleInfo)
	}

	ReturnHttpApi(c, 200, "成功", gameRuleList)
	return
}

type GameTableInfo struct {
	GameTableId 		int  	`json:"game_table_id"`
	GameTableName 		string  `json:"game_table_name"`
	Disable 			int  	`json:"disable"`
}

// 获取游戏分类目录，可用于商户方的面板展示
func (h *MerchantApiModule) GetGameTables(c *gin.Context) {
	gameSubCode := c.DefaultQuery("game_type_id", "")
	if gameSubCode == "" {
		ReturnHttpApi(c, 10001, "未传game_type_id", "")
		return
	}

	var gameTableList = make([]GameTableInfo, 0)

	gameTables, err := models.GetGameTables(gameSubCode)
	if err != nil {
		ReturnHttpApi(c, 10001, "获取游戏桌台列表失败", "")
		return
	}

	for _, gameTable := range gameTables {
		var gameTableInfo GameTableInfo

		gameTableInfo.GameTableId = int(gameTable.DeskNo)
		gameTableInfo.GameTableName = gameTable.Description
		gameTableInfo.Disable = int(1 - gameTable.Status) //数据status 1表示开启

		gameTableList = append(gameTableList, gameTableInfo)
	}

	ReturnHttpApi(c, 200, "成功", gameTableList)
	return
}

type OnlinePlayer struct {
	Username		string		`json:"username"`
	Code			string		`json:"currency"`
	LoginTime		string		`json:"login_time"`
	LoginIp			string		`json:"login_ip"`
	Balance 		FixedOutput 	`json:"balance"`
}

// OnlineUsersRequest 查询商户的在线会员列表
type MOnlinePlayersApiRet struct {
	PageIndex		int		`json:"page_index"`
	PageSize		int		`json:"page_size"`
	TotalRecord		int		`json:"total_record"`
	TotalPage		int		`json:"total_page"`

	Records			[]OnlinePlayer 	`json:"records"`
}

// OnlineUsers 查询在线会员列表
func (h *MerchantApiModule) OnlinePlayers(c *gin.Context) {
	var ret MOnlinePlayersApiRet

	merchantCode := c.GetHeader("MerchantCode")
	if merchantCode == "" {
		log.Error("Header里没有传MerchantCode")
		ReturnHttpApi(c, 10001, "Header里没有传MerchantCode", "")
		return
	}

	pageIndexStr := c.DefaultQuery("page_index", "")
	if pageIndexStr == "" {
	    ReturnHttpApi(c, 10001, "未传page_index", "")
	    return
	}

	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil {
	    ReturnHttpApi(c, 10002, "page_index 不是有效的整数", "")
	    return
	}

	code := c.DefaultQuery("code", "")
	if code == "" {
		fmt.Printf("注意查询币种code为空")
	}

	merchant, err := models.GetMerchantByCode(merchantCode)
	if err != nil {
		ReturnHttpApi(c, 10001, "没有找到商户信息", "")
		return
	}

	players, err := models.GetOnlinePlayers(merchant.Username, code, pageIndex)
	if err != nil {
		ReturnHttpApi(c, 10001, "没有找到玩家信息", "")
		return
	}

	var records = make([]OnlinePlayer, 0)
	for _, player := range players {
		var record OnlinePlayer
		record.Username = player.Username
		record.Code = player.Currency

		var loginTimeStr string
		if player.LoginTime != nil {
		    loginTimeStr = player.LoginTime.Format("2006-01-02T15:04:05-07:00")
		} else {
		    loginTimeStr = "" // 处理 nil 值，避免 panic
		}

		record.LoginTime = loginTimeStr
		record.LoginIp = player.LoginIp

		accountInfo, err := models.GetUserWalletByUid(player.Username)
		if err != nil {
			ReturnHttpApi(c, 10001, "没有找到用户!", "")
			return
		}

		fmt.Printf("accountInfo.Balance=%d, digits=%d", accountInfo.Balance, accountInfo.Digits)
		balance := float64(accountInfo.Balance) /  float64(accountInfo.Digits)
		record.Balance = WrapFixedFloat(balance, accountInfo.Currency)

		records = append(records, record)
	}

	ret.PageIndex = pageIndex
	ret.PageSize = 500
	ret.TotalRecord = len(records)
	ret.TotalPage = (len(records) + 500 - 1) / 500
	ret.Records = records

	ReturnHttpApi(c, 200, "", ret)
}
