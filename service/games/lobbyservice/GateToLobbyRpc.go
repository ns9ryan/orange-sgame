package lobbyservice

import (
	"errors"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/rpc"
)

// 玩家登录大厅，gate通过这个RPC调用通知大厅
func (self *LobbyService) RPCLogin(req *rpc.LoginToLobbyReq, ret *rpc.LoginToLobbyRet) error {

	//	fmt.Printf("Gate--->LobbyService RPCLogin user:%v \n", req)
	log.Infof("Gate--->LobbyService RPCLogin user:%v \n", req)

	//从redis读取帐号信息，若读不到了，则到数据库读取
	p := self.PlayersImp.GetPlayer(req.UserId)

	if p == nil { //创建用户

		p = self.PlayersImp.NewPlayer(req.UserId, &self.msgSender, self, &self.msgReceiver)

		if p == nil {
			ret.Ret = 1
			return errors.New("player not found!")
		}
	} else {
		//若原来的玩家已存在了,需要增加处理
	}

	//重新关联新连接
	self.attachConn(req, p)

	//读取玩家redis中的数据
	p.ReadUserInfo()

	//检测帐号是否暂停
	if p.IsBlocked() {
		ret.Ret = 2
		return errors.New("player is blocked")
	}

	//钱包初始化，检测是否单钱包模式
	if p.IsWalletAPIMode() == false {

		//初始化多钱包服
		self.InitPlayerWalletService(req)

	} else {
		//初始化单钱包服

		//多钱包服也要初始化，用于账目信息更新的RPC调用
		self.InitPlayerWalletService(req)

		//初始化单钱包服务
		self.InitPlayerSingleWalletService(req)

	}

	//加入投注服
	self.JoinPlayerBetService(req)

	//初始化玩家设置信息
	p.InitPlayerSetting()

	queueId := int64(1)

	//采用异步协程写入
	self.AsyncDoByQueue(queueId, func() bool {

		//检测是否首次登录,并写入数据表IP,时间
		if p.FirstLoginWite(req.Ip) == true {

			//需要更新到redis
			p.UpdateUserInfoToRedis()

		} else {
			//登陆IP,登陆时间写入表,玩家在线标志写入数据库
			p.WriteLoginIp(req.Ip)
		}

		p.WriteLoginLog(req.Ip, 1, "success")

		return false //这个不是表示不成功

	}, nil)

	//在线标志，登陆时间写入内存
	p.SetOnline(true)

	//更新上线时间写入内存
	p.UpdateLoginTime()

	ret.Ret = 0

	return nil
}

// 通知客户端成功送发后的回调消息，用于发送更多的大厅相关信息
func (self *LobbyService) RPCLoginSuccess(req *rpc.LoginToLobbyReq, ret *rpc.Int) error {

	//发送帐号的相关信息
	self.SendToPlayerUserInfo(req.UserId)

	//广播在线人数信息
	self.CastToPlayersPlayerNum()

	//发送游戏大厅相关的游戏配置
	self.SendToPlayerGameGroups(req.UserId)

	//发送游戏大厅的路子信息
	self.SendToPlayerAnswers(req.UserId)

	//发送每个游戏的在线人数情况
	self.SendToPlayerGamePlayerNum(req.UserId)

	//发送系统通知消息
	self.SendToPlayerNoticMessage(req.UserId)

	//fmt.Println("LobbyService   RPCSuccessLogin ok****************************")

	log.Infof("LobbyService   RPCSuccessLogin  %v ok****************************", req)

	ret.Value = 1

	return nil

}
