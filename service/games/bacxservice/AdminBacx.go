package bacxservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/util"
	"videogame/service/games/adminmodule"
)

/*************************************************

百家乐控端接口

********************************************** */

type AdminBacx struct {
	adminmodule.AdminControl
}

// 游戏开奖消息
func (mr *AdminBacx) OnGameOpenAnswer(cliId string, msg interface{}) {

	answer := msg.(*msg_struct.GameOpenAnswerBacx)

	fmt.Printf("ctrl--->server   bacx answer:%v   Winner: %d ppair:%d bpair:%d \n", answer, answer.Winner.Winner, answer.Winner.PPair, answer.Winner.BPair)

	var retMsg msg_struct.GameOpenAnswerBacx

	rpcFunc := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlToGameOpenAnswer")

	//本地服务调用
	err := mr.GetGs().Call(rpcFunc, answer, &retMsg)

	if err != nil {
		//发送给控端通知不成功
		var failedMsg msg_struct.GameChangeFailed

		failedMsg.Msgid = msg_struct.GameBacxOpenAnswerProtocol

		mr.SendMsg(cliId, msg_struct.GameMsgFailedProtocol, &failedMsg)

		log.Errorf("Ctrl OnGameOpenAnswer  To Service Error!!! rpcFunc=%s", rpcFunc)

		return

	}

	mr.BroadCastMsg(msg_struct.GameBacxOpenAnswerProtocol, &retMsg)

	//读取历史开答路子
	//通知广播开答路子
	var gamePerio msg_struct.GameBacXAnswersPeriodic
	var input int

	rpcFun := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlGetAnswerPeriodic")

	//读取当前游戏名开答的表
	err = mr.GetGs().Call(rpcFun, &input, &gamePerio)

	if err == nil {

		//通知各个控端开答消息
		mr.BroadCastMsg(msg_struct.GameBacxAnswerPeriodicProtocol, &gamePerio)

		fmt.Printf("历史开答列表，现在发给各个控端 \n")
	}

}

// 注册自定议的消息
func (mr *AdminBacx) RegisterCtrlMsg() {

	//游戏开答
	mr.RegMsgHandler(msg_struct.GameBacxOpenAnswerProtocol, &msg_struct.GameOpenAnswerBacx{}, mr.OnGameOpenAnswer)

}
func (mr *AdminBacx) BroadStatusMsg(statusMsg *msg_struct.GameStatusMsg) {

	mr.BroadCastMsg(msg_struct.GameBacxStatusProtocol, statusMsg)

}
