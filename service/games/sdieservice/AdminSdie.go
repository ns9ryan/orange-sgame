package sdieservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/util"
	"videogame/service/games/adminmodule"
)

/*************************************************
********************************************** */

type AdminSdie struct {
	adminmodule.AdminControl
}

// 注册自定议的消息
func (mr *AdminSdie) RegisterCtrlMsg() {

	//游戏开答

	mr.RegMsgHandler(msg_struct.GameSdieOpenAnswerProtocol, &msg_struct.SdieOpenAnswer{}, mr.OnGameOpenAnswer)
}

// 游戏开奖消息
func (mr *AdminSdie) OnGameOpenAnswer(cliId string, msg interface{}) {

	answer := msg.(*msg_struct.SdieOpenAnswer)

	//fmt.Printf("ctrl--->server  Sdie answer:%v  \n", answer)

	var retMsg msg_struct.SdieOpenAnswer

	rpcFunc := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlToGameOpenAnswer")

	//本地服务调用
	err := mr.GetGs().Call(rpcFunc, answer, &retMsg)

	if err != nil {

		//发送给控端通知不成功
		var failedMsg msg_struct.GameChangeFailed

		failedMsg.Msgid = msg_struct.GameSdieOpenAnswerProtocol

		mr.SendMsg(cliId, msg_struct.GameMsgFailedProtocol, &failedMsg)

		log.StackErrorf("Ctrl OnGameOpenAnswer  To Service Error!!!", rpcFunc)

		return

	}

	mr.BroadCastMsg(msg_struct.GameSdieOpenAnswerProtocol, &retMsg)

	//读取历史开答路子
	//通知广播开答路子
	var gamePerio msg_struct.SdieAnswersPeriodic
	var input int

	rpcFun := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlGetAnswerPeriodic")

	//读取当前游戏名开答的表
	err = mr.GetGs().Call(rpcFun, &input, &gamePerio)

	if err == nil {

		//通知各个控端开答消息
		mr.BroadCastMsg(msg_struct.GameSdieAnswerPeriodicProtocol, &gamePerio)

		//fmt.Printf("历史开答列表，现在发给各个控端 \n")
	}

}
func (mr *AdminSdie) BroadStatusMsg(statusMsg *msg_struct.GameStatusMsg) {

	mr.BroadCastMsg(msg_struct.GameSdieStatusProtocol, statusMsg)

}
