package adminmodule

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"runtime"
	"strconv"
	"sync"
	"time"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/msg_struct/msg_ctrl"
	netModule "videogame/common/netmodule"
	"videogame/common/util"
	"videogame/service/interfacedef"
)

type IAdmin interface {
}

type MsgEvent struct {
	cliId string

	conn *network.NetConn
	msg  interface{}

	funcDo func(cliId string, msg interface{})
}

type IAdminCtrl interface {

	//注册自定义的消息
	RegisterCtrlMsg()

	//广播游戏状态
	BroadStatusMsg(statusMsg *msg_struct.GameStatusMsg)
}

type AdminControl struct {
	service.Module

	gs interfacedef.IGSService

	netModule netModule.INetModule
	adminImp  *AdminImp //管理控端的管理器

	gameCtrl IAdminCtrl //每个游戏的控制接收器

	rawPackInfo processor.PBRawPackInfo

	//控端的列表
	adminMutex sync.RWMutex

	mapAdmin map[string]*Admin

	//消息注册器
	mapRegisterMsg map[int]*MsgEvent

	//消息队列
	chanMsg chan *MsgEvent

	bStopRun bool //退出标志

	ReportTimerId uint64 //定时报告器
}

func (mr *AdminControl) GetGs() interfacedef.IGSService {
	return mr.gs
}

// 消息处理函数注册
func (mr *AdminControl) RegMsgHandler(msgType int, msg interface{}, funcDo func(clientId string, msg interface{})) {
	mr.mapRegisterMsg[msgType] = &MsgEvent{msg: msg, funcDo: funcDo}
}

func (mr *AdminControl) NewMsgByMsgType(msgType int, cliId string) *MsgEvent {

	msg, ok := mr.mapRegisterMsg[msgType]

	if ok == false {
		return nil
	}

	//这边需要优化，后面有时间再搞，要用内存分配方式
	var msgEvent MsgEvent

	msgEvent.cliId = cliId
	msgEvent.msg = msg.msg
	msgEvent.funcDo = msg.funcDo

	return &msgEvent
}

func (mr *AdminControl) Init(process processor.IRawProcessor, is interfacedef.IGSService, gmCtrl IAdminCtrl) {

	process.SetConnectedHandler(mr.OnConnected)
	process.SetDisConnectedHandler(mr.OnDisconnected)
	process.SetRawMsgHandler(mr.RouterMessage)

	mr.gs = is

	mr.gameCtrl = gmCtrl
}

func (mr *AdminControl) OnInit() error {

	mr.chanMsg = make(chan *MsgEvent, 20)
	mr.mapRegisterMsg = make(map[int]*MsgEvent)

	//不用分配太多，这时控端，很少人登陆
	mr.mapAdmin = make(map[string]*Admin, 10)

	pservice := mr.GetService()
	servername := ""

	if pservice != nil {
		servername = pservice.GetName()
	}
	mr.adminImp = NewAdminImp(servername)

	//这个一定不要少，不然内存不分配
	mr.AddModule(mr.adminImp)

	//登陆处理
	mr.RegMsgHandler(msg_struct.LoginReqCtrlProtocol, &msg_ctrl.CtrlLoginReq{}, mr.OnGameLoginMessage)

	//ping处理
	mr.RegMsgHandler(msg_struct.Ping, &msg_struct.MsgPing{}, mr.OnGamePingMessage)

	//游戏状态
	mr.RegMsgHandler(msg_struct.GameStatusChangeProtocol, &msg_struct.GameStatusMsg{}, mr.OnGameChangeStatusMessage)

	//荷官加入
	mr.RegMsgHandler(msg_struct.GameDealerJoinProtocol, &msg_struct.GameDealerJoin{}, mr.OnGameDealerJoin)

	if mr.gameCtrl != nil {
		//自定议的消息
		mr.gameCtrl.RegisterCtrlMsg()
	}

	mr.bStopRun = false

	go mr.runLoop()

	return nil
}

func (mr *AdminControl) SetNetModule(netModule netModule.INetModule) {
	mr.netModule = netModule
}

func (mr *AdminControl) Start() error {

	return mr.netModule.Start()
}

