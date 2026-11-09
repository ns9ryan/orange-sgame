package longhuservice

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

type AdminDtx struct {
	adminmodule.AdminControl
}

// 注册自定议的消息
func (mr *AdminDtx) RegisterCtrlMsg() {

	//游戏开答
	mr.RegMsgHandler(msg_struct.GameDtxOpenAnswerProtocol, &msg_struct.GameOpenAnswerDTX{}, mr.OnGameOpenAnswer)

}

// 游戏开奖
func (mr *AdminDtx) OnGameOpenAnswer(cliId string, msg interface{}) {

	dtxanswer, ok := msg.(*msg_struct.GameOpenAnswerDTX)

	if ok == false {
		return
	}

	fmt.Println("ctrl--->server   dtx answer  : ", *dtxanswer)

	var retMsg msg_struct.GameOpenAnswerDTX

	rpcFunc := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlToGameOpenAnswer")

	//本地服务调用
	err := mr.GetGs().Call(rpcFunc, dtxanswer, &retMsg)

	if err != nil {
		//发送给控端通知不成功
		var failedMsg msg_struct.GameChangeFailed

		failedMsg.Msgid = msg_struct.GameDtxOpenAnswerProtocol

		mr.SendMsg(cliId, msg_struct.GameMsgFailedProtocol, &failedMsg)

		log.Errorf("Ctrl OnGameOpenAnswer  To Service Error!!!", rpcFunc)

		return

	}

	//通知各个控端开答消息
	mr.BroadCastMsg(msg_struct.GameDtxOpenAnswerProtocol, &retMsg)

	//读取历史开答路子
	//通知广播开答路子
	var gamePerio msg_struct.GameDtxAnswersPeriodic
	var input int

	rpcFun := util.FormatRpcMethod(mr.GetGs().GetName(), "RPCCtrlGetAnswerPeriodic")

	//读取当前游戏名开答的表
	err = mr.GetGs().Call(rpcFun, &input, &gamePerio)

	if err == nil {

		mr.BroadCastMsg(msg_struct.GameDtxAnswerPeriodicProtocol, &gamePerio)

		fmt.Printf("历史开答列表，现在发给各个控端 \n")
	}
}

func (mr *AdminDtx) BroadStatusMsg(statusMsg *msg_struct.GameStatusMsg) {

	mr.BroadCastMsg(msg_struct.GameDtxStatusProtocol, statusMsg)

}
