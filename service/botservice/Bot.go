package botservice

import (
	"context"
	"fmt"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/sysmodule/httpclientmodule"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
	betOdds "videogame/common/betodds"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
)

type botStatus = int

const gateAddr = "127.0.0.1:9001"        // kcp,tcp
const wsAddr = "ws://192.168.0.147:9444" // ws

const httpAddr = "http://192.168.0.112:9000/login"

const playerId = "player002"
const playerPass = "123456"

type NetType int

const gameId = msg_struct.GAMEID_BACXA

//const gameId = msg_struct.GAMEID_DTX
//const gameId = msg_struct.GAMEID_SICBO
//const gameId = msg_struct.GAMEID_FTAN

const deskNo = 6

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

	leavegame = 4
)

type IConn interface {
	ReadMsg() ([]byte, error)
	WriteMsg(args ...[]byte) error
}

type Bot struct {
	id     int
	token  string
	status botStatus

	gameId string

	orderkey int64

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
}

var mapRegisterMsg map[msg.MsgType]*MsgEvent

type MsgEvent struct {
	conn   *network.NetConn
	msg    proto.Message
	funcDo func(bot *Bot, msg proto.Message)
}

func regMsg(msgType msg.MsgType, msg proto.Message, funcDo func(bot *Bot, msg proto.Message)) {
	mapRegisterMsg[msgType] = &MsgEvent{msg: msg, funcDo: funcDo}
}

func NewMsgByMsgType(msgType msg.MsgType) *MsgEvent {
	msg, ok := mapRegisterMsg[msgType]
	if ok == false {
		return nil
	}

	var msgEvent MsgEvent

	msgEvent.msg = proto.Clone(msg.msg)
	msgEvent.funcDo = msg.funcDo

	return &msgEvent
}

