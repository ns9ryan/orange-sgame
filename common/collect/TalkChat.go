package collect

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"videogame/common/global/variable"
)

var TalkChatLogsCollectName = "talk_chat_logs"

type TalkChat struct {
	GameId   string `bson:"GameId"`   //游戏ID
	Deskno   int32  `bson:"Deskno"`   //桌号
	MsgType  int32  `bson:"MsgType"`  //消息类型 0：聊天内容 1：服务器公告  2：派彩公告
	Message  string `bson:"Message"`  //聊天的内容
	PlayerId string `bson:"PlayerId"` //发布聊的人
	CreateAt string `bson:"CreateAt"` //创建时间
	Unused1  int32  `bson:"Unused1"`  //暂时无用
	Unused2  int32  `bson:"Unused2"`  //暂时无用
}

// InsertOne 插入一个文档到指定集合
func (m *TalkChat) InsertOne() (*mongo.InsertOneResult, error) {
	result, err := variable.MongoClient.InsertOne(TalkChatLogsCollectName, m)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteByField 根据指定字段名和值删除匹配的文档
func (m *TalkChat) DeleteByField(filter interface{}) (*mongo.DeleteResult, error) {
	result, err := variable.MongoClient.DeleteOne(TalkChatLogsCollectName, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Find 根据指定字段和值获取所有匹配的文档，并返回结构体切片
func (m *TalkChat) Find(filter interface{}) ([]TalkChat, error) {
	var results []TalkChat
	err := variable.MongoClient.Find(TalkChatLogsCollectName, filter, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindWithPagination 根据指定字段和值获取所有匹配的文档，并返回结构体切片
func (m *TalkChat) FindWithPagination(filter interface{}, pageNum int64, pageSize int64, sort bson.D) ([]TalkChat, int64, error) {
	var results []TalkChat
	total, err := variable.MongoClient.FindWithPagination(TalkChatLogsCollectName, filter, &results, pageNum, pageSize, sort)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetChatListByGameIdAndDeskNo获取指定条数的最近的群聊记录
func (m *TalkChat) GetChatListByGameIdAndDeskNo(gameId string, DeskNo int32, pageNum, pageSize int64) ([]TalkChat, error) {
	var results []TalkChat
	filter := bson.D{
		{Key: "GameId", Value: gameId},
		{Key: "Deskno", Value: DeskNo},
	}
	//按时间 CreateAt 字段降序排序
	sort := bson.D{{"CreateAt", -1}}

	results, _, err := m.FindWithPagination(filter, pageNum, pageSize, sort)
	if err != nil {
		return nil, err
	}

	return results, nil
}
