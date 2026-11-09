package main

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/cluster"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "videogame/common/betodds"
	_ "videogame/common/bootstrap"
	"videogame/common/msg_struct"
	"videogame/common/msg_struct/msg_ctrl"
	"videogame/common/systemdata"
	"videogame/common/util"
	_ "videogame/service/botctrlservice"
	_ "videogame/service/botservice"
	_ "videogame/service/collectservice"
	_ "videogame/service/games/bacxservice"
	_ "videogame/service/games/betService"
	_ "videogame/service/games/coreservice"
	_ "videogame/service/games/fscservice"
	_ "videogame/service/games/ftanservice"
	_ "videogame/service/games/lobbyservice"
	_ "videogame/service/games/longhuservice"
	"videogame/service/games/player"
	_ "videogame/service/games/rotservice"
	_ "videogame/service/games/sdieservice"
	_ "videogame/service/games/sicboservice"
	_ "videogame/service/gateservice"
	_ "videogame/service/httpgateservice"
	_ "videogame/service/merchantapiservice"
	_ "videogame/service/singlewalletservice"
	_ "videogame/service/walletservice"
)

// 初始化一些全局变量
func InitGlobalData() {

	//初始化服务名列表，这个一定要初始化
	util.InitServiceName()

	//初始化全局抽码列表
	systemdata.InitChipsConfig()

	//初始化商户代理
	player.InitMerchantGroups()

}

// 1111 2222
func main() {

	//	TestMsgToJson()

	fmt.Println("init orange-same....")

	// 打开性能分析报告功能，并设置10秒汇报一次
	node.OpenProfilerReport(time.Second * 10)

	InitGlobalData()

	// 创建信号通道，用于接收系统信号
	stopChan := make(chan os.Signal, 1)

	// 捕获系统信号（Ctrl+C 或终止信号）
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// 启动节点服务
	go func() {
		node.Start()
		log.Errorf("Ctrl OnGameOpenAnswer  To Service Error!!!", "---test")
	}()

	//打开性能分析报告功能，并设置10秒汇报一次
	//node.OpenProfilerReport(time.Second * 10)
	//node.Start()

	// 等待信号
	<-stopChan

	// 停止服务
	cluster.GetCluster().Stop()

}

func TestMsgToJson() {

	//	var card msg_struct.DTXAnswerCard
	//	cardcont, _ := json.Marshal(&card)
	//	fmt.Println("cardCont", cardcont)
	var answerdtx msg_struct.GameOpenAnswerDTX
	answerdtx.Set = 10
	answerdtx.Round = 100
	answerdtx.DragonCard.Color = 1
	answerdtx.DragonCard.Number = 1
	answerdtx.TigerCard.Color = 2
	answerdtx.TigerCard.Number = 2

	cardcont, _ := json.Marshal(&answerdtx)

	fmt.Println("cardCont", cardcont)

	var loginres = msg_ctrl.CtrlLoginReq{}

	loginres.Room = "6"
	loginres.UserId = "abcde"
	loginres.GameName = "HongHu"

	ls, _ := json.Marshal(&loginres)

	fmt.Println("ls:", ls)

	ctrlresp := msg_ctrl.CtrlLoginResponse{}

	ctrlresp.Level = 1
	ctrlresp.Avator = "xxxvvv"
	ctrlresp.UserId = "90001"
	ctrlresp.UserName = "Hello"

	es, _ := json.Marshal(&ctrlresp)

	fmt.Println("es:", es)

	var retMsg msg_struct.RetMsg

	retMsg.Code = 1
	retMsg.Msg = "ok "
	retMsg.Data = ctrlresp

	ret, _ := json.Marshal(&retMsg)

	fmt.Println("ret:", ret)

	var cardanswer msg_struct.GameOpenAnswerDTX

	cardanswer.Set = 1

	card, _ := json.Marshal(&cardanswer)

	fmt.Println("card:", card)

	gameconfig := msg_struct.DTXToCtrlGameConfig{}

	config, _ := json.Marshal(&gameconfig)

	fmt.Println("gameconfig:", config)

	gamestatus := msg_struct.GameDtxAnswersPeriodic{}

	gamestatus.Winners = []int{1, 2, 0, 1, 1}

	gamestatus.WinnerNum = len(gamestatus.Winners)

	status, _ := json.Marshal(&gamestatus)

	fmt.Println("gamestatus:", status)

	var gameTime msg_struct.GameServerTime

	nowTime := time.Now()
	gameTime.ServerTime = nowTime.Format("2006-01-02 15:04:05")

	status1, _ := json.Marshal(&gameTime)

	fmt.Println("gamestatus:", status1)
}
