package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/cluster"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/rpc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	TimeLayout = "2006-01-02 15:04:05"
	TokenKey   = "!gmn$7DYu&h3bDcSfXd)fLQWhvC&dLxR" //32位Key

	TokenValidity = 1000000 * 3600 //10分钟

	IsEmpty = ""

	EVENTForCtrlToGame event.EventType = event.Sys_Event_User_Define + 1 // 自定义事件类型，必需从event.Sys_Event_User_Define开始,event.Sys_Event_User_Define以内给系统预留

	DefaultLocation = "Asia/Shanghai"
)

// 原始Rpc的MethodId定义
const (
	RawRpcMsgDispatch uint32 = 1 //其他服(GameService或其他)->GateService->Client,转发消息
	RawRpcCloseClient uint32 = 2 //其他服(GameService或其他)->GateService->Client,断开与Client连接
	RawRpcOnRecv      uint32 = 3 //Client->GateService->其他服(GameService或其他),转发消息
	RawRpcOnClose     uint32 = 4 //Client->GateService->其他服(GameService或其他),转发Client连接断开事件
)

var Empty struct{}

type WSCfg struct {
	ListenAddr      string
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	LittleEndian    bool //是否小端序
	NodeId          string
	Host            string
	Port            uint
}

type Pem struct {
	KeyFile  string
	CertFile string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewUnionId() string {
	return primitive.NewObjectID().Hex()
}

//// GetMasterCenterNodeId 获取主centerservice的nodeId
//func GetMasterCenterNodeId() string {
//	clientList := make([]*rpc.Client, 0, 16)
//	err, clientList := cluster.GetCluster().GetNodeIdByService(CenterService, clientList, true)
//	if err != nil || len(clientList) == 0 {
//		return ""
//	}
//
//	minId := ""
//	for i := 0; i < len(clientList); i++ {
//		if clientList[i] != nil {
//			if minId == "" || clientList[i].GetTargetNodeId() < minId {
//				minId = clientList[i].GetTargetNodeId()
//			}
//		}
//	}
//
//	return minId
//}

// GetBestNodeId 获取最优服务的nodeId
func GetBestNodeId(serviceMethod string, key uint64) string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetRpcClient("", serviceMethod, true, clientList)
	if err != nil || len(clientList) == 0 {
		return ""
	}

	return clientList[key%uint64(len(clientList))].GetTargetNodeId()
}

func GetBestNodeIdById(serviceMethod string, key string) string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetRpcClient("", serviceMethod, true, clientList)
	if err != nil || len(clientList) == 0 {
		return ""
	}

	return clientList[uint64(HashString2Number(key))%uint64(len(clientList))].GetTargetNodeId()
}

func GetAllNodeId(serviceMethod string, idList []string) []string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetRpcClient("", serviceMethod, true, clientList[:])
	if err != nil || len(clientList) == 0 {
		return idList
	}

	for _, clientItem := range clientList {
		if clientItem == nil {
			break
		}
		idList = append(idList, clientItem.GetTargetNodeId())
	}
	return idList
}

func EncryptToken(uid int32) (string, error) {
	origToken := fmt.Sprintf("%d#%d#%d", rand.Int(), uid, time.Now().Unix()+TokenValidity)
	return SpecialAesEncrypt(origToken, TokenKey)
}

func DecryptToken(token string) (string, int64, error) {

	var uid string
	var tokenValidity int64
	if token == "" {
		return uid, tokenValidity, errors.New("token is empty")
	}

	origToken, err := SpecialAesDecrypt(token, TokenKey)
	if err != nil {
		return uid, tokenValidity, err
	}

	tokenKey := strings.Split(origToken, "#")
	if len(tokenKey) != 3 {
		return uid, tokenValidity, fmt.Errorf("token is error %s", token)
	}
	tokenValidity, err = strconv.ParseInt(tokenKey[2], 10, 64)
	if err != nil {
		return uid, tokenValidity, err
	}
	uid = strings.TrimSpace(tokenKey[1])
	if time.Now().Unix() >= tokenValidity {
		validityTime := time.Unix(tokenValidity, 0)
		log.Error("token timeout", log.String("now", time.Now().Format(TimeLayout)), log.String("validity", validityTime.Format(TimeLayout)))
		return uid, tokenValidity, fmt.Errorf("token expired")
	}
	return uid, tokenValidity, nil
}

func RandomInt(start, end int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(end-start) + start
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano()) // 初始化随机种子
	result := make([]byte, length)
	for i := range result {
		result[i] = Letters[rand.Intn(len(Letters))] // 随机选取一个字符
	}
	return string(result)
}

// 读取这个有这个服务的结点nodeId
func GetGameServiceName(serviceName string) string {
	cliList := make([]*rpc.Client, 0, 16)
	err, cliList := cluster.GetCluster().GetNodeIdByService(serviceName, cliList, true)
	if err != nil || len(cliList) == 0 {
		return ""
	}
	return cliList[0].GetTargetNodeId()
}

func FormatRpcMethod(serviceName string, RpcMethodName string) string {
	return fmt.Sprintf("%s.%s", serviceName, RpcMethodName)
}

