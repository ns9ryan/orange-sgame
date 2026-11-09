package test

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"videogame/common/collect"
	"videogame/common/global/variable"
	"videogame/common/util"
	mongo_driver "videogame/common/util/mongodb"
	"videogame/common/util/yml_config"
)

func MapToTalkChatList(data []interface{}) ([]collect.TalkChat, error) {
	var result []collect.TalkChat

	for _, i := range data {
		item, _ := i.(map[string]interface{})

		// 删除 MongoDB 的 _id 字段（避免不兼容）
		delete(item, "_id")

		// 序列化 map 再反序列化成结构体
		jsonData, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		var chat collect.TalkChat
		err = json.Unmarshal(jsonData, &chat)
		if err != nil {
			return nil, err
		}

		result = append(result, chat)
	}

	return result, nil
}
func TestFindByField(t *testing.T) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient

	chat := &collect.TalkChat{}

	filter := bson.M{"GameId": "FTAN"}
	results, err := chat.Find(filter)
	if err != nil {
		fmt.Println("查询失败:", err)
	} else {
		for _, r := range results {
			fmt.Printf("聊天内容: %+v\n", r)
		}
	}

}

func TestAll(t *testing.T) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient

	talkChat := collect.TalkChat{}
	filter := bson.M{}

	result, err := talkChat.Find(filter)
	if err != nil {
		fmt.Println("获取全部失败")

	}
	fmt.Println("获取全部数据：", result)

}

func TestInsert(t *testing.T) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient

	talkChat := collect.TalkChat{
		GameId:   "FTAN",
		PlayerId: "player001",
		Deskno:   4,
		MsgType:  0,
		Message:  "今天运气真好，小赚一笔！",
		CreateAt: util.GetNowDatetime(),
	}
	res, err := talkChat.InsertOne()
	if err != nil {
		fmt.Println("插入失败")

	}

	fmt.Println("插入成功，ID：", res.InsertedID)

}

func TestDelete(t *testing.T) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient
	talkChat := &collect.TalkChat{}
	filter := bson.M{"GameId": "FTAN"}

	res, err := talkChat.DeleteByField(filter)

	if err != nil {
		fmt.Println("删除失败")

	}

	fmt.Println("删除成功，DeletedCount：", res.DeletedCount)

}

func TestFindWithPagination(t *testing.T) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient

	chat := &collect.TalkChat{}

	filter := bson.D{
		{Key: "GameId", Value: "FTAN"},
		// 可选时间范围条件
		//{Key: "CreateAt", Value: bson.D{
		//	{Key: "$gte", Value: "2025-05-01 00:00:00"},
		//	{Key: "$lte", Value: "2025-05-09 23:59:59"},
		//}},
	}

	page := int64(1)
	pageSize := int64(2)
	//sortField := "CreateAt"
	//sortDesc := true

	sort := bson.D{{"CreateAt", -1}} // 按时间倒序

	list, total, err := chat.FindWithPagination(filter, page, pageSize, sort)
	if err != nil {
		fmt.Println("分页获取失败：", err.Error())
		return
	}

	fmt.Printf("分页获取成功，total=%v ,list=%v \n", total, list)

}

func TestDelAll(t *testing.T) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient

	res, err := variable.MongoClient.DeleteAll("talk_chat_logs")
	if err != nil {
		fmt.Println("删除失败：", err.Error())
		return
	}

	fmt.Printf("DeletedCount=%v \n", res.DeletedCount)

}
