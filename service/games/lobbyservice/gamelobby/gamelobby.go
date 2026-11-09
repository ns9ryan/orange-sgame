package gamelobby

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/service"
	"time"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/util"
	"videogame/service/games/player"
)

// 游戏房间名称
type RoomInfo struct {
	RoomNo    int    //处在房间号
	Title     string //游戏桌标题
	GroupName string //分组的名称
	VideoURL  string //视频源的地址

	ServiceName string //服务名
}

// 游戏组结构体
type GameGroup struct {
	GroupNo   int    //组号
	GroupName string //组名称
	SubCode   string //二级分类明称

	Comment string

	IsHide uint //是否需要隐藏

	GameNum int //游戏个数
}

// 每张桌子的信息
type TableInfo struct {
	IsEnable bool //是否工作

	GameId     string //游戏名称,唯一
	DeskNo     int    //游戏桌号
	Set        int    //游戏的场次
	Round      int    //游戏的局数
	GameStatus int    //游戏状态

	TotalBetTime int       //总的投注时间
	StartBetTime time.Time //开始下注的时间
	IsAnswered   bool      //是否已开奖

	Dealer msg_struct.DealerInfo //本桌的荷官

	RoomInfo RoomInfo //本桌在的房间信息

	PlayerNum int //玩家人数

	AnswerLast msg.STOC_GameAnswer //当前开答

	HistoryAnswerPerio *msg.NotifyLobbyHistoryAnswers //游戏历史开答
}

type GameLobby struct {
	service.Module

	//通知的广播消息
	NoticMessage string

	//游戏桌子
	MapGameTable map[string]*TableInfo //map[gameId]*TableInfo

	//游戏组对应的桌号ID
	MapGameGroup map[string][]string //map[ groupName ] []gameId

	//游戏分组
	GameGroup []GameGroup
}

func NewGameLobby() *GameLobby {
	return &GameLobby{}
}

func (lobby *GameLobby) OnInit() error {

	lobby.MapGameTable = make(map[string]*TableInfo)
	lobby.MapGameGroup = make(map[string][]string)

	lobby.InitGameGroup()
	lobby.InitGameRoom()
	lobby.InitNoticeMessage()

	return nil

}

func (gm *GameLobby) OnDestroy() error {
	return nil
}

func (gm *GameLobby) OnStart() error {

	return nil
}

// 初始化游戏分组
func (gm *GameLobby) InitGameGroup() bool {

	gm.GameGroup = make([]GameGroup, 5)

	var groupTable models.SGameGroup

	//先从数据库读取分组信息
	groups, err := groupTable.GetGameGroups()

	if err != nil {
		return false
	}

	size := len(groups)

	//若数据库读取条数=0，则用初始化值
	if size == 0 {
		gm.GameGroup = make([]GameGroup, 6)

		//需要从数据库读取
		gm.GameGroup[0].GroupNo = 0
		gm.GameGroup[0].GroupName = "BACX"
		gm.GameGroup[1].GroupNo = 1
		gm.GameGroup[1].GroupName = "DTX"
		gm.GameGroup[2].GroupNo = 2
		gm.GameGroup[2].GroupName = "ROT"
		gm.GameGroup[3].GroupNo = 3
		gm.GameGroup[3].GroupName = "SICBO"
		gm.GameGroup[4].GroupNo = 4
		gm.GameGroup[4].GroupName = "FTAN"
		gm.GameGroup[5].GroupNo = 5
		gm.GameGroup[5].GroupName = "SDIE"

	} else {

		count := 0

		gm.GameGroup = make([]GameGroup, 0)

		//先统计得到哪此组是启用的
		for _, group := range groups {

			if group.Status != 1 { //若是0不需要读取
				continue
			}

			//此游戏组没有启用
			if group.Status == 0 {
				continue
			}

			groupInfo := GameGroup{}

			groupInfo.GroupNo = count
			groupInfo.GroupName = group.GameCategoryCode
			groupInfo.Comment = group.Description
			groupInfo.IsHide = group.IsHide
			groupInfo.SubCode = group.SubCode

			gm.GameGroup = append(gm.GameGroup, groupInfo)
			gm.MapGameGroup[group.GameCategoryCode] = make([]string, 0)

			count++
		}

	}

	return true
}

