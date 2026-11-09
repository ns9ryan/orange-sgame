package walletService

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"time"
	models "videogame/common/model"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/player"
)

func (self *CollectService) AccountToTick(tickId uint64, nil interface{}) {

	//fmt.Printf("结算入库。。。。\n")

	//检测是否需要结算入库
	if self.AccountWorkStart() == false {

		duration := time.Now().Sub(self.CheckTime)

		if duration.Minutes() > 1 {

			self.CheckTime = time.Now()

			fmt.Printf("结算入库没有到时间 。。。。%s \n", self.CheckTime.String())

			self.NotifyWalletServiceReadingTable(false)
		}

	} else {

		if self.IsAccountWorking == false {

			fmt.Println("重新要入库存调用！！！")

			self.IsAccountWorking = true

			log.Infof("启动入库协程。。。")

			//暂停检测
			self.CancelTimerId(&self.AccountTickerId)

			//启动协程工作
			go self.AsynAccountToDB()

		}

	}

}

// 通知钱包服正在读取数据，让它不要更新数据
func (self *CollectService) NotifyWalletServiceReadingTable(IsRead bool) bool {

	nodeId := util.GetGameServiceName(util.WalletService)

	rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCSetReadingTableFlag")

	var isReading rpc.Bool
	var retInt rpc.Int

	isReading.Value = IsRead

	//钱包服的余额增加
	err := self.CallNodeWithTimeout(time.Second*5, nodeId, rpcFunc, &isReading, &retInt)

	if err != nil {

		log.Errorf("通知钱包服，帐目要入库不成功！！！")
		return false
	}

	if retInt.Value == 0 {

		log.Errorf("通知钱包服，帐目要入库不成功！！！")
		return false
	}

	log.Infof("通知钱包服，帐目要是否要结算入库：%v  ", IsRead)

	return true

}

func (self *CollectService) AsynAccountToDB() {

	player.ResetReadMerchant()

	//最近结算的时间
	self.AccountTimer.LastAccountTime = time.Now()

	if player.InitMerchantGroups() == false {
		log.Errorf("初始化商户不成功！")
	}

	player.PrintMerchantGroups()

	//初始化商户代理
	//通知钱包服要开始结算了
	self.NotifyWalletServiceReadingTable(true)

	var userTable models.SUserWallet

	//1.先读取有投注的玩家列表
	userPlayers, err := userTable.GetBettedUsers()

	log.Infof("读取结算的玩家。。。。%d", len(userPlayers))

	if err != nil { //有可能没有人有投注

		self.bAccounted = false
		self.IsAccountWorking = false

		//重新启动计时
		self.SafeNewTicker(&self.AccountTickerId, time.Second, nil, self.AccountToTick)

		return
	}

	if len(userPlayers) > 0 {

		//2.玩家投注帐目清0
		userTable.ResetBettedUsers()

		// 玩家打赏清0
		userTable.ResetTipUsers()

		var merchantTable models.MerchantCurrency

		//商户投注帐目也清0
		merchantTable.ResetZeroLastAccount()
		merchantTable.ResetZeroLastTip()

		log.Infof("玩家帐号数据清0,商户帐号数据清0。。。。, 玩家个数:%d ", len(userPlayers))
	}

	//2 通知钱包服已读取表数据了,它们可以正常工作了
	self.NotifyWalletServiceReadingTable(false)

	nowTime := time.Now()

	yeasterday := self.AccountTimer.NowAccountTime //nowTime.AddDate(0, 0, 0 )

	year, month, day := yeasterday.Date()

	//结算日期算前一天的
	totayDay := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	day1 := totayDay
	day2 := nowTime

	log.Infof("玩家帐号入库开始。。。")

	saveNum := 0

	//3.统计每天玩家数据，存入玩家结算表
	for ind, userData := range userPlayers {

		var userAccountTable models.SUserAccount

		//读取此商户的数据
		merchantNode, ok := player.GetMerchantNode(userData.MerchantId)

		if ok == false {

			log.Errorf("读取商户不成功！！！")
			continue //出错了
		}

		userAccountTable.Digits = userData.Digits
		userAccountTable.Username = userData.Username
		userAccountTable.Currency = userData.Currency

		userAccountTable.ParentId = merchantNode.MerchantId
		userAccountTable.MerchantId = merchantNode.TopMerchantId

		userAccountTable.RootKey = merchantNode.MerchantKey

		userAccountTable.LastBet = userData.LastBet
		userAccountTable.LastWater = userData.LastWater
		userAccountTable.LastBetEffective = userData.LastBetEffective
		userAccountTable.LastWinLost = userData.LastWinLost
		userAccountTable.LastTip = userData.LastTip

		userAccountTable.AccountDate = &totayDay
		userAccountTable.StarTime = &day1
		userAccountTable.EndTime = &day2

		//生成帐目数据
		err := userAccountTable.Create()

		if err != nil {
			//存到专门的文件
			//重新再存一次

			log.Errorf("存玩家帐目数据不成功！！！err:%s", err)
			//错误处理
			continue
		}
		saveNum++

		log.Infof("存玩家帐目数据:%d, userId:%s 成功！！！", ind, userData.Username)
	}

	plen := len(userPlayers)

	log.Infof("玩家帐目数据入库完成 记录条数:%d / %d", saveNum, plen)

	if saveNum > 0 {

		//4 结算玩商的上级商户的投注统计
		self.AccountMerchantToDB(totayDay, userPlayers)
	}

	//标记重置
	self.bAccounted = false
	self.IsAccountWorking = false

	log.Infof("工作状态重置。。。。")

	//重新启动计时
	self.SafeNewTicker(&self.AccountTickerId, time.Second, nil, self.AccountToTick)

}

