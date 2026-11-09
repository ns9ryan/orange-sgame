package util

import (
	"encoding/json"
	"errors"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/global/variable"
	"videogame/common/msg_struct"
)

const (
	HttpSerivce         = "HttpGateService"
	CoreService         = "CoreService"         //中心服，负责管理全部服务器的结点
	GateService         = "GateService"         //游戏网关服务名
	LobbyService        = "LobbyService"        //大厅的服务名称
	BetService          = "BetService"          //投注服名称
	WalletService       = "WalletService"       //钱包服务名称
	SingleWalletService = "SingleWalletService" //单钱包服务名称

	//BacxService   = "BacxService"
	BacxServiceA = "BacxAService" //游戏服百家乐1
	BacxServiceB = "BacxBService" //游戏服百家乐2
	BacxServiceC = "BacxCService" //游戏服百家乐3

	LongHuService = "LongHuService" //龙虎游戏服务名称
	RotService    = "RotService"
	SicboService  = "SicboService"

	FtanService = "FtanService"
	SdieService = "SdieService"
	FscService  = "FscService"

	MultiBetBacxService = "MultiBetBacxService"
)

type ServiceNameBridge struct {
	GameId      string
	ServiceName string
}

var MapServiceName map[string]*ServiceNameBridge

func InitServiceName() {

	MapServiceName = make(map[string]*ServiceNameBridge)

	// 游戏服务名设置
	var serviceNameList []ServiceNameBridge

	commonServiceNameListStr := variable.ConfigServiceNameYml.GetString("game_services")

	err := json.Unmarshal([]byte(commonServiceNameListStr), &serviceNameList)

	if err != nil {

		err = errors.New("解析service_name.yml文件key=game_services的内容失败！")

		log.Errorf("InitServiceName err:%v", err.Error())

		panic(err)
	}

	for _, sName := range serviceNameList {

		MapServiceName[sName.GameId] = &ServiceNameBridge{GameId: sName.GameId, ServiceName: sName.ServiceName}

	}
}

func GetServerNameFromGameId(gameId string) (string, bool) {

	item, ok := MapServiceName[gameId]

	if !ok {
		return "", false
	}

	return item.ServiceName, true
}

// GetServerNameFromGameId 读取GameServiceId
func GetServerNameFromGameId2(gameId string) (string, bool) {

	var gameName string
	var bFind bool = true

	switch gameId {

	case msg_struct.GAMEID_BACXA:
		gameName = BacxServiceA
	case msg_struct.GAMEID_BACXB:
		gameName = BacxServiceB
	case msg_struct.GAMEID_BACXC:
		gameName = BacxServiceC
	case msg_struct.GAMEID_DTX:
		gameName = LongHuService
	case msg_struct.GAMEID_ROT:
		gameName = RotService
	case msg_struct.GAMEID_SICBO:
		gameName = SicboService
	case msg_struct.GAMEID_FTAN:
		gameName = FtanService
	case msg_struct.GAMEID_SDIE:
		gameName = SdieService
	default:
		bFind = false
	}

	return gameName, bFind
}
