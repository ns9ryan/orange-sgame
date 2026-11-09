package gateservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/cluster"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	log2 "log"
	"strings"
	"time"
	"videogame/common/global/variable"
	netModule "videogame/common/netmodule"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

type MsgRouter struct {
	service.Module

	netModule netModule.INetModule

	rawPackInfo    processor.PBRawPackInfo
	pbRawProcessor processor.PBRawProcessor

	mapRouterCache map[string]*ClientSessionInfo //用于存储玩家端的数组

	mapGameServer map[string]string //用于存储各个游戏对应的服务名

}

// 每个前端用户相关信息
type ClientSessionInfo struct {
	nodeId    string //gate的节点id
	uid       string //用户的UID,框架低层产生的id
	isLogined bool   //是否登陆成功

	lastMsgTime  int64 //最近一次消息的时间
	sendPingTime int64 //最近发送给大厅ping消息的时间

	loginIp string //登陆的IP

	clientType int    //客户端的设备的类型 Ｈ５，ＡＰＰ，ＩＳＯ
	msgCount   uint64 //消息计数

	lobbyNodeId string //大厅的NodeId
	isBlocked   bool   //是否被锁

}

func (mr *MsgRouter) OnInit() error {

	mr.mapRouterCache = make(map[string]*ClientSessionInfo, 4096)

	mr.setGameServer()

	//启动输出在线玩家列表
	var timerId uint64

	mr.SafeNewTicker(&timerId, time.Second*30, nil, mr.ReportInfo)

	var timerId2 uint64
	mr.SafeNewTicker(&timerId2, time.Second*3, nil, mr.RegisterGateEvent)

	return nil
}

// 设置服务名
func (mr *MsgRouter) setGameServer() {

	mr.mapGameServer = make(map[string]string, 0)

	//游戏服务名称设置
	var gameServiceNameList []util.ServiceNameBridge

	serviceNameListStr := variable.ConfigServiceNameYml.GetString("game_services")

	err := json.Unmarshal([]byte(serviceNameListStr), &gameServiceNameList)

	if err != nil {

		err = errors.New("解析service_name.yml文件key=game_services的内容失败！")

		log.Errorf("setGameServer err:%v", err.Error())

		panic(err)
	}

	for _, sName := range gameServiceNameList {
		mr.AddGameServer(sName.ServiceName)
	}

	// lobbyService betService 公共服务名设置
	var commonServiceNameList []util.ServiceNameBridge

	commonServiceNameListStr := variable.ConfigServiceNameYml.GetString("common_services")

	err = json.Unmarshal([]byte(commonServiceNameListStr), &commonServiceNameList)

	if err != nil {

		err = errors.New("解析service_name.yml文件key=common_services的内容失败！")

		log.Errorf("setGameServer err:%v", err.Error())

		panic(err)
	}

	for _, sName := range commonServiceNameList {

		mr.AddGameServer(sName.ServiceName)
	}

}

func (mr *MsgRouter) ReportInfo(tickId uint64, usned interface{}) {

	log2.Printf("nodeID=%s,    在线人数=%v \n", node.GetNodeId(), mr.GetClientNum())

}

// 向HTTPSERVER 注册网关地址
func (mr *MsgRouter) RegisterGateEvent(timerId uint64, usned interface{}) {

	var req rpc.String

	req.Value = node.GetNodeId()
	rpcFun := util.FormatRpcMethod(util.HttpSerivce, "RPCRegisterGateUrl")

	nodeId := util.GetGameServiceName(util.HttpSerivce)

	err := mr.AsyncCallNode(nodeId, rpcFun, &req, func(ret *rpc.Int, err error) {

		if err != nil {
			log.Errorf("gateService RPCRegisterGateUrl GoNode fail:%s", err.Error())
			return
		}

		if ret.Value == 1 {
			mr.CancelTimerId(&timerId)
		}

	})

	if err != nil {

	}

}

func (mr *MsgRouter) SetNetModule(netModule netModule.INetModule) {
	mr.netModule = netModule
}

func (mr *MsgRouter) Init(process processor.IRawProcessor) {
	process.SetConnectedHandler(mr.OnConnected)
	process.SetDisConnectedHandler(mr.OnDisconnected)
	process.SetRawMsgHandler(mr.RouterMessage)
}

