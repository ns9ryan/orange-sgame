package botctrlservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
)

const botNum = 1

type BotCtrlService struct {
	service.Service
}

func init() {

	node.Setup(&BotCtrlService{})

	InitAnswersMapBacx()

}

func (ts *BotCtrlService) OnInit() error {

	ts.OpenConcurrent(1, botNum, 10)

	iConfig := ts.GetService().GetServiceCfg()
	if iConfig == nil {
		return fmt.Errorf("%s config is error", ts.GetService().GetName())
	}
	mapTcpCfg := iConfig.(map[string]interface{})

	fmt.Println(mapTcpCfg)

	//_, wsOk := mapTcpCfg["gateAddr"]
	//if wsOk == true {
	//	gateAddr = mapTcpCfg["gateAddr"].(string)
	//} else {
	//	return errors.New("WSCfg  are not configured")
	//}

	_, wsOk2 := mapTcpCfg["wsAddr"]

	if wsOk2 == true {
		wsAddr = mapTcpCfg["wsAddr"].(string)
	} else {
		return errors.New("WSCfg  are not wsAddr  configured")
	}

	_, wsOk3 := mapTcpCfg["gameId"]

	if wsOk3 == true {
		gameId = mapTcpCfg["gameId"].(string)
	} else {
		//return errors.New("WSCfg  are not gameId  configured")
	}

	return nil
}

func (ts *BotCtrlService) OnStart() {
	for i := 0; i < botNum; i++ {
		var b Bot
		b.Init(i)

		ts.AsyncDo(b.runBot, nil)
	}

}
