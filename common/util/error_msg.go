package util

import "videogame/common/proto/msg"

const ()

const (
	SystemError         = "System Error"
	MsgForbidGuestLogin = "Forbid Guest Login"
	CloseServerError    = "Close Server Error"
	SessionInvalid      = "Session Invalid"
	ParamError          = "Param Error"
)

func GetErrCodeMsg(code msg.ErrCode) string {
	return msg.ErrCode_name[int32(code)]
}

func GetMsgByCode(code int) string {
	return msg.ErrCode_name[int32(code)]
}
