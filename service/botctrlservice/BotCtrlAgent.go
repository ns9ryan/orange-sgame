package botctrlservice

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
)

type BotAgent struct {
	bt *Bot
}

func (ba *BotAgent) Run() {

	fmt.Println("BotAgent Run")

	for {
		bytes, err := ba.bt.conn.ReadMsg()

		if err != nil {

			log.Error("rclient read msg is failed", log.ErrorField("error", err))

			//ba.bt.HttpLogin()

			break
		}

		data, err := ba.bt.pbRawProcessor.Unmarshal("", bytes)
		if err != nil {
			log.Error("data error", log.ErrorField("err", err))
			return
		}

		rawPack := data.(*processor.PBRawPackInfo)
		rawPack.GetMsg()

		msgProto := NewMsgByMsgType(int(rawPack.GetPackType()))

		if msgProto == nil {
			//不关注的消息忽略
			log.SWarn("msg type error", log.Any("msgType", rawPack.GetPackType()))
			continue
		}

		err = json.Unmarshal(rawPack.GetMsg()[2:], msgProto.msg)
		if err != nil {
			log.Error("Unmarshal error", log.ErrorField("err", err))
			return
		}

		ba.bt.chanMsg <- msgProto
	}

	fmt.Println("BotAgent Run end ")

}

func (ba *BotAgent) OnClose() {
	ba.bt.tcpClient = nil
	ba.bt.setGateLogging(ba.bt.token)
}
