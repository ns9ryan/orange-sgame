package ftanservice

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
func (self *FtanService) RPCAdminLogin(req *msg_ctrl.LoginToCtrlService, dealer *msg_struct.DealerInfo) error {

	fmt.Printf(" FtanService ... RPCLogin  %v  \n", req)

	//荷管加入

	fmt.Println("FtanService  RPCAdminLogin ok ")
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
func (self *FtanService) RPCAdminLoginOk(loginOk *msg_ctrl.CtrlLoginSuccess, ret *int) error {

	fmt.Println(" FtanService ... RPCAdminLoginOk ok ok ok \r\n ", loginOk.UserId)

	*ret = self.GameTable.GetDeskNo() //桌号

	//发送服务器时间， 发送游戏配置 ， 当前游戏状态， 当前游戏的开答路子
	var gameTime msg_struct.GameServerTime
	nowTime := time.Now()
	gameTime.ServerTime = nowTime.Format("2006-01-02 15:04:05")
	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameServerTimeProtocol, gameTime)

	//发送游戏服务器配置
	var gameConfig msg_struct.GameToCtrlGameConfig

	self.GameTable.GetTableConfig(&gameConfig)
	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameFtanConfigProtocol, gameConfig)

	//发送当前游戏状态
	var gmStatus msg_struct.GameStatusMsg

	self.GameTable.GetGameStatus(&gmStatus)
	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameFtanStatusProtocol, gmStatus)

	fmt.Printf("send to admin  status:%v  ", gmStatus.GameStatus)

	//发送当前游戏开答的路子消息
	var gamePerio msg_struct.FtanAnswersPeriodic

	self.GameTable.GetAnswerPeriodic(&gamePerio)

	self.adminCtrl.SendMsg(loginOk.ClientId, msg_struct.GameFtanAnswerPeriodicProtocol, gamePerio)

	fmt.Printf("发送历史开答给控端 to admin  status:%v  \n ", gamePerio.Winners)

	return nil
}

// 荷官进入房间
func (self *FtanService) RPCCtrlDealerJoinRoom(dealerId *string, ret *msg_ctrl.CtrlDealerInfo) error {

	fmt.Println("FtanService  RPCCtrlDealerJoinRoom ok ")
	//去数据库或redis读取此荷官的信息

	var dealer msg_struct.DealerInfo

	//从数据库读取荷管信息
	dealerInfo, err := models.GetDealerByUid(*dealerId)

	if err != nil {
		ret.Success = false
		return err
	}

	dealer.Name = dealerInfo.Nickname
	dealer.Avator = dealerInfo.Avatar
	dealer.DealerId = *dealerId
	dealer.Level = dealerInfo.Level

	dealer.Name = dealerInfo.Nickname
	dealer.Avator = dealerInfo.Avatar
	dealer.DealerId = *dealerId

	self.GameTable.UpdateDealer(&dealer)

	ret.GameId = self.GameTable.GameInfo.GameId
	ret.Name = dealerInfo.Nickname
	ret.Avator = dealerInfo.Avatar
	ret.Level = dealerInfo.Level
	ret.DearlerId = *dealerId

	ret.Success = true

	return nil
}

// RPC控端更新游戏状态的通知
func (self *FtanService) RPCCtrlChangeGameStatus(msg *msg_struct.GameStatusMsg, retmsg *msg_struct.GameStatusMsg) error {

	//fmt.Println("*****FtanService .... RPCCtrlToGameChangeStatus ok      ")

	//改变游戏状态
	update := self.GameTable.SetGameStatus(msg.Set, msg.Round, msg.GameStatus)

	if update == true {

		self.GameTable.GetGameStatus(retmsg)

	} else {

		*retmsg = *msg
		retmsg.GameStatus = msg_struct.GameStatusError

		return errors.New("update game status failed!!!")
	}

	return nil
}

// RPC控端上传游戏服的开奖消息
func (self *FtanService) RPCCtrlToGameOpenAnswer(answer *msg_struct.FtanOpenAnswer, retmsg *msg_struct.FtanOpenAnswer) error {

	log.Infof("Ftan Service   RPCCtrlToGameOpenAnswer  ok   , set:%d round:%d , answer.Winner:%d", answer.Set, answer.Round, answer.Winner)

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
func (self *FtanService) RPCCtrlGetAnswerPeriodic(input *int, answer *msg_struct.FtanAnswersPeriodic) error {

	log.Info("FtanService  RPCCtrlGetAnswerPeriodic  ok...")

	self.GameTable.GetAnswerPeriodic(answer)

	return nil
}
