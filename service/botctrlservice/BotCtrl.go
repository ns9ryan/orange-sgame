package botctrlservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/gorilla/websocket"
	"math/rand"
	"sync"
	"time"
	"videogame/common/msg_struct"
	"videogame/common/msg_struct/msg_ctrl"
)

type botStatus = int

type NetType int

var gateAddr = "127.0.0.1:9001"        //默认
var wsAddr = "ws://192.168.0.117:9556" //默认
var gameId = msg_struct.GAMEID_BACXA

//const gameId = msg_struct.GAMEID_BACXA
//const gameId = msg_struct.GAMEID_BACXA

const (
	Tcp NetType = iota
	WS
	Kcp
)

var netType = WS

const (
	httpLogging       = 0
	gateLogging       = 1
	receivingUserInfo = 2
	finish            = 3
)

type IConn interface {
	ReadMsg() ([]byte, error)
	WriteMsg(args ...[]byte) error
}

type Bot struct {
	id    int
	token string

	status botStatus

	GameStatus int
	Set        int
	Round      int

	SleepNum int

	GameId string
	DeskNo int

	tcpClient *network.TCPClient
	wsClient  *network.WSClient
	kcpClient *network.KCPClient

	tcpConn *network.NetConn
	wsConn  *network.WSConn
	kcpConn *network.NetConn

	conn      IConn
	connMutex sync.Mutex

	chanMsg        chan *MsgEvent
	pbRawProcessor processor.PBRawProcessor
	timeAfter      time.Timer
}

var mapRegisterMsg map[int]*MsgEvent

type MsgEvent struct {
	conn *network.NetConn
	msg  interface{}

	funcDo func(bot *Bot, msg interface{})
}

func regMsg(msgType int, msg interface{}, funcDo func(bot *Bot, msg interface{})) {
	mapRegisterMsg[msgType] = &MsgEvent{msg: msg, funcDo: funcDo}
}

func NewMsgByMsgType(msgType int) *MsgEvent {

	msg, ok := mapRegisterMsg[msgType]

	if ok == false {
		return nil
	}

	var msgEvent MsgEvent
	msgEvent.msg = msg.msg
	msgEvent.funcDo = msg.funcDo

	return &msgEvent
}

func (bt *Bot) runBot() bool {

	for {
		cxt, cancel := context.WithTimeout(context.Background(), time.Second*5)
		select {
		case <-cxt.Done():
			cancel()
			bt.doTimeOut()

			fmt.Printf("doTimeOut ...\n")
		case receiveMsg := <-bt.chanMsg:
			if receiveMsg.funcDo != nil {
				receiveMsg.funcDo(bt, receiveMsg.msg)
			}
		}
	}
	return true
}

func (bt *Bot) doTimeOut() {

	switch bt.status {
	case httpLogging:
		bt.HttpLogin()
	case gateLogging:
		bt.gateLogin()
	case finish:
		bt.work()
	}

	fmt.Printf("do TimeOut ...")
}

func (bt *Bot) setTcpConn(conn *network.NetConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	bt.conn = conn
	bt.tcpConn = conn
}

func (bt *Bot) setWsConn(conn *network.WSConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	bt.conn = conn
	bt.wsConn = conn
}

func (bt *Bot) setKcpConn(conn *network.NetConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	bt.conn = conn
	bt.kcpConn = conn
}

func (bt *Bot) getTcpConn() *network.NetConn {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	return bt.tcpConn
}

func (bt *Bot) getWsConn() *network.WSConn {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	return bt.wsConn
}

func (bt *Bot) getConn() IConn {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	return bt.conn
}

func (bt *Bot) Init(id int) {
	bt.id = id
	bt.chanMsg = make(chan *MsgEvent, 100)
}

func (bt *Bot) gateLogin() {

	if bt.initConnect() == false {
		time.Sleep(5 * time.Second)
		return
	}

	//发送消息
	time.Sleep(1 * time.Second)

	if bt.getConn() == nil {
		return
	}
	//fmt.Printf("gateLogin enter gateCtrlService time=%v\n", util.GetNowDatetime())

	var req msg_ctrl.CtrlLoginReq

	req.Token = "abc123" //bt.token
	req.UserId = "admin3"
	req.GameName = gameId
	req.Room = "1"

	err := bt.SendMsg(msg_struct.LoginReqCtrlProtocol, &req)

	if err != nil {

		fmt.Printf("don't connect ")
		return
	}

	bt.setLoginFinish()
}

func (bt *Bot) work() {

	//	var ping msg_struct.MsgPing
	//	ping.NowTime = time.Now().String()
	//	bt.SendMsg(msg_struct.Ping, &ping)

}