func (mr *MsgRouter) OnDisconnected(clientId string) {

	ipstr := mr.netModule.GetClientIp(clientId)

	fmt.Printf("*******gateservice OnDisconnected IP:%s  clientId:%s \n    ", ipstr, clientId)

	//2.通过clientId获取nodeId
	nodeInf := mr.GetNodeInfo(clientId)

	if nodeInf == nil {
		mr.netModule.Close(clientId)
		return
	}

	//网关服务的nodeId
	nodeId := nodeInf.nodeId
	if nodeId == "" {
		mr.netModule.Close(clientId)
		return
	}

	//通知httpGateService 玩家家人数-1
	mr.updateBalance(-1)

	//2.转发客户端连接断开,
	var rawInputArgs rpc.RawInputArgs

	rawInputArgs.ClientIdList = []string{clientId}
	rawInputBytes, err := proto.Marshal(&rawInputArgs)

	if err != nil {
		log.Error("Router.OnDisconnected proto.Marshal err", log.ErrorField("err", err))
		return
	}

	//通知各个游戏服玩家下线了
	for key := range mr.mapGameServer {

		nodeId := util.GetGameServiceName(key)

		err = mr.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcOnClose, key, rawInputBytes)

		if err != nil {
			log.SError("Router.OnDisconnected RawGoNode to lobby err:", err.Error())
		}

	}

	//最后需要把这个通道从路由缓存中清除
	delete(mr.mapRouterCache, clientId)

	fmt.Printf("玩家主动下线了...OnDisconnect :%s \n", clientId)
}

func (mr *MsgRouter) OnConnected(clientId string) {

	ipstr := mr.netModule.GetClientIp(clientId)

	fmt.Printf("*******gateservice OnConnected IP:%s  clientId:%s \n    ", ipstr, clientId)

	//需要在这边加入检测是否黑名单IP
}

// RouterMessage 函数返回后，msgBuff内存将被内存池回收
func (mr *MsgRouter) RouterMessage(cliId string, msgType uint16, msgBuff []byte) {

	//fmt.Println("RouterMessage--->msgType: ", msgType)

	//1.登陆消息单独处理
	switch msg.MsgType(msgType) {

	case msg.MsgType_LoginRequest: //登录大厅的消息

		//fmt.Println("***LoginRequest:****** ")

		if mr.login(cliId, msgBuff[2:]) == false {

			log.Errorf("clientId:%s login error ", cliId)

			mr.netModule.Close(cliId)
		}
		return
	case msg.MsgType_Ping: //PING值消息，用于客户端的心跳
		mr.ping(cliId, msgBuff)
		return
	}

	//通过clientId获取nodeId
	nodeInf := mr.GetNodeInfo(cliId)

	//检测是否有效的玩家端
	if nodeInf == nil {
		mr.netModule.Close(cliId)
		return
	}

	if nodeInf.nodeId == "" {
		mr.netModule.Close(cliId)
		return
	}

	//检测登陆认证是否已通过
	if mr.IsLoginedPass(cliId) == false {
		mr.netModule.Close(cliId)
		return
	}

	nodeInf.lastMsgTime = time.Now().Unix()
	nodeInf.msgCount++

	//游戏大厅nodeId
	nodeId := nodeInf.lobbyNodeId

	//检查得到后，不需要每次都查了，提高性能
	if nodeId == "" {
		nodeId = util.GetGameServiceName(util.LobbyService)
		nodeInf.lobbyNodeId = nodeId
	}

	//进入大厅
	if msgType == uint16(msg.MsgType_EnterLobby) {

		if mr.EnterToLobby(cliId, msgBuff[2:]) == false {
			mr.netModule.Close(cliId)
		}
		return
	}

	//根据游戏消息得到要分发的服务器
	gsName := mr.DispatchServieName(msgType)

	//检测此服务是否存在
	if !util.CheckServiceName(gsName) {
		log.SWarn("cannot find gsName:", gsName)
		return
	}

	nodeId = util.GetGameServiceName(gsName)

	//组装原始Rpc参数用于转发
	var inputArgs rpc.RawInputArgs

	rawData := msgBuff //一定要拷贝

	inputArgs.ClientIdList = []string{cliId}
	inputArgs.RawData = rawData[2:]
	inputArgs.MsgType = uint32(msgType)

	inputArgBytes, err := proto.Marshal(&inputArgs)

	if err != nil {
		log.Error("RouterMessage proto.Marshal err", log.ErrorField("err", err))
		return
	}

	//转发消息到对应结点的服务器,架构思路主要通过，大厅服和投注服
	err = mr.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcOnRecv, gsName, inputArgBytes)

	if err != nil {
		//关闭连接
		mr.netModule.Close(cliId)

		log.SError("RawGoNode fail ", err.Error())

		fmt.Printf("gate .... RawGoNode err:%s \n  ", err.Error())
	}

	//fmt.Printf("*****cliId3:%v,msgType:%v, nodeId:%s   gsName:%v    success \n", cliId, msg.MsgType(msgType), nodeId, gsName)

	//后面需要开发把消息发送到日志服存储
}

