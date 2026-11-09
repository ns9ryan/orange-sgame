package gateservice

import "videogame/common/proto/rpc"

// 踢掉玩家
func (gate *GateService) RPCKickOutPlayer(clientId *rpc.String, ret *rpc.Int) error {

	uid := clientId.Value

	if gate.msgRouter.KickOutPlayer(uid) == true {

		ret.Value = 1
	} else {
		ret.Value = 0
	}

	return nil
}