func (mr *AdminControl) OnDisconnected(clientId string) {

	fmt.Printf("AdminControl MsgRoute disconnect clientId:%v\n ", clientId)

	mr.adminMutex.Lock()

	admin, ok := mr.mapAdmin[clientId]
	delete(mr.mapAdmin, clientId)

	mr.adminMutex.Unlock()

	if ok == true {
		mr.adminImp.ReleaseAdmin(admin.Id)
	}

	fmt.Printf("控端下线了。。。。。 %s \n", clientId)

}

func (mr *AdminControl) OnConnected(clientId string) {

	fmt.Printf("AdminControl OnConnected clientId:%v\n", clientId)

	clientIP := mr.netModule.GetClientIp(clientId)

	//检测白名单IP
	if clientIP == "" {
		mr.netModule.Close(clientId)
	}

}

func (mr *AdminControl) tickerLoadAreaInfo(timerId uint64, face interface{}) {
	mr.adminImp.ReportInfo()
}

func (mr *AdminControl) runLoop() {

	defer func() {
		if r := recover(); r != nil {

			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.SError("core dump info:", err.Error(), "\n")
		}
	}()

	mr.SafeNewTicker(&mr.ReportTimerId, time.Second*30, nil, mr.tickerLoadAreaInfo)

	for {
		cxt, cancel := context.WithTimeout(context.Background(), time.Second*5)

		select {
		case <-cxt.Done():
			cancel()
		case receiveMsg := <-mr.chanMsg:
			if receiveMsg.funcDo != nil {
				receiveMsg.funcDo(receiveMsg.cliId, receiveMsg.msg)
			}
		}

		if mr.bStopRun == true {
			break
		}

	}

}

// RouterMessage 函数返回后，msgBuff内存将被内存池回收, 处理控端发送上来的消息,然后转发给其它服务器
func (mr *AdminControl) RouterMessage(cliId string, msgType uint16, msgBuff []byte) {

	fmt.Printf("RouterMessage :clientid:%s   msgType:%d \r\n  ", cliId, msgType)

	//先检测是否登陆消息
	if msgType != msg_struct.LoginReqCtrlProtocol {

		//若不是登陆消息，需要检测是否登陆通过了,每个消息都要检测，以后需要优化
		ret := mr.AdminIsLogin(cliId)

		if ret == false {

			mr.netModule.Close(cliId)
			return
		}
	}

	//处理登陆后的消息
	msgEvent := mr.NewMsgByMsgType(int(msgType), cliId)

	if msgEvent == nil {

		mr.netModule.Close(cliId)
		return
	}

	//解析到实际的消息结构体中
	err := json.Unmarshal(msgBuff[2:], msgEvent.msg)

	if err != nil {
		mr.netModule.Close(cliId)
		return
	}

	//消息入队列
	mr.chanMsg <- msgEvent

}

func (self *AdminControl) GetAllClientId() []string {

	var clientIds []string

	self.adminMutex.RLock()

	for clientId, _ := range self.mapAdmin {
		clientIds = append(clientIds, clientId)
	}

	self.adminMutex.RUnlock()

	return clientIds
}

func (mr *AdminControl) SendMsg(clientId string, msgType int, msg interface{}) error {

	byteMsg, err := json.Marshal(msg)

	if err != nil {
		log.SError("SendMsg fail", log.String("clientId", clientId), log.ErrorField("error", err), log.String("msgType", fmt.Sprint(msgType)))
		return err
	}

	mr.rawPackInfo.SetPackInfo(uint16(msgType), byteMsg)

	bytes, err := mr.netModule.GetProcessor().Marshal(clientId, &mr.rawPackInfo)

	if err != nil {
		return err
	}

	err = mr.netModule.SendRawMsg(clientId, bytes)

	if err != nil {
		log.Error("SendMsg fail", log.String("clientId", clientId), log.ErrorField("error", err), log.String("msgType", fmt.Sprint(msgType)))
	}

	return err
}

// 广播消息
func (mr *AdminControl) BroadCastMsg(msgType int, msg interface{}) {

	cliIdList := mr.GetAllClientId()

	if len(cliIdList) == 0 {
		return
	}

	byteMsg, err := json.Marshal(msg)

	if err != nil {
		log.SError("SendMsg fail", log.ErrorField("error", err), log.String("msgType", fmt.Sprint(msgType)))

		return
	}

	mr.rawPackInfo.SetPackInfo(uint16(msgType), byteMsg)

	for _, cliId := range cliIdList {

		bytes, err := mr.netModule.GetProcessor().Marshal(cliId, &mr.rawPackInfo)

		if err != nil {
			return
		}

		err = mr.netModule.SendRawMsg(cliId, bytes)

		if err != nil {
			log.Error("SendMsg fail", log.String("clientId", cliId), log.ErrorField("error", err), log.String("msgType", fmt.Sprint(msgType)))
		}
	}

}