func (bt *Bot) StartBet() {

	var changeStatus msg_struct.GameStatusMsg

	changeStatus.GameStatus = msg_struct.GameStatusBet

	changeStatus.Set = bt.Set
	changeStatus.Round = bt.Round
	changeStatus.DeskNo = bt.DeskNo
	changeStatus.GameId = bt.GameId

	bt.SendMsg(msg_struct.GameStatusChangeProtocol, &changeStatus)

	fmt.Printf("startbet ... \n")
}

func (bt *Bot) OpenAnswerDTX() {

	var openan msg_struct.GameOpenAnswerDTX

	openan.Set = bt.Set
	openan.Round = bt.Round
	openan.GameId = bt.GameId
	openan.DeskNo = bt.DeskNo

	openan.Winner = 1

	openan.DragonCard.Used = 1
	openan.DragonCard.Number = byte(1 + (rand.Uint64()+uint64(time.Now().Unix()))%13)
	openan.DragonCard.Color = byte(rand.Uint64() % 4)

	openan.TigerCard.Used = 1
	openan.TigerCard.Number = byte(1 + (rand.Uint64()*3+uint64(time.Now().Unix()*2))%13)
	openan.TigerCard.Color = byte(rand.Uint64() % 4)

	result := int(openan.DragonCard.Number - openan.TigerCard.Number)

	switch result {
	case 0:
		openan.Winner = 0 //龙胜
	case 1:
		openan.Winner = 1 //庄胜
	case 2:
		openan.Winner = 2 //和局
	}

	bt.SendMsg(msg_struct.GameDtxOpenAnswerProtocol, &openan)

}

func (bt *Bot) OpenAnswerBacx() {

	var openan msg_struct.GameOpenAnswerBacx

	openan.Set = bt.Set
	openan.Round = bt.Round

	//ind := int(time.Now().Unix()+10) % 3

	ind := bt.Round % 3

	openanswer := AnswersMapBacx[ind]

	openan = *openanswer
	openan.Set = bt.Set
	openan.Round = bt.Round

	openan.GameId = bt.GameId
	openan.DeskNo = bt.DeskNo

	bt.SendMsg(msg_struct.GameBacxOpenAnswerProtocol, &openan)

	fmt.Println("发送百家乐开奖成功！！！ 。。。。")

}

func (bt *Bot) OpenAnswerSicbo() {

	var openan msg_struct.SicboOpenAnswer

	openan.Set = bt.Set
	openan.Round = bt.Round
	//openan.Winner = int(time.Now().Unix()) % 10
	openan.GameId = bt.GameId
	openan.DeskNo = 6
	openan.Winner = int(time.Now().Unix()) % 50

	bt.SendMsg(msg_struct.GameSicboOpenAnswerProtocol, &openan)

}

func (bt *Bot) OpenAnswerFtan() {

	var openan msg_struct.FtanOpenAnswer

	openan.Set = bt.Set
	openan.Round = bt.Round
	openan.GameId = bt.GameId
	openan.DeskNo = 7
	openan.Winner = rand.Intn(10)%4 + 1

	bt.SendMsg(msg_struct.GameFtanOpenAnswerProtocol, &openan)

}

func (bt *Bot) OpenAnswerSdie() {

	var openan msg_struct.SdieOpenAnswer

	openan.Set = bt.Set
	openan.Round = bt.Round
	openan.GameId = bt.GameId
	openan.DeskNo = 8
	openan.Winner = rand.Intn(1000) % 5

	bt.SendMsg(msg_struct.GameSdieOpenAnswerProtocol, &openan)

}

func (bt *Bot) OpenAnswerFSC() {

	var openan msg_struct.FscOpenAnswer

	openan.Set = bt.Set
	openan.Round = bt.Round
	openan.GameId = bt.GameId
	openan.DeskNo = 9
	openan.Winner = rand.Intn(10000000) % 36

	bt.SendMsg(msg_struct.GameFscOpenAnswerProtocol, &openan)

}

// Shuffle 洗牌
func (bt *Bot) Shuffle() {

	var changeStatus msg_struct.GameStatusMsg

	changeStatus.GameStatus = msg_struct.GameStatusInterMission

	changeStatus.Set = bt.Set
	changeStatus.Round = bt.Round
	changeStatus.DeskNo = bt.DeskNo

	bt.SendMsg(msg_struct.GameStatusChangeProtocol, &changeStatus)

	fmt.Printf("发送洗牌状态成功 ... \n")

}

// 休场
func (bt *Bot) SleepTime() {

	var changeStatus msg_struct.GameStatusMsg

	changeStatus.GameStatus = msg_struct.GameStatusSleep

	changeStatus.Set = bt.Set
	changeStatus.Round = bt.Round
	changeStatus.DeskNo = bt.DeskNo
	changeStatus.GameId = bt.GameId

	bt.SendMsg(msg_struct.GameStatusChangeProtocol, &changeStatus)

	fmt.Printf("发送休场状态成功 ... \n")

}

