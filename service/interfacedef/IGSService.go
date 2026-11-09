package interfacedef

import (
	"github.com/duanhf2012/origin/v2/rpc"
)

type IGSService interface {
	rpc.IRpcHandler
	GetGateNodeIdByClientId(clientId string) string
	GetClientPlayer(clientID string) IPlayer
	GetClientIdByPlayerId(playerId string) string
	CloseClient(clientId string)
	DestroyPlayer(playerId string) bool

	GetAllClientId() []string
	GetMsgReceiver() IMsgReceiver
	GetMsgSender() IMsgSender
	GetPlayerTimer() IPlayerTimer
}

type IGSServiceJson interface {
	rpc.IRpcHandler
	GetGateNodeIdByClientId(clientId string) string
	GetClientPlayer(clientID string) IPlayer
	GetClientIdByPlayerId(playerId string) string
	CloseClient(clientId string)
	DestroyPlayer(playerId string) bool

	GetMsgReceiver() IMsgReceiverJson
	GetMsgSender() IMsgSenderJson
	GetPlayerTimer() IPlayerTimer
}