// DispatchServieName从消息ID得到对应分发的服务名
func (mr *MsgRouter) DispatchServieName(msgType uint16) string {

	var gsName string

	//消息对应的服务，以后要做成一个表
	msgId := msg.MsgType(msgType)

	if (msgId == msg.MsgType_JoinGame) || (msgId == msg.MsgType_ExitGame || msgId == msg.MsgType_EnterGame) {

		gsName = util.LobbyService

	} else if msgId == msg.MsgType_GameTable_Bet || msgId == msg.MsgType_GameTable_CancelBet ||
		msgId == msg.MsgType_GameTable_TipMoney || msgId == msg.MsgType_GameTable_TalkChat {
		//处理钱的去处的服务
		gsName = util.BetService

	} else {
		gsName = util.LobbyService
	}

	return gsName
}

// 是否有效玩家端
func (mr *MsgRouter) IsValidClientId(clientId string) bool {

	if mr.GetNodeInfo(clientId) == nil {
		return false
	}

	return true
}

// 是否经过认证的玩家端
func (mr *MsgRouter) IsLoginedPass(clientId string) bool {

	nodeInf, ok := mr.mapRouterCache[clientId]

	if ok == false {
		return false
	}

	return nodeInf.isLogined
}

// 回复ping值处理
func (mr *MsgRouter) ping(cliId string, data []byte) {

	//先检测是否有效的客户端，经过认证的才是有效的
	if mr.IsValidClientId(cliId) == false {

		fmt.Printf("------------------ping failed ....%s \n", cliId)
		return
	}

	nntime := time.Now().Unix()

	retMsg := msg.MsgPong{}
	retMsg.NowTime = uint32(nntime)

	err := mr.SendMsg(cliId, msg.MsgType_Pong, &retMsg)

	if err != nil {
		log.SError("ping SendMsg error:", err.Error())
		return
	}

	fmt.Printf("---------ping ok ....%s \n", cliId)

	//若检测间隔少于20秒则不发给大厅
	if nntime-mr.mapRouterCache[cliId].sendPingTime < 10 {
		return
	}

	mr.mapRouterCache[cliId].sendPingTime = time.Now().Unix()

	//同时需要发往大厅服务器
	nodeId := util.GetGameServiceName(util.LobbyService)

	//组装原始Rpc参数用于转发
	var inputArgs rpc.RawInputArgs

	rawData := data //一定要拷贝

	inputArgs.ClientIdList = []string{cliId}
	inputArgs.RawData = rawData[2:]
	inputArgs.MsgType = uint32(msg.MsgType_Ping)

	inputArgBytes, err := proto.Marshal(&inputArgs)

	if err != nil {
		log.Error("ping  proto.Marshal err", log.ErrorField("err", err))
		return
	}

	//转发消息到对应结点的服务器,架构思路主要通过，大厅服和投注服
	err = mr.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcOnRecv, util.LobbyService, inputArgBytes)

	if err != nil {
		//关闭连接
		mr.netModule.Close(cliId)

		log.SError("RawGoNode fail ", err.Error())

		fmt.Printf("gate .... RawGoNode err:%s \n  ", err.Error())

		return
	}

	fmt.Println("    发送心跳给大厅服务器")

}

func (mr *MsgRouter) login(cliId string, data []byte) bool {

	var loginReq msg.LoginReq

	err := proto.Unmarshal(data, &loginReq)

	if err != nil {
		mr.loginError(cliId, msg.ErrCode_Fail, "json error !!!")
		return false
	}

	if loginReq.UserId == "" {
		mr.loginError(cliId, msg.ErrCode_Fail, "json error !!!")
		return false
	}

	//是否超出人数检测
	return mr.ToLobbyRpc(cliId, &loginReq)

}