// 检测控端是否合法,若认证没有过，任何消息都是非法的
func (mr *AdminControl) AdminIsLogin(clientId string) bool {

	//需要优化
	mr.adminMutex.RLock()
	_, ok := mr.mapAdmin[clientId]
	mr.adminMutex.RUnlock()

	return ok

}

func (mr *AdminControl) addAdminInfo(cliId string, admin *Admin) {

	mr.adminMutex.Lock()
	mr.adminMutex.Unlock()

	mr.mapAdmin[cliId] = admin

}

func (mr *AdminControl) GetClientNum() int {

	mr.adminMutex.RLock()
	mr.adminMutex.RUnlock()

	return len(mr.mapAdmin)
}

func (mr *AdminControl) Fail(cliId string, msgType int, code int, message string) {

	var retMsg msg_struct.RetMsg
	retMsg.Code = code
	retMsg.Msg = message

	errSend := mr.SendMsg(cliId, msgType, &retMsg)
	if errSend != nil {
		log.SError("gatectrlservice func:Fail SendMsg err:", errSend.Error())
	}
}

func (mr *AdminControl) Success(cliId string, msgType int, data interface{}) {

	var retMsg msg_struct.RetMsg

	retMsg.Code = msg_struct.Ok
	retMsg.Msg = msg_struct.OkMsg
	retMsg.Data = data

	errSend := mr.SendMsg(cliId, msgType, &retMsg)
	if errSend != nil {
		log.SError("gatectrlservice func:Success SendMsg err:", errSend.Error())
	}
}

// ping 心跳返回
func (mr *AdminControl) Ping(cliId string) {

	retMsg := msg_struct.MsgPong{

		NowTime: util.GetNowDatetime(),
	}
	err := mr.SendMsg(cliId, msg_struct.Pong, &retMsg)

	if err != nil {
		log.SError("ping SendMsg error:", err.Error())
		return
	}
}

func (self *AdminControl) AddAdmin(cliId string, userId string, ip string) bool {

	fmt.Printf(" AdminControl ... RPCLogin  %v  \n", userId)

	p := self.adminImp.GetAdmin(userId)

	if p == nil { //创建用户

		p = self.adminImp.NewAdmin(userId, self.gs, ip)

		self.attachAdmin(cliId, userId, p)

	} else {

		if p.GetClientId() != "" {
			delete(self.mapAdmin, p.GetClientId())
		}
		p = self.adminImp.NewAdmin(userId, self.gs, ip)

		self.attachAdmin(cliId, userId, p)
	}

	self.addAdminInfo(cliId, p)

	//荷管加入
	//	self.DealerJoin(req.UserId, loginRes)

	return true
}

// 返回给控端的不成功的信息
func (mr *AdminControl) handleLoginFail(cliId string, message string) {

	var retMsg msg_struct.RetMsg

	retMsg.Msg = message
	retMsg.Code = msg_struct.Fail

	errSend := mr.SendMsg(cliId, msg_struct.LoginResCtrlProtocol, &retMsg)

	if errSend != nil {
		log.SError("gate handleLoginError SendMsg err:", errSend.Error())
	}

	log.SError(" login cliId error :", cliId)
}

func (mr *AdminControl) handleLoginSuccess(cliId string, data interface{}) {

	var retMsg msg_struct.RetMsg

	retMsg.Msg = msg_struct.OkMsg
	retMsg.Code = msg_struct.Ok
	retMsg.Data = data

	errSend := mr.SendMsg(cliId, msg_struct.LoginResCtrlProtocol, &retMsg)

	fmt.Printf("send retmsg code to %d  client  \n ", retMsg.Code)

	if errSend != nil {
		log.SError("gate handleLoginError SendMsg err:", errSend.Error())
	}
}

func (mr *AdminControl) attachAdmin(cliId string, userId string, admin *Admin) {

	if admin.GetClientId() != "" {
		delete(mr.mapAdmin, admin.GetClientId())
	}

	mr.mapAdmin[cliId] = admin

	//modify coolz
	mr.mapAdmin[cliId].UserId = userId

	admin.AttachConn(cliId, "", "")

}

