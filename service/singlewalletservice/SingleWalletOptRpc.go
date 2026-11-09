package singleWalletService

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	sysrpc "github.com/duanhf2012/origin/v2/rpc"
	models "videogame/common/model"
	"runtime"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"errors"
	"time"
)

type OptType int32

const (
	WOpt_Add  OptType = 0 //加点操作
	WOpt_Dec  OptType = 1 //减点操作
	WOpt_Get  OptType = 2 //读取操作
	WOpt_Init OptType = 3 //初始化钱包
	WOpt_Answer OptType = 4 //通知商户单钱包服务，游戏开答了
)

//根据dataType得到tranId的订单类型备注
type TransType int32

const (
	TransType_Bet 		TransType = 0		//下注订单
	TransType_CancelBet TransType = 1 		//取消下注订单
	TransType_Tip 		TransType = 2		//小费订单
	TransType_Win 		TransType = 3		//结算订单
	TransType_Other 	TransType = 4		//其他订单
)

var TranIdDataTypeMap = map[TransType]string{
    TransType_Bet: 		 "bet",        //发给商户的下注订单备注
    TransType_CancelBet: "cancel_bet", //发给商户的取消下注订单备注
    TransType_Tip: 		 "tip",        //发给商户的小费订单备注
    TransType_Win: 		 "win",        //发给商户的结算订单备注
    TransType_Other: 	 "other",      //发给商户的其他订单备注
}

// 操作钱包
func (self *SingleWalletService) RPCWalletOpt(responder sysrpc.Responder, req *rpc.WalletOptReq) {

	ukey := uint64(util.StringToUintKey(req.UserId))

	//ukey := uint64(0)

	fmt.Printf("****SingleWalletService--->RPCWalletOpt  userId:%s ,opt:%d  \n", req.UserId, req.Opt)

	//检测是否小于2，防止除数为0
	if self.goroutineNum < 2 {
		self.goroutineNum = 2
	}

	index := ukey % uint64(self.goroutineNum-1)

	if len(self.channelWalletOpts[index]) == cap(self.channelWalletOpts[index]) {

		log.SError("channel is full ", index)

		responder(nil, sysrpc.RpcError("channel is full"))
		return
	}

	var walRequest WalletRequest

	walRequest.request = req
	walRequest.responder = responder

	//存入操作队列
	self.channelWalletOpts[index] <- walRequest

	fmt.Printf("钱包请求操作:%v \n", req)

	return
}