func (bt *Bot) SendMsg(msgType int, msg interface{}) error {
	var pbPackInfo processor.PBRawPackInfo

	byteMsg, _ := json.Marshal(msg)

	pbPackInfo.SetPackInfo(uint16(msgType), byteMsg)

	bytes, err := bt.pbRawProcessor.Marshal("", &pbPackInfo)

	if err != nil {
		return err
	}

	conn := bt.getConn()
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	return conn.WriteMsg(bytes)
}

func (bt *Bot) HttpLogin() {
	//loginInfo := &ctrl_msg.HttpLoginReq{}
	//loginInfo.UserName = "admin1"
	//loginInfo.PassWord = "333333"
	//var httpClient httpclientmodule.HttpClientModule
	//httpClient.Init("", 1, time.Second*5, time.Second*5, time.Second*5, time.Second*5)
	//byteLogin, _ := json.Marshal(&loginInfo)
	//endpoint := "http://" + gateAddr + "/ctrl/login"
	//response := httpClient.Request("POST", endpoint, byteLogin, nil)
	//log.Debugf("response", log.String("body", string(response.Body)))
	//
	//var httpRespone struct {
	//	Code int
	//	Msg  string
	//	Data ctrl_msg.HttpLoginRet
	//}
	//err := json.Unmarshal(response.Body, &httpRespone)
	//if err != nil {
	//	log.Error("json unmarshal error", log.ErrorField("err", err))
	//	return
	//}
	//if httpRespone.Code == 0 {
	//	bt.setGateLogging(httpRespone.Data.Token)
	//}

	bt.setGateLogging("testest")
}

func (bt *Bot) setGateLogging(token string) {

	bt.status = gateLogging
	bt.token = token
}

func (bt *Bot) setLoginReceive() {
	bt.status = receivingUserInfo
}

func (bt *Bot) setLoginFinish() {

	bt.status = finish

	//time.AfterFunc(time.Second*3, bt.StartBet)

}

func (bt *Bot) dealerJoin() {

	var dealerJoin msg_struct.GameDealerJoin

	dealerJoin.UserId = "D00000005"

	bt.SendMsg(msg_struct.GameDealerJoinProtocol, &dealerJoin)

}

func (bt *Bot) initConnect() bool {

	if bt.tcpClient != nil {
		bt.tcpClient.Close(false)
	}

	if bt.wsClient != nil {
		bt.wsClient.Close()
	}

	switch netType {
	case Tcp:
		bt.tcpClient = &network.TCPClient{}
		bt.tcpClient.Addr = gateAddr
		bt.tcpClient.ConnectInterval = time.Second * 5
		bt.tcpClient.ConnNum = 1
		bt.tcpClient.AutoReconnect = false
		bt.tcpClient.ReadDeadline = time.Second * 600
		bt.tcpClient.WriteDeadline = time.Second * 600

		bt.tcpClient.NewAgent = func(conn *network.NetConn) network.Agent {
			agent := BotAgent{}
			agent.bt = bt

			bt.setTcpConn(conn)
			return &agent
		}

		bt.tcpClient.Start()
	case WS:
		bt.wsClient = &network.WSClient{}
		bt.wsClient.MessageType = websocket.BinaryMessage
		bt.wsClient.MaxMsgLen = 65535
		bt.wsClient.PendingWriteNum = 100
		bt.wsClient.AutoReconnect = false
		bt.wsClient.ConnNum = 1
		bt.wsClient.ConnectInterval = time.Second * 5
		bt.wsClient.Addr = wsAddr
		bt.wsClient.AutoReconnect = true //自动链接

		bt.wsClient.NewAgent = func(conn *network.WSConn) network.Agent {
			agent := BotAgent{}
			agent.bt = bt

			bt.setWsConn(conn)
			return &agent
		}
		bt.wsClient.Start()
	case Kcp:
		bt.kcpClient = &network.KCPClient{}
		bt.kcpClient.Addr = gateAddr
		bt.kcpClient.ConnectInterval = time.Second * 5
		bt.kcpClient.ConnNum = 1
		bt.kcpClient.AutoReconnect = false
		bt.kcpClient.ReadDeadline = time.Second * 600
		bt.kcpClient.WriteDeadline = time.Second * 600
		bt.kcpClient.NewAgent = func(conn *network.NetConn) network.Agent {
			agent := BotAgent{}
			agent.bt = bt

			bt.setKcpConn(conn)
			return &agent
		}

		bt.kcpClient.Start()
	}

	return true
}
