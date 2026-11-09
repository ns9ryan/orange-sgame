package bacxservice

/*
  本文件主要接收控端服务传送过来的RPC调用
*/

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"time"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/msg_struct/msg_ctrl"
)

// 荷官进入控端
func (self *BacxService) RPCAdminLogin(req *msg_ctrl.LoginToCtrlService, dealer *msg_struct.DealerInfo) error {

	//fmt.Printf(" %s ... RPCLogin  %v  \n", self.GetName(), req)

	//荷管加入
	fmt.Println("BacxService  RPCAdminLogin ok ")

	//去数据库或redis读取此荷官的信息
	dealer.DealerId = req.UserId
	dealer.Name = "system"
	dealer.Avator = ""
	dealer.Level = 99

	self.GameTable.UpdateDealer(dealer)

	//通知大厅此游戏是哪个荷官
	return nil
}

// 荷官成功进入游戏了，可以发送配置信息了
func (self *BacxService) RPCAdminLoginOk(loginOk *msg_ctrl.CtrlLoginSuccess, ret *int) error {

	//.Printf(" %s --admin:%s ... RPCAdminLoginOk ok ok ok \r\n ", self.GetName(), loginOk.UserId)

	log.Infof(" %s --admin :%s ... RPCAdminLoginOk ok o \r\n ", self.GetName(), loginOk.UserId)

	*ret = self.GameTable.GetDeskNo() //桌号

	//发送服务器时间， 发送游戏配置 ， 当前游戏状态， 当前游戏的开答路子
	var gameTime msg_struct.GameServerTime

	nowTime := time.Now()
	gameTime.ServerTime = nowTime.Format("2006-01-02 15:04:05")
	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameServerTimeProtocol, gameTime)

	//发送游戏服务器配置
	var gameConfig msg_struct.GameToCtrlGameConfig

	self.GameTable.GetTableConfig(&gameConfig)
	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameBacxConfigProtocol, gameConfig)

	//发送当前游戏状态
	var gmStatus msg_struct.GameStatusMsg

	self.GameTable.GetGameStatus(&gmStatus)
	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameBacxStatusProtocol, gmStatus)

	fmt.Printf("send to admin  status:%v  ", gmStatus.GameStatus)

	//发送当前游戏开答的路子消息
	var gamePerio msg_struct.GameBacXAnswersPeriodic

	self.GameTable.GetAnswerPeriodic(&gamePerio)

	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameBacxAnswerPeriodicProtocol, gamePerio)

	fmt.Printf("发送历史开答给控端 to admin  status:%v  \n ", gamePerio.Winners)

	return nil
}

// 荷官进入房间
func (self *BacxService) RPCCtrlDealerJoinRoom(dealerId *string, ret *msg_ctrl.CtrlDealerInfo) error {

	//fmt.Println("BacxService  RPCCtrlDealerJoinRoom ok ")

	var dealer msg_struct.DealerInfo

	dealerInfo, err := models.GetDealerByUid(*dealerId)

	if err != nil {

		log.Errorf("GetDealerByUid :%s, error:%v", *dealerId, err)

		ret.Success = false

		return err
	}

	ret.Success = true

	//去数据库或redis读取此荷官的信息
	dealer.Name = dealerInfo.Nickname
	dealer.Avator = dealerInfo.Avatar
	dealer.DealerId = *dealerId
	dealer.Level = dealerInfo.Level

	self.GameTable.UpdateDealer(&dealer)

	//通知大厅此游戏是哪个荷官
	ret.GameId = self.GameTable.GameInfo.GameId
	ret.Level = dealer.Level
	ret.Name = dealerInfo.Nickname
	ret.DearlerId = *dealerId

	log.Infof("Dealer Work Ok :%s ", *dealerId)

	return nil
}

// RPC控端更新游戏状态的通知
func (self *BacxService) RPCCtrlChangeGameStatus(msg *msg_struct.GameStatusMsg, retmsg *msg_struct.GameStatusMsg) error {

	//改变游戏状态
	update := self.GameTable.SetGameStatus(msg.Set, msg.Round, msg.GameStatus)

	if update == true {

		self.GameTable.GetGameStatus(retmsg)

		fmt.Println("*****BacxService .... RPCCtrlToGameChangeStatus ok      ")

	} else {

		*retmsg = *msg
		retmsg.GameStatus = msg_struct.GameStatusError

		return errors.New("update game status failed!!!")
	}

	return nil
}

// RPC控端上传游戏服的开奖消息
func (self *BacxService) RPCCtrlToGameOpenAnswer(answer *msg_struct.GameOpenAnswerBacx, retmsg *msg_struct.GameOpenAnswerBacx) error {

	log.Infof("BacxService   RPCCtrlToGameOpenAnswer  ok   , set:%d round:%d , answer.Winner:%d \n", answer.Set, answer.Round, answer.Winner)

	//先检测是否已经开奖过
	if self.GameTable.IsOpenAnswer() == true {

		return errors.New("repeat error")
	}

	//游戏开奖
	ret := self.GameTable.OpenAnswer(answer)

	if ret == false {

		retmsg.Set = 0
		retmsg.Round = 0

		return errors.New("open answer failed")
	}

	*retmsg = *answer

	return nil
}

// 通过RPC读取游戏当前的开答记录表
func (self *BacxService) RPCCtrlGetAnswerPeriodic(input *int, answer *msg_struct.GameBacXAnswersPeriodic) error {

	//fmt.Println("BacxService  RPCCtrlGetAnswerPeriodic  ok...")

	self.GameTable.GetAnswerPeriodic(answer)

	return nil
}
