package sdieservice

import (
	"videogame/common/msg_struct"
)

/*
  本文件主要是供本地服务模块调用本地的RPC服务
*/

// 当服务器触发停止下注时，需要本地RPC通知控端服务停止下注
func (self *SdieService) RPCStopBetToCtrlService(msg *msg_struct.GameStatusMsg, ret *int) error {

	//fmt.Println("RPCStopBetToCtrlService Call ok ")

	self.adminCtrl.BroadCastMsg(msg_struct.GameSdieStatusProtocol, msg)

	return nil
}