// 检测控端帐号合法性
func (mr *AdminControl) CheckAccountIsValid(userId string, password string) bool {

	//读取数据库的控端帐号
	Adminer, err := models.GetAdminByUid(userId)

	if err != nil { //测试时暂时关闭
		return false
	}

	//检测帐号密码是正确
	if util.BcryptCheck(password, Adminer.Password) == false {
		return false
	}

	return true

}

// 处理控端登陆消息
func (mr *AdminControl) OnGameLoginMessage(cliId string, msg interface{}) {

	loginReq, ok := msg.(*msg_ctrl.CtrlLoginReq)

	if ok == false {
		mr.netModule.Close(cliId)
		return
	}

	//	fmt.Printf(" gamename:%s room:%s", loginReq.GameName, loginReq.Room)
	//解析处理Token
	userId := loginReq.UserId
	password := loginReq.Token
	ipaddr := mr.netModule.GetClientIp(cliId)

	//帐号不能出空
	if userId == "" {

		//测试阶段放开不返回
		mr.handleLoginFail(cliId, "admin empty  error!!!")
		mr.netModule.Close(cliId)

		return
	}

	//检测控端帐号是否正确
	if mr.CheckAccountIsValid(userId, password) == false {

		log.Errorf("控端帐号：adminId:%s  , password:%s 不对！！！！", userId, password)

		//存储控端进入日志
		var loginTable models.ControlLoginLog

		nowTime := time.Now()

		loginTable.Username = loginReq.UserId
		loginTable.LoginTime = &nowTime
		loginTable.LoginIp = ipaddr
		loginTable.IsOk = 0
		loginTable.Reason = "passsword err"

		loginTable.Create()

		//测试阶段放开不返回
		//	mr.handleLoginFail(cliId, "password  error!!!")
		//	return
	}

	if mr.adminImp.AdminIsExist(userId) == true {

		mr.handleLoginFail(cliId, "admin had relogin !!!")

		//清除一下原来的帐号
		mr.adminImp.ReleaseAdmin(userId)

		//存储控端进入日志
		var loginTable models.ControlLoginLog

		nowTime := time.Now()

		loginTable.Username = loginReq.UserId
		loginTable.LoginTime = &nowTime
		loginTable.LoginIp = ipaddr
		loginTable.IsOk = 0
		loginTable.Reason = "relogin"

		loginTable.Create()

		return
	}

	//解析出荷官的ID
	loginReq.UserId = userId

	//请求控端服务结构
	var req msg_ctrl.LoginToCtrlService

	req.UserId = loginReq.UserId
	req.ClientId = cliId
	req.GateNodeId = node.GetNodeId()
	req.Ip = ipaddr

	//得到服务名跟函数名称
	srName := mr.gs.GetName()

	rpcLoginFun := util.FormatRpcMethod(srName, "RPCAdminLogin")

	//向游戏服发送荷官进入游戏的消息，同
	err := mr.gs.AsyncCall(rpcLoginFun, &req, func(dealer *msg_struct.DealerInfo, err error) {

		if err != nil {

			//存储控端进入日志
			var loginTable models.ControlLoginLog

			nowTime := time.Now()

			loginTable.Username = loginReq.UserId
			loginTable.LoginTime = &nowTime
			loginTable.LoginIp = req.Ip
			loginTable.IsOk = 0
			loginTable.Reason = err.Error()

			loginTable.Create()

			message := "登录控端失败,请稍后再试!"

			mr.handleLoginFail(cliId, message)

			return
		}

		//fmt.Printf(" AsyncCallNode Login .... callback ok  userid :%s   username :%s level:%d  \r\n", dealer.DealerId, dealer.Name, dealer.Level)

		//存储用户
		mr.AddAdmin(cliId, req.UserId, req.Ip)

		//登录成功
		mr.handleLoginSuccess(cliId, dealer.DealerId)

		//通知处理发送游戏配置等消息
		var loginOK msg_ctrl.CtrlLoginSuccess

		rpcLoginFun = util.FormatRpcMethod(srName, "RPCAdminLoginOk")

		loginOK.UserId = req.UserId
		loginOK.ClientId = req.ClientId

		mr.gs.AsyncCall(rpcLoginFun, &loginOK, func(ret *int, err1 error) {

			if err1 != nil {
				return
			}

			//存储控端进入日志
			var loginTable models.ControlLoginLog

			nowTime := time.Now()

			loginTable.Username = loginReq.UserId
			loginTable.LoginTime = &nowTime
			loginTable.LoginIp = req.Ip
			loginTable.IsOk = 1
			loginTable.Reason = "deskno:" + strconv.Itoa(*ret)

			loginTable.Create()

			fmt.Printf("发送给控端的消息成功了 ok:%v \r\n", ret)
		})

	})

	if err != nil {

		message := "登录控端失败,请稍后再试!xxxxxx"

		mr.handleLoginFail(cliId, message)
		return
	}

}