// 读取这个服务名方法的结点和名称
func GetServiceRpcMethodName(serviceName string, rpcFuncName string) (string, string) {

	nodeId := GetGameServiceName(serviceName)
	rpcFunc := FormatRpcMethod(serviceName, rpcFuncName)

	return nodeId, rpcFunc
}

func CheckServiceName(serviceName string) bool {

	mapNodeId := cluster.GetNodeByServiceName(serviceName)

	if mapNodeId == nil {
		return false
	}
	return true
}

// GetGlobalGateServiceCfgList 获取gateService所有nodeid的全局配置文件列表
func GetGlobalGateServiceCfgList() ([]*WSCfg, error) {

	wSCfg := []*WSCfg{}

	globalCfg := cluster.GetCluster().GetGlobalCfg()

	if globalCfg == nil {
		return nil, fmt.Errorf("globalCfg is nil")
	}
	mapWSCfg := globalCfg.(map[string]interface{})

	nodeService := mapWSCfg["NodeService"]
	if nodeService == nil {
		return nil, fmt.Errorf("globalCfg NodeService key no exist")
	}
	nodeServiceCfg := nodeService.(map[string]interface{})

	gateService := nodeServiceCfg["GateService"]

	if gateService == nil {
		return nil, fmt.Errorf("globalCfg GateService key no exist")
	}

	gateServiceCfg := gateService.(map[string]interface{})

	cfgList, wsOk := gateServiceCfg["WSCfg"]
	if wsOk == false {
		return nil, fmt.Errorf("globalCfg GateService WSCfg key no exist")
	}

	for _, c := range cfgList.([]interface{}) {

		var cfg WSCfg

		byteWSCfg, err := json.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("GetGlobalGateCfgByNodeId Marshal is error:%s", err.Error())
		}

		err = json.Unmarshal(byteWSCfg, &cfg)
		if err != nil {
			return nil, fmt.Errorf("GetGlobalGateCfgByNodeId Unmarshal is error:%s", err.Error())
		}
		wSCfg = append(wSCfg, &cfg)
	}

	return wSCfg, nil

}

// GetGlobalGateCfgByNodeId 根据nodeId获取全局配置文件中的gateService的配置
func GetGlobalGateCfgByNodeId(nodeId string) (*WSCfg, error) {

	wSCfg := &WSCfg{}

	globalCfg := cluster.GetCluster().GetGlobalCfg()

	if globalCfg == nil {
		return wSCfg, fmt.Errorf("globalCfg is nil")
	}

	mapWSCfg := globalCfg.(map[string]interface{})

	nodeService := mapWSCfg["NodeService"]
	if nodeService == nil {
		return wSCfg, fmt.Errorf("globalCfg NodeService key no exist")
	}
	nodeServiceCfg := nodeService.(map[string]interface{})

	gateService := nodeServiceCfg["GateService"]

	if gateService == nil {
		return wSCfg, fmt.Errorf("globalCfg GateService key no exist")
	}

	gateServiceCfg := gateService.(map[string]interface{})

	cfgList, wsOk := gateServiceCfg["WSCfg"]
	if wsOk == false {
		return nil, fmt.Errorf("globalCfg GateService WSCfg key no exist")
	}

	//这边读取网关的配置情况,有多少网关读取一一对检测是否自已的
	for _, c := range cfgList.([]interface{}) {
		var cfg WSCfg

		byteWSCfg, err := json.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("GetGlobalGateCfgByNodeId Marshal is error:%s", err.Error())
		}
		err = json.Unmarshal(byteWSCfg, &cfg)

		if err != nil {
			return nil, fmt.Errorf("GetGlobalGateCfgByNodeId Unmarshal is error:%s", err.Error())
		}

		//检测是否指定的gate
		if cfg.NodeId == nodeId {
			return &cfg, nil
		}
	}

	return nil, nil

}

// GetGlobalPemFile 获取全局配置文件中的wss证书
func GetGlobalPemFile() (*Pem, error) {

	globalCfg := cluster.GetCluster().GetGlobalCfg()

	if globalCfg == nil {
		return nil, fmt.Errorf("globalCfg is nil")
	}

	mapWSCfg := globalCfg.(map[string]interface{})

	p := mapWSCfg["Pem"]
	if p == nil {
		return nil, fmt.Errorf("globalCfg Pem key no exist")
	}

	var pem Pem
	bytePem, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("GetGlobalPemFile Marshal is error:%s", err.Error())
	}

	err = json.Unmarshal(bytePem, &pem)
	if err != nil {
		return nil, fmt.Errorf("GetGlobalPemFile Unmarshal is error:%s", err.Error())
	}

	return &pem, nil
}

// GetGlobalDebug 获取全局配置文件中的debug
func GetGlobalDebug() (bool, error) {

	globalCfg := cluster.GetCluster().GetGlobalCfg()

	if globalCfg == nil {
		return true, fmt.Errorf("globalCfg is nil")
	}

	mapWSCfg := globalCfg.(map[string]interface{})

	debug := mapWSCfg["Debug"]

	if debug == nil {
		return true, fmt.Errorf("globalCfg Debug key no exist")
	}

	d := debug.(bool)

	return d, nil
}
