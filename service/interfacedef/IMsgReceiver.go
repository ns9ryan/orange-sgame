package interfacedef

import (
	"videogame/common/proto/msg"
)

type IMsgHandler interface {
	MsgCb(p IPlayer, msg []byte)
	GmCb(p IPlayer, msgBody []byte)
	GetMsgType() msg.MsgType
}

type IMsgReceiver interface {
	GmReceiver(p IPlayer, msgType msg.MsgType, msgBody []byte) bool
}

type IMsgHandlerJson interface {
	MsgCb(p IPlayer, msg []byte)
	GmCb(p IPlayer, msgBody []byte)
	GetMsgType() int
}

type IMsgReceiverJson interface {
	GmReceiver(p IPlayer, msgType int, msgBody []byte) bool
}