func (bt *Bot) runBot() bool {

	bt.status = httpLogging

	for {
		cxt, cancel := context.WithTimeout(context.Background(), time.Second*5)
		select {
		case <-cxt.Done():
			cancel()
			bt.doTimeOut()
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
		bt.httpLogin()
	case gateLogging:
		bt.gateLogin()
	case receivingUserInfo:
		time.Sleep(time.Second * 1)

		bt.enterGame()

	case finish:
		//bt.work()
		//bt.userBet()

		time.Sleep(time.Second * 1)

		bt.status = leavegame

	case leavegame:
		//	bt.leaveGame()
		time.Sleep(time.Second * 1)
		//bt.wsClient.Close()

	}
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
	time.Sleep(2 * time.Second)

	if bt.getConn() == nil {
		return
	}

	fmt.Println("bot service connect ...ok!!!!")

	bt.gameId = gameId

	var req msg.LoginReq
	req.Token = bt.token
	req.UserId = playerId // "test" + strconv.Itoa(rand.Int())

	bt.SendMsg(msg.MsgType_LoginRequest, &req)

	bt.setLoginReceive()

}

func (bt *Bot) enterGame() {

	var req msg.JoinToGame

	req.GameId = gameId // msg_struct.GAMEID_BACX
	req.DeskNo = "6"
	req.UserId = playerId

	fmt.Println("bot service JoinToGame ...")

	bt.SendMsg(msg.MsgType_JoinGame, &req)

	time.Sleep(1 * time.Second)

	bt.setLoginFinish()
}

func (bt *Bot) leaveGame() {

	var req msg.LeaveFromGame

	req.GameId = gameId
	req.DeskNo = "1"
	req.UserId = "test001"

	fmt.Println("bot service LeaveToGame ...")

	bt.SendMsg(msg.MsgType_ExitGame, &req)

}

func (bt *Bot) BetBacx() {

	var betmsg msg.CTOS_BetGame

	betmsg.OrderKey = uint32(time.Now().Unix())
	betmsg.BetAreadId = betOdds.Bet_Bacx_Player

	betmsg.Digits = 100
	betmsg.BetPoint = uint32(0.1 * float32(betmsg.Digits))

	betmsg.BetMode = 0
	betmsg.DeskNo = 2
	betmsg.GameId = msg_struct.GAMEID_BACXA

	bt.SendMsg(msg.MsgType_GameTable_Bet, &betmsg)

	//variable.RedisStorage.Set(strconv.Itoa(int(betmsg.OrderKey)), strconv.Itoa(int(betmsg.BetPoint)), 0)

	//time.Sleep(1 * time.Second)

	//betmsg.OrderKey = time.Now().Unix()
	//betmsg.BetAreadId = betOdds.Bet_Bacx_Player
	//betmsg.BetPoint = 10
	//betmsg.BetMode = 0
	//betmsg.DeskNo = 2
	//betmsg.GameId = msg_struct.GAMEID_BACXA
	//
	//bt.orderkey = betmsg.OrderKey

	//bt.SendMsg(msg.MsgType_GameTable_Bet, &betmsg)

	//time.Sleep(1 * time.Second)
	//
	//betmsg.OrderKey = time.Now().Unix()
	//betmsg.BetAreadId = betOdds.Bet_Bacx_Banker
	//betmsg.BetPoint = 10
	//betmsg.BetMode = 0
	//betmsg.DeskNo = 2
	//betmsg.GameId = msg_struct.GAMEID_BACXA
	//
	//bt.orderkey = betmsg.OrderKey
	//
	//bt.SendMsg(msg.MsgType_GameTable_Bet, &betmsg)

}

func (bt *Bot) BetDtx() {

	var betmsg msg.CTOS_BetGame

	betmsg.OrderKey = uint32(time.Now().Unix())
	betmsg.BetAreadId = betOdds.Bet_Dtx_Dragon
	betmsg.Digits = 100

	betmsg.BetPoint = uint32(10 * float32(betmsg.Digits))

	betmsg.BetMode = 0
	betmsg.DeskNo = 4
	betmsg.GameId = msg_struct.GAMEID_DTX

	bt.orderkey = int64(betmsg.OrderKey)

	bt.SendMsg(msg.MsgType_GameTable_Bet, &betmsg)

}

func (bt *Bot) BetSicbo() {

	var betmsg msg.CTOS_BetGame

	betmsg.OrderKey = uint32(time.Now().Unix())
	betmsg.BetAreadId = betOdds.Bet_Sicbo_One_Odd
	betmsg.Digits = 100

	betmsg.BetPoint = uint32(0.1 * float32(betmsg.Digits))

	betmsg.BetMode = 0
	betmsg.DeskNo = 6
	betmsg.GameId = msg_struct.GAMEID_SICBO

	bt.orderkey = int64(betmsg.OrderKey)

	bt.SendMsg(msg.MsgType_GameTable_Bet, &betmsg)

}
func (bt *Bot) BetFtan() {

	var betmsg msg.CTOS_BetGame

	betmsg.OrderKey = uint32(time.Now().Unix())
	betmsg.BetAreadId = betOdds.Bet_Ftan_1Nian3
	betmsg.Digits = 100

	betmsg.BetPoint = uint32(1 * float32(betmsg.Digits))

	betmsg.BetMode = 0
	betmsg.DeskNo = 7
	betmsg.GameId = msg_struct.GAMEID_FTAN

	bt.orderkey = int64(betmsg.OrderKey)

	bt.SendMsg(msg.MsgType_GameTable_Bet, &betmsg)

}

// 玩家下注
func (bt *Bot) userBet() {

	//聊天
	bt.TalkChat()

	if bt.gameId == msg_struct.GAMEID_BACXA ||
		bt.gameId == msg_struct.GAMEID_BACXB ||
		bt.gameId == msg_struct.GAMEID_BACXC {
		//bt.BetBacx()
		//	time.AfterFunc(time.Second*3, bt.CancelBet)

		bt.TipMoney()

	} else if bt.gameId == msg_struct.GAMEID_DTX {

		//bt.BetDtx()
		//	time.AfterFunc(time.Second*3, bt.CancelBet)

		bt.TipMoney()

	} else if bt.gameId == msg_struct.GAMEID_SICBO {

		//bt.BetSicbo()

		bt.TipMoney()

	} else if bt.gameId == msg_struct.GAMEID_FTAN {

		//bt.BetFtan()

		bt.TipMoney()
	}

	bt.status = finish

}

// 玩家下注
func (bt *Bot) CancelBetDTX() {

	var betmsg msg.CTOS_CancelBetGame

	betmsg.OrderKey = uint32(bt.orderkey)

	betmsg.BetMode = 0
	betmsg.DeskNo = 2
	betmsg.GameId = msg_struct.GAMEID_DTX

	bt.SendMsg(msg.MsgType_GameTable_CancelBet, &betmsg)

}

// 玩家下注
func (bt *Bot) CancelBetBacX() {

	var betmsg msg.CTOS_CancelBetGame

	betmsg.OrderKey = uint32(bt.orderkey)

	betmsg.BetMode = 0
	betmsg.DeskNo = 2
	betmsg.GameId = msg_struct.GAMEID_BACXA

	bt.SendMsg(msg.MsgType_GameTable_CancelBet, &betmsg)

}

// 撤消玩家下注
func (bt *Bot) CancelBet() {

	if bt.gameId == msg_struct.GAMEID_BACXA {
		bt.CancelBetBacX()
	} else if bt.gameId == msg_struct.GAMEID_DTX {

		bt.CancelBetDTX()
	}

	bt.status = finish

}

func (bt *Bot) SendSuperSixSet() {

	var superSix msg.CTOS_GameSuperSixSet

	superSix.DeskNo = 1
	superSix.GameId = bt.gameId
	superSix.Supersix = 1

	bt.SendMsg(msg.MsgType_GameTable_SuperSixSet, &superSix)

}

func (bt *Bot) TipMoney() {

	var tipMoney msg.CTOS_TipMoney

	tipMoney.DeskNo = 1
	tipMoney.GameId = bt.gameId

	tipMoney.TipMoney = 10
	tipMoney.Digits = 100

	tipMoney.OrderKey = uint32(time.Now().Unix())

	bt.SendMsg(msg.MsgType_GameTable_TipMoney, &tipMoney)

}

// 打赏
func (bt *Bot) TalkChat() {

	var talkChatMsg msg.CTOS_TalkChat

	talkChatMsg.PlayerId = playerId
	talkChatMsg.GameId = bt.gameId
	talkChatMsg.Deskno = 1
	talkChatMsg.Message = "你妈个屁，都有什么人玩！"

	bt.SendMsg(msg.MsgType_GameTable_TalkChat, &talkChatMsg)

}

var tickNum int

func (bt *Bot) work() {

	var ping msg.MsgPing

	bt.SendMsg(msg.MsgType_Ping, &ping)

	nowTime := time.Now()

	fmt.Println("===>发送玩家心跳给服务器!!! 时间  ", nowTime)

	tickNum++

	if tickNum < 3 {
		time.AfterFunc(time.Second*3, bt.work)
	}

}

func (bt *Bot) SendMsg(msgType msg.MsgType, msg proto.Message) error {
	var pbPackInfo processor.PBRawPackInfo
	byteMsg, _ := proto.Marshal(msg)
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

func (bt *Bot) httpLogin() {

	loginInfo := &msg.HttpLoginReq{}
	loginInfo.UserName = playerId
	loginInfo.PassWord = playerPass

	var httpClient httpclientmodule.HttpClientModule

	httpClient.Init("", 1, time.Second*5, time.Second*5, time.Second*5, time.Second*5)

	byteLogin, _ := json.Marshal(&loginInfo)

	response := httpClient.Request("POST", httpAddr, byteLogin, nil)

	var httpRespone struct {
		Code int
		Msg  string
		Data msg.HttpLoginRet
	}
	err := json.Unmarshal(response.Body, &httpRespone)
	if err != nil {
		//log.Error("json unmarshal error", log.ErrorField("err", err))
		return
	}

	//fmt.Printf("返回客户端【http】 Code=%v,Msg=%v, Data=%v\n", httpRespone.Code, httpRespone.Msg, util.ProtoMessageToMap(&httpRespone.Data))

	if httpRespone.Code == 0 {
		bt.setGateLogging(httpRespone.Data.Token)
	}

	//coolz
	bt.setGateLogging(httpRespone.Data.Token)

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

		fmt.Println("wsconnect ．．．", wsAddr)

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