func (self *SingleWalletService) ExecuteSingleWalletOpt(index uint32, channelWalletOpt chan WalletRequest) {

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.SError("core dump info:", err.Error(), "\n")
			self.ExecuteSingleWalletOpt(index, channelWalletOpt)
		}
	}()

	for {
		select {
		case optData := <-channelWalletOpt:

			switch OptType(optData.request.Opt) {

			case WOpt_Add:
				fmt.Printf("钱包操作加点，可能是结算订单或取消下注订单:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)

				//这边要扣玩家的点数
				balance, err := self.BalanceOptSingle(index, optData.request)

				if err != nil {
					optData.responder(nil, sysrpc.RpcError(err.Error()))

					fmt.Printf("钱包操加点失败:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)
				} else {

					var walletret rpc.WalletOptRet

					walletret.Opt = optData.request.Opt
					walletret.UserId = optData.request.UserId
					walletret.Point = optData.request.Point
					walletret.Balance = balance
					walletret.Success = 1

					optData.responder(&walletret, "")

					fmt.Printf("钱包操作加点成功:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)
				}

			case WOpt_Dec:

				fmt.Printf("钱包操作减点，可能是下注订单或小费订单:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)

				//这边要扣玩家的点数
				balance, err := self.BalanceOptSingle(index, optData.request)

				if err != nil {
					optData.responder(nil, sysrpc.RpcError(err.Error()))

					fmt.Printf("钱包操作减点失败:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)
				} else {

					var walletret rpc.WalletOptRet

					walletret.Opt = optData.request.Opt
					walletret.UserId = optData.request.UserId
					walletret.Point = optData.request.Point
					walletret.Balance = balance
					walletret.Success = 1

					optData.responder(&walletret, "")

					fmt.Printf("钱包操作减点成功:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)
				}
			case WOpt_Get:

				balance, err := self.GetBalanceSingle(index, optData.request.UserId)

				if err != nil {

					optData.responder(nil, sysrpc.RpcError(err.Error()))
					fmt.Printf("钱包读取余额不成功:userId: %s  \n", optData.request.UserId)

				} else {

					var walletret rpc.WalletOptRet

					walletret.Opt = optData.request.Opt
					walletret.UserId = optData.request.UserId
					walletret.Point = optData.request.Point
					walletret.Balance = balance
					walletret.Success = 1

					optData.responder(&walletret, "")
				}
			case WOpt_Init:

				fmt.Printf("=======收到WOpt_Init动作(注意本动作已被替换为RPCWalletInit调用)=========\n")

				//初始化玩家钱包配置
				self.InitPlayerSingleWallet(optData.request.UserId)

				optData.responder(nil, "")

			case WOpt_Answer:

				fmt.Printf("=======收到WOpt_Answer动作，需通知商户单钱包服务游戏开答了\n")
				self.RoundOpenAnswer(index, optData.request)

				optData.responder(nil, "")

			default:
				//log.SError("optype ", optData.request.GetType(), " is error.")
			}
		}
		if self.bCloseStatus == true {
			break
		}

	}
}

// 初始化玩家的钱包配置，包括币种/商户码等
func (self *SingleWalletService) InitPlayerSingleWallet(userId string) {

	pplayer := self.playerFactory.GetPlayer(userId)

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {
		return
	}

	pplayer.InitBalance()
}

func (self *SingleWalletService) BalanceOptSingle(index uint32, request *rpc.WalletOptReq) (float64, error) {

	fmt.Println("SingleWalletService BalanceOptSingle---------------------")

	pplayer := self.playerFactory.GetPlayer(request.UserId)

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {
		return 0, errors.New("not find player !!!")
	}

	balance := float64(0)
	transState := "unset"
	reason := ""
	var err error
	var finalErr error

	//加点操作
	queueId := int64(1000 + index) //协程池，单钱包队列id从1000-1004
	if request.Opt == int32(WOpt_Add) {

		if request.DataType == int32(TransType_Win) {

			err, balance = self.merchantApiCall.BalanceAdd(queueId, request, pplayer)
			if err != nil {
				fmt.Printf("请求商户BalanceAdd failed, err=%v\n", err)

				//存入挂起订单数据库，启动后台周期携程，尝试再加点，成功进历史库，失败一致试
				transState = "error"
				reason = err.Error()
			}

			transState = "set"

		}else if request.DataType == int32(TransType_CancelBet) {

			err, balance = self.merchantApiCall.BalanceCancelBet(queueId, request, pplayer)
			if err != nil {
				fmt.Printf("请求商户BalanceAdd failed, err=%v\n", err)
			}

			//取消下注不管商户那边返回与否，我们这边算处理完成
			transState = "cancel"
		}

	} else if request.Opt == int32(WOpt_Dec) { //减点操作

		err, balance = self.merchantApiCall.BalanceDec(queueId, request, pplayer)
		if err != nil {

			fmt.Printf("请求商户BalanceDec failed, err=%v\n", err)

			//存入挂起订单数据库，启动后台周期携程，不再尝试减点，转入历史订单库
			transState = "error"
			reason = err.Error()

			finalErr = err
		}

	} else {

		finalErr = errors.New("wallet opt no supper!!!")

	}

	//只有正常加扣款，才更新or_player_wallet
	if finalErr == nil {
		if pplayer.Digits < 1 {
			pplayer.Digits = 1
		}

		credit := uint64(balance) * uint64(pplayer.Digits)

		//更新or_player_wallet数据库及redis,缓存先更新，数据延后1秒左右更新，单钱包只是用于记录
		//在高并发的情况下，这一步不一定要做
		pplayer.Credit = uint(credit)
		pplayer.Currency = pplayer.Code
		pplayer.SaveWalletToRedis()

		//入队列，让排队存入数据库
		dbItem := WalletDBItem{
			UserId:       request.UserId,
			DBDataType:   DBData_Balance,
			Balance:      credit,
			Bet:          0,
			BetEffective: 0,
			Water:        0,
			WinLost:      0,
			Tip:          0,
		}

		if self.dbRefreshQueue != nil {
			self.dbRefreshQueue.Add(&dbItem)
		}
	}

	//不管订单成功或失败，都会排队存入临时订单库
	now := time.Now()
	if request.DataType != int32(TransType_CancelBet) {

		dbTransItem := models.SWalletTransPending{
			UserId:     	request.UserId,  		//玩家ID
			MerchantId:     pplayer.MerchantKey,  	//玩家ID
			WalletType:     1,  					//多钱包0单钱包1
			GameId:   		request.GameId,  		//游戏子类ID ROT
			DeskNo:     	request.DeskNo,  		//桌号 5
			RoundId:		request.RoundId, 		//对局ID OR25042809050122
			TransId:		request.TranId, 		//订单号 174580467174892375782966
			Opt:			request.Opt,	 		//加点还是减点
			DataType:		request.DataType,		//订单详细操作类型,0下注/1取消下注/2小费/3结算/4其他订单
			Point: 			request.Point, 	 		//0.5
			AfterBalance: 	balance, 		 		//操作后最新总金额
			State:			transState,
			Reason:			reason,		 		 	//取消下注订单会写玩家撤销/网络异常
			CreateTime: 	&now,
			UpdateTime: 	&now,
		}

		if self.dbTransQueue != nil {
			self.dbTransQueue.Add(&dbTransItem)
		}

	}else {
		var dbTransItem models.SWalletTransPending

		reason := "Player canceled the bet."
		err := dbTransItem.CancelBetById(request.TranId, reason, balance, now)
		if err != nil {
			fmt.Printf("Update wallet trans failed, err=%v\n", err)
		}

	}

	return balance, finalErr

}

// 读取玩家余额
func (self *SingleWalletService)GetBalanceSingle(index uint32, userId string) (float64, error) {

	fmt.Println("WalletService   GetBalanceSingle---------------------")

	pplayer := self.playerFactory.GetPlayer(userId)

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {
		return 0, errors.New("find player !!!")
	}

	queueId := int64(1000 + index) //协程池，单钱包队列id从1000-1004

	err, balance := self.merchantApiCall.GetBalance(queueId, userId, pplayer.Code, pplayer.MerchantKey)
	if err != nil {
		fmt.Printf("get balance failed, err=%v\n", err)
	}

	return balance, nil

}

// 通知对局结算
func (self *SingleWalletService) RoundOpenAnswer(index uint32, request *rpc.WalletOptReq) (error) {

	fmt.Println("SingleWalletService RoundOpenAnswer---------------------")

	pplayer := self.playerFactory.GetPlayer(request.UserId)

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {
		return errors.New("not find player !!!")
	}

	var err error

	//加点操作
	queueId := int64(1000 + index) //协程池，单钱包队列id从1000-1004

	err = self.merchantApiCall.OpenAnswer(queueId, request, pplayer)
	if err != nil {
		fmt.Printf("OpenAnswer failed, err=%v\n", err)
	}			


	var dbTransItem models.SWalletTransPending
	err = dbTransItem.UpdateTrans(request.RoundId)
	if err != nil {
		fmt.Printf("Update wallet trans failed, err=%v\n", err)
	}


	return nil
}

// 存储到数据库的协程 ，定时检测队列是否有数据需要存储
func (self *SingleWalletService) ExecuteWalletSaveToDB() {

	for self.bCloseStatus == false {

		if self.dbRefreshQueue == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		//若队列中数据没有要更新的
		if self.dbRefreshQueue.Len() == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		//检测是否正在结算入库操作
		if self.ReadingTableStatus == true {

			time.Sleep(100 * time.Second)

			log.Errorf("结算系统正在结算。。钱包服暂时不存数据表")

			continue
		}

		//可以读取队列数据了
		item := self.dbRefreshQueue.Pop()

		if item == nil {
			continue
		}

		dbItem := item.(*WalletDBItem)

		if dbItem != nil {
			self.WalletSaveToDB(dbItem)
		}

	}

}

// 存储钱包余额到数据表
func (self *SingleWalletService) WalletSaveToDB(dbItem *WalletDBItem) {

	var dbWallet models.SUserWallet

	dataType := dbItem.DBDataType

	//只处理or_player_wallet的更新，不处理账目数据和打赏数据，那都由多钱包服处理了
	if dataType == int(DBData_Balance) { //若不是帐目数据

		err := dbWallet.UpdateBalance(dbItem.UserId, dbItem.Balance)

		if err != nil {
			log.SError("save walletDB UpdateBalance failed!! err:%s \n", err.Error())
			return
		}

		fmt.Printf("数据库存余额成功:userId:%s,Balance: %d", dbItem.UserId, dbItem.Balance)

	}

}

// 存储订单到数据库的协程 ，定时检测队列是否有订单需要存储
func (self *SingleWalletService) ExecuteTransSaveToDB() {

	for self.bCloseStatus == false {

		if self.dbTransQueue == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		//若队列中数据没有要更新的订单数据
		if self.dbTransQueue.Len() == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		//可以读取队列数据了
		item := self.dbTransQueue.Pop()

		if item == nil {
			continue
		}

		dbTransItem := item.(*models.SWalletTransPending)

		if dbTransItem != nil {
			self.TransSaveToDB(dbTransItem)
		}

	}

}

// 存储订单历史到Trans订单表
func (self *SingleWalletService) TransSaveToDB(dbTransItem *models.SWalletTransPending) {

	fmt.Printf("===数据库存订单:dbTransItem:%v\n", dbTransItem)
	err := dbTransItem.Save()
	if err != nil {
		fmt.Printf("===数据库存订单失败err:%v\n", err)
	}

}
