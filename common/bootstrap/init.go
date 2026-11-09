package bootstrap

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/oschwald/geoip2-golang"
	"io/ioutil"
	"os"
	"path/filepath"
	"videogame/common/global/my_errors"
	"videogame/common/global/variable"
	models "videogame/common/model"
	mongo_driver "videogame/common/util/mongodb"
	"videogame/common/util/redis_factory"
	"videogame/common/util/yml_config"
)

// 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
func checkRequiredFolders() {
	//1.检查配置文件是否存在
	if _, err := os.Stat(variable.BasePath + "/config/data/config.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigYamlNotExists + err.Error())
	}
	if _, err := os.Stat(variable.BasePath + "/config/data/gorm_v2.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigGormNotExists + err.Error())
	}

	if _, err := os.Stat(variable.BasePath + "/config/data/AreaName/area_name.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigAreaNameNotExists + err.Error())
	}

	if _, err := os.Stat(variable.BasePath + "/config/data/ServiceName/service_name.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigServiceNameNotExists + err.Error())
	}
}

func init() {

	//1.检查配置文件以及日志目录等非编译性的必要条件
	checkRequiredFolders()

	// 2.配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	//初始化数据库mysql
	err := models.InitMysql()

	if err != nil {
		log.SError("gorm_v2.InitMysql:", err.Error()+"\n")
		//直接panic退出程序
		panic(fmt.Sprintf("panic InitMysql err: %v", err))
	}

	//建表
	models.CreateTables(variable.GormDbMysql)

	// 初始化redis
	initRedis()

	// 初始化ip库文件到内存
	initGeoLite2()

	// 初始化下注区域名称配置文件（多语言）
	initAreaNameYml()

	// 初始化Mongo
	initMongo()

	// 初始化游戏服务名
	initServiceName()
}

func initMongo() {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	url := variable.ConfigYml.GetString("Mongo.Url")
	dBName := variable.ConfigYml.GetString("Mongo.DBName")
	mongoClient, err := mongo_driver.NewMongoDBClient(url, dBName)
	if err != nil {
		log.Fatalf("初始化 MongoDB 失败: %v", err)
		panic(fmt.Sprintf("初始化 MongoDB 失败: %v", err))
	}
	variable.MongoClient = mongoClient
}

func initRedis() {
	if variable.RedisStorage == nil {
		host := variable.ConfigYml.GetString("Redis.Host")
		port := variable.ConfigYml.GetString("Redis.Port")
		auth := variable.ConfigYml.GetString("Redis.Auth")
		indexDb := variable.ConfigYml.GetInt("Redis.IndexDb")
		var err error
		err, variable.RedisStorage = redis_factory.InitRedis(host, port, auth, indexDb)

		if err != nil || variable.RedisStorage == nil {
			log.SError("redis_factory.InitRedis:", err.Error()+"\n")
			//直接panic退出程序
			panic(fmt.Sprintf("panic InitRedis err: %v", err))
		}
	}
}

func initGeoLite2() {
	cityPath := filepath.Join(variable.BasePath, "config", "data", "GeoLite2", "GeoLite2-City.mmdb")

	dbBytes, err := ioutil.ReadFile(cityPath)
	if err != nil {
		log.Fatalf("无法读取 IP 库文件: %v", err)
	}

	// 从内存加载ip库
	cityDB, err := geoip2.FromBytes(dbBytes)
	if err != nil {
		log.Fatalf("初始化 IP 库失败: %v", err)
	}

	variable.CityDB = cityDB
}

func initAreaNameYml() {
	variable.ConfigYml = yml_config.CreateYamlFactory("area_name", "/config/data/AreaName")
	variable.ConfigAreaNameYml = variable.ConfigYml.Clone("area_name")
	variable.ConfigAreaNameYml.ConfigFileChangeListen()
}

func initServiceName() {
	variable.ConfigYml = yml_config.CreateYamlFactory("service_name", "/config/data/ServiceName")
	variable.ConfigServiceNameYml = variable.ConfigYml.Clone("service_name")
	variable.ConfigServiceNameYml.ConfigFileChangeListen()
}