func (mr *AdminControl) OnGamePingMessage(cliId string, msg interface{}) {

	retMsg, ok := msg.(*msg_struct.MsgPing)

	if ok == false {
		mr.netModule.Close(cliId)
		return
	}

	//fmt.Printf("msg ping NowTime:%v\n", msg.NowTime)
	fmt.Printf("ctrl --> server OnGamePingMessage ：%s \n", retMsg.NowTime)

	mr.Ping(cliId)
}

// 处理游戏状态的消息
func (mr *AdminControl) OnGameChangeStatusMessage(cliId string, msg interface{}) {

	gameStatus, ok := msg.(*msg_struct.GameStatusMsg)

	if ok == false {
		mr.netModule.Close(cliId)
		return
	}

	//fmt.Printf("ctrl --> server GameStatusMessage ：%d \n", gameStatus.GameStatus)

	var retMsg msg_struct.GameStatusMsg

	rpcFunc := util.FormatRpcMethod(mr.gs.GetName(), "RPCCtrlChangeGameStatus")

	//本地服务调用
	err := mr.gs.Call(rpcFunc, gameStatus, &retMsg)

	if err != nil {
		//发送给控端通知不成功
		var failedMsg msg_struct.GameChangeFailed

		failedMsg.Msgid = msg_struct.GameStatusChangeProtocol

		failedMsg.Round = retMsg.Round
		failedMsg.Set = retMsg.Set
		failedMsg.GameId = retMsg.GameId
		failedMsg.DeskNo = retMsg.DeskNo

		mr.SendMsg(cliId, msg_struct.GameMsgFailedProtocol, &failedMsg)

		return

	} else if retMsg.GameStatus == msg_struct.GameStatusError { //检测状态是否更新成功

		var failedMsg msg_struct.GameChangeFailed

		failedMsg.Msgid = msg_struct.GameStatusChangeProtocol

		failedMsg.Round = retMsg.Round
		failedMsg.Set = retMsg.Set
		failedMsg.GameId = retMsg.GameId
		failedMsg.DeskNo = retMsg.DeskNo

		mr.SendMsg(cliId, msg_struct.GameMsgFailedProtocol, &failedMsg)

		//fmt.Printf("****ctrl --> server GameStatusMessage ：%d  failed!!!! \n", retMsg.GameStatus)

		return
	} else {

		//fmt.Println("ctrl --> server OnGameChangeStatusMessage-->BroadCastMsg ")

	}

	if mr.gameCtrl != nil {

		mr.gameCtrl.BroadStatusMsg(&retMsg)
	}

	//mr.BroadCastMsg(msg_struct.GameBacxStatusProtocol, &gameStatus)
	//mr.BroadCastMsg(msg_struct.GameBacxStatusProtocol, &retMsg)

	//fmt.Println("ctrl --> server OnGameChangeStatusMessage-->BroadCastMsg ")

}

// 荷官上班
func (mr *AdminControl) OnGameDealerJoin(cliId string, msg interface{}) {

	dealerMsg, ok := msg.(*msg_struct.GameDealerJoin)

	if ok == false {
		return
	}

	rpcFun := util.FormatRpcMethod(mr.gs.GetName(), "RPCCtrlDealerJoinRoom")

	var retDealer msg_ctrl.CtrlDealerInfo

	//通知游戏，荷管加入
	err := mr.gs.Call(rpcFun, &dealerMsg.UserId, &retDealer)

	var joinret msg_struct.GameDealerJoinResult

	joinret.RetCode = "ok"

	if err != nil || retDealer.Success == false {
		joinret.RetCode = "failed"

	} else {

		joinret.UserId = retDealer.DearlerId
		joinret.UserName = retDealer.Name
		//joinret.Supervisor = retDealer     //需要增加上级主管
	}

	//荷官加入的回复
	mr.SendMsg(cliId, msg_struct.GameDealerJoinProtocolResult, &joinret)

}