func (mr *MsgRouter) ToLobbyRpc(cliId string, loginReq *msg.LoginReq) bool {

	routerNodeId, _ := mr.GetRouterId(cliId) //获取连接信息

	if routerNodeId != "" {
		mr.loginError(cliId, msg.ErrCode_RepeatLogin, util.GetErrCodeMsg(msg.ErrCode_RepeatLogin))
		return false
	}

	if loginReq == nil {
		mr.loginError(cliId, msg.ErrCode_TokenError, "Router.loginOk msgLoginReq is nil or UserId is empty")
		return false
	}

	//全部转成小写
	loginReq.UserId = strings.ToLower(loginReq.UserId)

	userId, err := util.ValidateUserToken(loginReq.Token)

	if err != nil {
		mr.loginError(cliId, msg.ErrCode_TokenError, "toker error!")
		return false
	}

	if userId != loginReq.UserId {
		mr.loginError(cliId, msg.ErrCode_TokenError, "toker error!")
		return false

	}

	var req rpc.LoginToLobbyReq

	req.UserId = loginReq.UserId //玩家帐号
	req.ClientId = cliId         //系统低层分配的ID
	req.Ip = mr.netModule.GetClientIp(cliId)
	req.GateNodeId = node.GetNodeId() //本机的nodeId

	//存储用户
	mr.AddNodeInfo(cliId, &ClientSessionInfo{nodeId: req.GateNodeId, uid: req.UserId, lastMsgTime: time.Now().Unix(), isLogined: false})

	//更新负载到httpGateService
	mr.updateBalance(1)

	//登录到游戏大厅
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {

		log.SWarn("cannot find servicename nodeId:", util.LobbyService, nodeId)

		nodeId = "game"

		//检测nodeid是否联上本机
		if cluster.GetCluster().IsNodeConnected(nodeId) == false || cluster.GetCluster().IsNodeRetire(nodeId) == true {
			return false
		}

		return false
	}

	rpcLoginFun := util.FormatRpcMethod(util.LobbyService, "RPCLogin")

	fmt.Printf("nodeId:%v,rpcLoginFun:%v\n", nodeId, rpcLoginFun)

	//调用大厅结点通知玩家要登陆大厅
	err = mr.AsyncCallNode(nodeId, rpcLoginFun, &req, func(ret *rpc.LoginToLobbyRet, err error) {

		if err != nil {

			Code := msg.ErrCode_Fail
			Msg := "login server error!!!"

			mr.loginError(cliId, Code, Msg)

			return
		}

		//处理大厅通知不成功的情况
		if ret.Ret > 0 {

			Code := msg.ErrCode_Fail
			Msg := "login server error!!!"

			mr.loginError(cliId, Code, Msg)

			return
		}

		//	fmt.Printf("***************ret UserName:%v  \n", ret.UserName)

		//设置认证通过
		mr.SetNodeLogin(cliId, true, req.Ip)

		//发送给玩家通知登陆成功
		mr.loginSuccess(cliId, req.UserId, ret.UserName)

		//通知大厅发送其它配置消息
		rpcLoginFun = util.FormatRpcMethod(util.LobbyService, "RPCLoginSuccess")

		err = mr.GoNode(nodeId, rpcLoginFun, &req)

		if err != nil {

			Code := msg.ErrCode_Fail
			Msg := "clustor server error!!!"

			mr.loginError(cliId, Code, Msg)

			return
		}
	})

	if err != nil {

		log.SError("AsyncCallNode "+util.LobbyService+" fail :", err.Error(), ",userId:", req.UserId)

		message := "login lobby server error!!!"

		mr.loginError(cliId, msg.ErrCode_Fail, message)

		return false
	}

	return true
}

// 进入大厅
func (mr *MsgRouter) EnterToLobby(cliId string, data []byte) bool {

	var enterLobby msg.EnterToLobby

	err := proto.Unmarshal(data, &enterLobby)

	if err != nil {
		mr.loginError(cliId, msg.ErrCode_Fail, "json error !!!")
		return false
	}

	if enterLobby.UserId == "" {
		mr.loginError(cliId, msg.ErrCode_Fail, "json error !!!")
		return false
	}

	//需要增加检测玩家是否符合通道跟帐号对应
	clientInfo := mr.GetNodeInfo(cliId)
	if clientInfo == nil {
		return false
	}

	//检测玩家帐号是否致
	if clientInfo.uid != enterLobby.UserId {
		return false
	}

	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		log.SWarn("cannot find servicename nodeId:", util.LobbyService, nodeId)
		return false
	}

	var req rpc.LoginToLobbyReq

	req.UserId = clientInfo.uid //玩家帐号
	req.ClientId = cliId        //系统低层分配的ID
	req.Ip = clientInfo.loginIp
	req.GateNodeId = node.GetNodeId() //本机的nodeId

	//通知大厅服务器，可以发送玩家帐号消息了
	rpcLoginFun := util.FormatRpcMethod(util.LobbyService, "RPCLoginSuccess")

	err = mr.GoNode(nodeId, rpcLoginFun, &req)

	if err != nil {

		Code := msg.ErrCode_Fail
		Msg := "clustor server error!!!"

		mr.loginError(cliId, Code, Msg)

		return false
	}

	return true
}