// 结算入库是否开始检测
func (self *CollectService) AccountWorkStart() bool {

	curTime := time.Now()

	//检测小时
	if curTime.Hour() != self.AccountTimer.Hour {
		return false
	}

	//检测分
	if curTime.Minute() != self.AccountTimer.Minute {
		return false
	}

	//检测秒
	if curTime.Second() != self.AccountTimer.Second {
		return false
	}

	//检测是否在入库
	if self.IsAccountWorking == true {
		return false
	}

	//检测否有重复结算了
	if curTime.Hour() == self.AccountTimer.LastAccountTime.Hour() &&
		curTime.Minute() == self.AccountTimer.LastAccountTime.Minute() &&
		curTime.Second() == self.AccountTimer.LastAccountTime.Second() {

		////检测是同一天时间，若是，表示重复结算了
		//if curTime.Day() != self.AccountTimer.LastAccountTime.Day() {
		//	//重复结算了
		//	self.bAccounted = false
		//
		//	log.Errorf(" 重复入库存了。。。。")
		//
		//	return false
		//}
	}

	if self.bAccounted == false {
		self.bAccounted = true
	} else {
		return false
	}

	//相当于现在要结算的时间
	self.AccountTimer.NowAccountTime = curTime

	return true
}

// 结算玩家的钱包帐目到商户代理中
func (self *CollectService) AccountMerchantToDB(accountTm time.Time, betPlayers []*models.SUserWallet) {

	merchantMap := make(map[string]*models.SMerchantAccount)

	//
	nowTime := time.Now()
	day1 := accountTm
	day2 := nowTime

	log.Infof("商户帐目数据入库开始....")

	for _, betItem := range betPlayers {

		if betItem.MerchantId == "" {
			log.Errorf("商户帐目数据空")
			continue
		}
		//1.找到它的上级直属商户，累加额度
		pmerchant, ok := merchantMap[betItem.MerchantId]

		if ok == false {

			merchantNode, ok := player.GetMerchantNode(betItem.MerchantId)

			if ok == false {
				log.Errorf("商户帐目数据空2 ")
				continue //需要处理找不到的情况
			}
			//若没有找到，则需要新生成
			merchantMap[betItem.MerchantId] = &models.SMerchantAccount{}
			pmerchant = merchantMap[betItem.MerchantId]

			pmerchant.MerchantId = betItem.MerchantId  //当前商户的ID
			pmerchant.ParentId = merchantNode.ParentId //上级商户的ID
			pmerchant.RootKey = merchantNode.MerchantKey

			pmerchant.Currency = merchantNode.Currency
			pmerchant.Digits = merchantNode.Digits

			pmerchant.CommissionRate = merchantNode.Commission_rate
			pmerchant.RollingRate = merchantNode.Water_rate

			//	pmerchant
			pmerchant.AccountDate = &accountTm
			pmerchant.StarTime = &day1
			pmerchant.EndTime = &day2

		}

		pmerchant.Bet += betItem.LastBet
		pmerchant.BetEffective += betItem.LastBetEffective
		pmerchant.Water += betItem.LastWater
		pmerchant.WinLost += betItem.LastWinLost
		pmerchant.Tip += betItem.Tip

		//若商户的上级不是公司，则表示它不是最上级商户
		if pmerchant.ParentId == "root" {
			continue
		}

		//2.找到上级商户的顶层商户，累加它的额度
		topMerchant, ok := merchantMap[pmerchant.ParentId]

		if ok == false {

			topNode, ok := player.GetMerchantNode(pmerchant.ParentId)

			if ok == false {
				//需要去数据库查找此商户记录

				continue
			}

			merchantMap[pmerchant.ParentId] = &models.SMerchantAccount{}

			topMerchant = merchantMap[pmerchant.ParentId]

			topMerchant.MerchantId = topNode.MerchantId //当前商户的ID
			topMerchant.ParentId = topNode.ParentId     //上级商户的ID
			topMerchant.RootKey = topNode.MerchantKey

			topMerchant.Currency = topNode.Currency
			topMerchant.Digits = topNode.Digits

			topMerchant.CommissionRate = topNode.Commission_rate
			topMerchant.RollingRate = topNode.Water_rate

			topMerchant.AccountDate = &accountTm
			topMerchant.StarTime = &day1
			topMerchant.EndTime = &day2
		}

		topMerchant.Bet += betItem.LastBet
		topMerchant.BetEffective += betItem.LastBetEffective
		topMerchant.Water += betItem.LastWater
		topMerchant.WinLost += betItem.LastWinLost
		topMerchant.Tip += betItem.Tip

	}

	//统计完代理商的帐目了，需要存入库
	oknum := 0

	for _, merchatItem := range merchantMap {

		err := merchatItem.Create()

		if err != nil {

			log.Errorf("商户帐目数据存不成功！！%s  -->err :%s", merchatItem.MerchantId, err)

			//err = merchatItem.Delete()
			//
			//if err != nil {
			//	continue
			//}
			//err = merchatItem.Create()

			continue
		}

		oknum++

		log.Infof("商户帐目数据存入成功:%s", merchatItem.MerchantId)

	}

	mlen := len(merchantMap)

	log.Infof("商户帐目数据入库完成 条数:%d 总条数:%d ....", oknum, mlen)

}
