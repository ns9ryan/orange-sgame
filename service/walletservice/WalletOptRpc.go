package walletService

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	sysrpc "github.com/duanhf2012/origin/v2/rpc"
	"runtime"
	"time"
	models "videogame/common/model"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

// 操作钱包,供外服调用
func (self *WalletService) RPCWalletOpt(responder sysrpc.Responder, req *rpc.WalletOptReq) {

	ukey := uint64(util.StringToUintKey(req.UserId))

	//ukey := uint64(0)

	//fmt.Printf("****WalletService--->RPCWalletOpt  userId:%s ,opt:%d  \n", req.UserId, req.Opt)

	log.Infof("*********钱包服WalletService--->RPCWalletOpt  userId:%s ,opt:%d  gameId:%s, point: %v  ", req.UserId, req.Opt, req.GameId, req.Point)

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

	//fmt.Printf("钱包请求操作:%v \n", req)

	return
}

// 操作更新玩家的投注帐目,供游戏服RPC调用
func (self *WalletService) RPCWalletAddAccount(responder sysrpc.Responder, req *rpc.WalletAccount) {

	//fmt.Printf("222222 RPCWalletAddAccount\n")

	userId := req.UserId

	//入队列，让排队存入数据库
	dbItem := WalletDBItem{
		UserId:       userId,
		DBDataType:   DBData_Account,
		Balance:      0,
		Bet:          req.Bet,
		BetEffective: req.BetEffective,
		Water:        req.Water,
		WinLost:      req.WinLost,
		Tip:          0}

	if self.dbRefreshQueue != nil {
		self.dbRefreshQueue.Add(&dbItem)
	}

	//fmt.Printf("3333 存玩家帐目到队列: userId: %s \n , bet:%d water:%d ,winLost:%d \n", userId, req.Bet, req.Water, req.WinLost)
	log.Infof("存玩家帐目到队列: userId: %s \n , bet:%d water:%d ,winLost:%d \n", userId, req.Bet, req.Water, req.WinLost)

	responder(nil, sysrpc.RpcError(""))

}

// 增加打赏的金额
func (self *WalletService) RPCWalletAddTipAmount(responder sysrpc.Responder, req *rpc.WalletTipAmout) {

	userId := req.UserId

	//入队列，让排队存入数据库
	dbItem := WalletDBItem{
		UserId:       userId,
		DBDataType:   DBData_Tip,
		Balance:      0,
		Bet:          0,
		BetEffective: 0,
		Water:        0,
		WinLost:      0,
		Tip:          req.TipAmout,
	}

	if self.dbRefreshQueue != nil {
		self.dbRefreshQueue.Add(&dbItem)
	}

	//fmt.Printf("存玩家打赏数据到队列: userId: %s \n , tip:%d ", userId, req.TipAmout)

	log.Infof("存玩家打赏数据到队列: userId: %s \n , tip:%d ", userId, req.TipAmout)

	responder(nil, sysrpc.RpcError(""))

}

func (self *WalletService) ExecuteWalletOpt(channelWalletOpt chan WalletRequest) {

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.SError("core dump info:", err.Error(), "\n")
			self.ExecuteWalletOpt(channelWalletOpt)
		}
	}()

	for {
		select {
		case optData := <-channelWalletOpt:

			switch OptType(optData.request.Opt) {

			case WOpt_Add:
				fmt.Printf("钱包操作加点:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)

				//这边要扣玩家的点数
				balance, err := self.BalanceOpt(optData.request.UserId, optData.request.Point, optData.request.Opt)

				if err != nil {
					optData.responder(nil, sysrpc.RpcError(err.Error()))

					fmt.Printf("钱包操加点死败:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)
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

				//这边要扣玩家的点数
				balance, err := self.BalanceOpt(optData.request.UserId, optData.request.Point, optData.request.Opt)

				if err != nil {
					optData.responder(nil, sysrpc.RpcError(err.Error()))

					fmt.Printf("钱包操作减点死败:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)

				} else {

					var walletret rpc.WalletOptRet

					walletret.Opt = optData.request.Opt
					walletret.UserId = optData.request.UserId
					walletret.Point = optData.request.Point
					walletret.Balance = balance
					walletret.DataType = optData.request.DataType

					walletret.Success = 1

					optData.responder(&walletret, "")

					fmt.Printf("钱包操作减点成功:userId: %s  point:%v \n", optData.request.UserId, optData.request.Point)

				}
			case WOpt_Get:

				balance, err := self.GetBalance(optData.request.UserId)

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

				//初始化玩家钱包
				self.InitPlayerWallet(optData.request.UserId)

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

// 存储钱包余额到数据表
func (self *WalletService) WalletSaveToDB(dbItem *WalletDBItem) {

	var dbWallet models.SUserWallet

	dataType := dbItem.DBDataType

	if dataType == int(DBData_Balance) { //若不是帐目数据

		err := dbWallet.UpdateBalance(dbItem.UserId, dbItem.Balance)

		if err != nil {
			log.SError("save walletDB UpdateBalance failed!! err:%s \n", err.Error())
			return
		}

		fmt.Printf("数据库存余额成功:userId:%s,Balance: %d", dbItem.UserId, dbItem.Balance)

	} else if dataType == int(DBData_Account) { //处理帐目数据

		err := dbWallet.UpdateAccount(dbItem.UserId, uint(dbItem.Bet), uint(dbItem.BetEffective), uint(dbItem.Water), int(dbItem.WinLost))

		if err != nil {
			log.SError("save walletDB UpdateAccount BetWater failed!! err:%s \n", err.Error())
			return
		}

		fmt.Printf("数据库存帐号成功:userId:%s,Bet: %d WinLost:%d", dbItem.UserId, dbItem.Bet, dbItem.WinLost)

	} else if dataType == int(DBData_Tip) { //处理打赏的数据

		err := dbWallet.UpdateTip(dbItem.UserId, uint(dbItem.Tip))

		if err != nil {
			log.SError("save walletDB UpdateAccount Tip  failed!! err:%s \n", err.Error())
		}

	}

}

// 存储到数据库的协程 ，定时检测队列是否有数据需要存储
func (self *WalletService) ExecuteWalletSaveToDB() {

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