func (mr *MsgRouter) loginError(cliId string, code msg.ErrCode, logMsg string) {

	var retMsg msg.LoginRes

	retMsg.Success = 1
	retMsg.ErrCode = int32(code)
	retMsg.Msg = logMsg

	errSend := mr.SendMsg(cliId, msg.MsgType_LoginResult, &retMsg)

	if errSend != nil {
		log.SError("gate handleLoginError SendMsg err:", errSend.Error())
	}

}

func (mr *MsgRouter) loginSuccess(cliId string, userId string, userName string) {

	//返回给前端统一的结构
	retMsg := msg.LoginRes{}

	retMsg.Success = 0
	retMsg.ErrCode = 0
	retMsg.Msg = userName

	//发送
	errSend := mr.SendMsg(cliId, msg.MsgType_LoginResult, &retMsg)

	if errSend != nil {
		log.SError(" SendMsg err:", errSend.Error())
		return
	}
	//	fmt.Printf("username:%s \n", userName)
}

func (mr *MsgRouter) SendMsg(clientId string, msgType msg.MsgType, msg proto.Message) error {

	byteMsg, err := proto.Marshal(msg)

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

func (mr *MsgRouter) GetRouterId(clientId string) (string, string) {

	noInfo, ok := mr.mapRouterCache[clientId]

	if ok == false {
		return "", ""
	}

	return noInfo.nodeId, noInfo.uid
}

func (mr *MsgRouter) AddNodeInfo(cliId string, info *ClientSessionInfo) {

	mr.mapRouterCache[cliId] = info

}

// GetNodeInfo 获取游戏节点信息
func (mr *MsgRouter) GetNodeInfo(clientId string) *ClientSessionInfo {

	nodeInf, ok := mr.mapRouterCache[clientId]

	if ok == false {
		return nil
	}

	return nodeInf
}

func (mr *MsgRouter) GetClientNum() int {
	return len(mr.mapRouterCache)
}

func (mr *MsgRouter) SetNodeLogin(clientId string, logined bool, ipAddr string) {

	nodeInf, ok := mr.mapRouterCache[clientId]

	if ok == true {
		nodeInf.isLogined = logined
		nodeInf.loginIp = ipAddr
	}
}
func (mr *MsgRouter) AddGameServer(serverId string) {

	mr.mapGameServer[serverId] = serverId
}

// 踢除玩家
func (mr *MsgRouter) KickOutPlayer(clientId string) bool {

	return true
}

// 更新负载到httpGateService
func (mr *MsgRouter) updateBalance(increaseNum int64) {

	var req rpc.GateServiceBalance

	req.NodeId = node.GetNodeId()

	req.IncreaseNum = increaseNum
	rpcFun := util.FormatRpcMethod("HttpGateService", "RPCUpdateBalance")

	err := mr.Go(rpcFun, &req)

	if err != nil {
		log.Errorf("gateService updateBalance GoNode fail:", err.Error())
	}
}

// 通知游戏服下线这个通知
func (mr *MsgRouter) NotifyGameServiceRemoveClid(clientId []string) {

	//2.转发客户端连接断开,
	var rawInputArgs rpc.RawInputArgs

	rawInputArgs.ClientIdList = clientId
	rawInputBytes, err := proto.Marshal(&rawInputArgs)

	if err != nil {
		log.Error("Router.OnDisconnected proto.Marshal err", log.ErrorField("err", err))
		return
	}

	//通知各个游戏服玩家下线了
	for key := range mr.mapGameServer {

		nodeId := util.GetGameServiceName(key)

		err = mr.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcOnClose, key, rawInputBytes)

		if err != nil {
			log.SError("Router.OnDisconnected RawGoNode to lobby err:", err.Error())
		}

	}

}