// 初始化游戏桌
func (gm *GameLobby) InitGameRoom() bool {

	var gameRooms models.SGameTable

	//读取数据库桌的信息
	rooms, err := gameRooms.GetGameTables()

	if err != nil {
		return false
	}

	//读取房间名
	for ind, room := range rooms {

		//先检测此房间否启用
		if room.Status == 0 { //只有启用的才传
			continue
		}

		code := room.GameCategoryCode

		//查找对应在的分组
		for _, group := range gm.GameGroup {

			if code == group.GroupName {

				gameId := room.GameTableCode

				gm.MapGameTable[gameId] = &TableInfo{GameId: gameId}

				gm.MapGameTable[gameId].DeskNo = int(room.DeskNo)
				gm.MapGameTable[gameId].RoomInfo.Title = room.Description
				gm.MapGameTable[gameId].RoomInfo.VideoURL = room.StreamUrl
				gm.MapGameTable[gameId].RoomInfo.RoomNo = ind
				gm.MapGameTable[gameId].RoomInfo.GroupName = code
				gm.MapGameTable[gameId].RoomInfo.ServiceName = room.GameServiceName
				gm.MapGameTable[gameId].TotalBetTime = int(room.BetInterval) //投注时间

				gm.MapGameGroup[code] = append(gm.MapGameGroup[code], gameId)

				group.GameNum++
				break
			}
		}

	}

	return true
}

// 初始化通知消息
func (gm *GameLobby) InitNoticeMessage() bool {

	gm.NoticMessage = "这是测试用的通知消息"

	return true
}

// 读取指定游戏桌的信息
func (gm *GameLobby) GetTableInfo(gameId string) (*TableInfo, bool) {

	if gm.MapGameTable[gameId] == nil {
		return nil, false
	}

	return gm.MapGameTable[gameId], true
}

func (gm *GameLobby) UpdateTableInfo(gameId string, videoUrl string, gameTitle string) bool {

	if gm.MapGameTable[gameId] == nil {
		return false
	}

	if gm.MapGameTable[gameId].RoomInfo.VideoURL == videoUrl {
		gm.MapGameTable[gameId].RoomInfo.VideoURL = videoUrl
	}

	if gameTitle != "" {
		gm.MapGameTable[gameId].RoomInfo.Title = gameTitle
	}

	return true

}

// 更新各游戏桌的游戏状态,有增加人数字段
func (gm *GameLobby) UpdateGameStatus(msg *msg.NotifyLobbyGameStatus) {

	//fmt.Printf("GameLobby UpdateGameStatus :%v \n ", *msg)

	if gm.MapGameTable[msg.GameId] == nil {
		gm.MapGameTable[msg.GameId] = new(TableInfo)
		gm.MapGameTable[msg.GameId].GameId = msg.GameId
		gm.MapGameTable[msg.GameId].DeskNo = int(msg.DeskNo)
	}

	gm.MapGameTable[msg.GameId].GameStatus = int(msg.GameStatus)
	gm.MapGameTable[msg.GameId].Set = int(msg.Set)
	gm.MapGameTable[msg.GameId].Round = int(msg.Round)
	gm.MapGameTable[msg.GameId].IsAnswered = util.ConvertIntToBool(int(msg.IsAnswered))

	if msg.GameStatus == msg_struct.GameStatusBet {
		gm.MapGameTable[msg.GameId].StartBetTime = time.Now()
	}

	gm.MapGameTable[msg.GameId].PlayerNum = int(msg.PlayerNum)
}

// 更新游戏桌的开答表
func (gm *GameLobby) UpdateGameAnswers(ans *msg.NotifyLobbyHistoryAnswers) {

	if ans.GameId == "" {
		return
	}

	if gm.MapGameTable[ans.GameId] == nil {

		gm.MapGameTable[ans.GameId] = new(TableInfo)
		gm.MapGameTable[ans.GameId].GameId = ans.GameId
		gm.MapGameTable[ans.GameId].DeskNo = int(ans.DeskNo)

		gm.MapGameTable[ans.GameId].GameStatus = int(ans.GameStatus)
		gm.MapGameTable[ans.GameId].Set = int(ans.Set)
		gm.MapGameTable[ans.GameId].Round = int(ans.Round)
	}

	gm.MapGameTable[ans.GameId].HistoryAnswerPerio = ans
	gm.MapGameTable[ans.GameId].IsEnable = true

}

