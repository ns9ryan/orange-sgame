package rotservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/util"
	"videogame/service/games/adminmodule"
)

/*************************************************
********************************************** */

type AdminRot struct {
	adminmodule.AdminControl
}

// 注册自定议的消息
func (mr *AdminRot) RegisterCtrlMsg() {

	//游戏开答
	mr.RegMsgHandler(msg_struct.GameRotOpenAnswerProtocol, &msg_struct.GameOpenAnswerRot{}, mr.OnGameOpenAnswer)
}

// 游戏开奖
func (mr *AdminRot) OnGameOpenAnswer(cliId string, msg interface{}) {

	answer := msg.(*msg_struct.GameOpenAnswerRot)

	fmt.Println("ctrl--->server   rot answer  : ", answer)

	var retMsg msg_struct.GameOpenAnswerRot

	rpcFunc := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlToGameOpenAnswer")

	//本地服务调用
	err := mr.GetGs().Call(rpcFunc, answer, &retMsg)

	if err != nil {
		//发送给控端通知不成功
		var failedMsg msg_struct.GameChangeFailed

		failedMsg.Msgid = msg_struct.GameRotOpenAnswerProtocol

		mr.SendMsg(cliId, msg_struct.GameMsgFailedProtocol, &failedMsg)

		log.Errorf("Ctrl OnGameOpenAnswer  To Service Error!!! rpcFunc=%s", rpcFunc)

		return

	}

	mr.BroadCastMsg(msg_struct.GameRotOpenAnswerProtocol, &retMsg)

	//读取历史开答路子
	//通知广播开答路子
	var gamePerio msg_struct.GameRotAnswersPeriodic
	var input int

	rpcFun := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlGetAnswerPeriodic")

	//读取当前游戏名开答的表
	err = mr.GetGs().Call(rpcFun, &input, &gamePerio)

	if err == nil {

		//通知各个控端开答消息
		mr.BroadCastMsg(msg_struct.GameRotAnswerPeriodicProtocol, &gamePerio)

		fmt.Printf("历史开答列表，现在发给各个控端 \n")
	}

}

func (mr *AdminRot) BroadStatusMsg(statusMsg *msg_struct.GameStatusMsg) {

	mr.BroadCastMsg(msg_struct.GameRotStatusProtocol, statusMsg)

}
