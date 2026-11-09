package coreservice

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"videogame/common/msg_struct_crm"
)

// Fail 用于返回失败响应
func ResponseFail(c *gin.Context, msgId uint, reqId uint, errCode int) {

	var RetMsg msg_struct_crm.ResponseMsg

	RetMsg.ResponseId = reqId
	RetMsg.MsgId = msgId
	RetMsg.Success = 0
	RetMsg.ErrCode = errCode

	c.JSON(http.StatusOK, gin.H{
		"Ret":  0,
		"Data": RetMsg,
	})

	return
}

// Fail 用于返回失败响应
func ResponseSuccess(c *gin.Context, msgId uint, reqId uint) {

	var RetMsg msg_struct_crm.ResponseMsg

	RetMsg.ResponseId = reqId
	RetMsg.MsgId = msgId
	RetMsg.Success = 1
	RetMsg.ErrCode = 0

	c.JSON(http.StatusOK, gin.H{
		"Ret":  1,
		"Data": RetMsg,
	})

	return
}