// 更新玩家人数
func (gm *GameLobby) UpdateGamePlayerNum(gameId string, playerNum int) int {

	if gm.MapGameTable[gameId] == nil {
		return 0
	}

	gm.MapGameTable[gameId].PlayerNum = playerNum

	if gm.MapGameTable[gameId].PlayerNum < 0 {
		gm.MapGameTable[gameId].PlayerNum = 0
	}

	fmt.Printf("游戏%s 在线人数变成: %d\n", gameId, playerNum)

	return gm.MapGameTable[gameId].PlayerNum
}

// 增加玩家人数
func (gm *GameLobby) GamePlayerNumAdd(gameId string, playerNum int) int {

	if gm.MapGameTable[gameId] == nil {
		return 0
	}

	gm.MapGameTable[gameId].PlayerNum += playerNum

	if gm.MapGameTable[gameId].PlayerNum < 0 {
		gm.MapGameTable[gameId].PlayerNum = 0
	}

	fmt.Printf("游戏%s 在线人数变成: %d\n", gameId, playerNum)

	return gm.MapGameTable[gameId].PlayerNum
}

func (gm *GameLobby) UpdateDealerInfo(gameId string, info *msg_struct.DealerInfo) {

	if gm.MapGameTable[gameId] == nil {
		return
	}

	gm.MapGameTable[gameId].Dealer = *info

	fmt.Printf("更新荷管信息：%v", *info)

}

// 没用也=
func (gm *GameLobby) OpenAnswer(ansmsg *msg.STOC_GameAnswer) {

	if gm.MapGameTable[ansmsg.GameId] == nil {
		return
	}

	gm.MapGameTable[ansmsg.GameId].AnswerLast.GameId = ansmsg.GameId
	gm.MapGameTable[ansmsg.GameId].AnswerLast.Set = ansmsg.Set
	gm.MapGameTable[ansmsg.GameId].AnswerLast.Round = ansmsg.Round
	gm.MapGameTable[ansmsg.GameId].AnswerLast.DataType = ansmsg.DataType
	gm.MapGameTable[ansmsg.GameId].AnswerLast.Answer = ansmsg.Answer
}

// 发送各游戏的开答路子给玩家
func (gm *GameLobby) SendHistoryAnswerToPlayer(player *player.Player) {

	for _, table := range gm.MapGameTable {

		if table == nil {
			continue
		}

		if table.IsEnable == false {
			continue
		}
		if table.HistoryAnswerPerio == nil {
			continue
		}

		player.SendMsg(msg.MsgType_Lobby_GameHistoryAnswers, table.HistoryAnswerPerio)
	}
}

// 发送每张桌的在线人数玩家
func (gm *GameLobby) SendGamePlayerNumToPlayer(player *player.Player) {

	for _, table := range gm.MapGameTable {

		if table == nil {
			continue
		}

		var playerNum msg.STOC_PlayerNum

		playerNum.DeskNo = int32(table.DeskNo)
		playerNum.GameId = table.GameId
		playerNum.OnlineNum = int32(table.PlayerNum)

		player.SendMsg(msg.MsgType_GameInfo_PlayNum, &playerNum)
	}

}

// 发送指定桌的在线人数玩家
func (gm *GameLobby) SendOneGamePlayerNumToPlayer(gameId string, player *player.Player) {

	table, ok := gm.MapGameTable[gameId]

	if ok == false {
		return
	}

	if table == nil {
		return
	}

	var playerNum msg.STOC_PlayerNum

	playerNum.DeskNo = int32(table.DeskNo)
	playerNum.GameId = table.GameId
	playerNum.OnlineNum = int32(table.PlayerNum)

	player.SendMsg(msg.MsgType_GameInfo_PlayNum, &playerNum)

}
