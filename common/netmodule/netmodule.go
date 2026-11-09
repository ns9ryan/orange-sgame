package netModule

import "github.com/duanhf2012/origin/v2/network/processor"

type INetModule interface {
	SendRawMsg(clientId string, data []byte) error
	SendMsg(clientId string, data interface{}) error
	Close(clientId string)
	GetClientIp(clientId string) string
	GetProcessor() processor.IRawProcessor
	Start() error
}
