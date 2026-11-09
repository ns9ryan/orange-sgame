package interfacedef

import (
	"google.golang.org/protobuf/proto"
	"videogame/common/proto/msg"
)

type IMsgSender interface {
	SendToClient(clientId string, msgType msg.MsgType, message proto.Message) int
	CastToClient(clientIdList []string, msgType msg.MsgType, message proto.Message) int
	SendToPlayer(playerUserId string, msgType msg.MsgType, message proto.Message) int
	CastToPlayer(playerUserId []string, msgType msg.MsgType, message proto.Message) int
}

type IMsgSenderJson interface {
	SendToClient(clientId string, msgType int, message interface{}) int
	CastToClient(clientIdList []string, msgType int, message interface{}) int

	SendToPlayer(playerUserId string, msgType int, message interface{}) int
	CastToPlayer(playerUserId []string, msgType int, message interface{}) int
}
