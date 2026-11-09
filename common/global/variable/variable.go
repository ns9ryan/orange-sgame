package variable

import (
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"strings"
	"videogame/common/global/my_errors"
	mongo_driver "videogame/common/util/mongodb"
	redis_instance "videogame/common/util/redis_factory"
	"videogame/common/util/yml_config/ymlconfig_interf"
)

var (
	BasePath           string
	EventDestroyPrefix = "Destroy_" //  程序退出时需要销毁的事件前缀
	ConfigKeyPrefix_   = "Config_"  //  配置文件键值缓存时，键的前缀
	GormDbMysql        *gorm.DB     // 全局gorm的客户端连接

	RedisStorage *redis_instance.RedisStorage // 全局redis的客户端连接

	MongoClient *mongo_driver.MongoDBClient

	ConfigYml       ymlconfig_interface.YmlConfigInterf // 全局配置文件指针
	ConfigGormv2Yml ymlconfig_interface.YmlConfigInterf // 全局配置文件指针

	ZapLog     *zap.Logger
	DateFormat = "2006-01-02 15:04:05" //  设置全局日期时间格式

	CityDB              *geoip2.Reader
	InvalidCountryCodes = []string{"CN"} //过滤该地区IP登录的用户！

	ConfigAreaNameYml    ymlconfig_interface.YmlConfigInterf // 投注区域名称（多语言）
	ConfigServiceNameYml ymlconfig_interface.YmlConfigInterf // 服务名称配置文件
)

// ReadFromWrite 强制从主库（WriteDB）读取数据【write库 就是 主库】
func ReadFromWrite() *gorm.DB {
	return GormDbMysql.Clauses(dbresolver.Write).Session(&gorm.Session{})
}

func init() {

	// 1.初始化程序根目录
	if curPath, err := os.Getwd(); err == nil {
		// 路径进行处理，兼容单元测试程序程序启动时的奇怪路径
		if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-test") {
			BasePath = strings.Replace(strings.Replace(curPath, `\test`, "", 1), `/test`, "", 1)
		} else {
			BasePath = curPath
		}
	} else {
		log.Fatal(my_errors.ErrorsBasePath)
	}
}
