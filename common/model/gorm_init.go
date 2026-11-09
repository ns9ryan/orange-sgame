package models

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"reflect"
	"strings"
	"time"
	"videogame/common/global/variable"
	"videogame/common/util/my_errors"
)

type ConfigParams struct {
	Write ConfigParamsDetail
	Read  ConfigParamsDetail
}
type ConfigParamsDetail struct {
	Host     string
	DataBase string
	Port     int
	Prefix   string
	User     string
	Pass     string
	Charset  string
}

// 初始化
func InitMysql() error {
	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	confPrams, err := GetMysqlConfig()

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get MySQL config: %v\n", err))
	}

	if gormDbMysql, err := GetSqlDriver("mysql", variable.ConfigGormv2Yml.GetInt("Gormv2.Mysql.IsOpenReadDb"), confPrams); err == nil {
		variable.GormDbMysql = gormDbMysql
		return nil
	}

	return err
}

// GetMysqlConfig 从配置中解析 MySQL 的 Write 和 Read 配置
func GetMysqlConfig() (ConfigParams, error) {
	confPrams := ConfigParams{
		Write: struct {
			Host     string
			DataBase string
			Port     int
			Prefix   string
			User     string
			Pass     string
			Charset  string
		}{
			Host:     variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Write.Host"),
			DataBase: variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Write.DataBase"),
			Port:     variable.ConfigGormv2Yml.GetInt("Gormv2.Mysql.Write.Port"),
			Prefix:   variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Write.Prefix"),
			User:     variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Write.User"),
			Pass:     variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Write.Pass"),
			Charset:  variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Write.Charset"),
		},

		Read: struct {
			Host     string
			DataBase string
			Port     int
			Prefix   string
			User     string
			Pass     string
			Charset  string
		}{
			Host:     variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Read.Host"),
			DataBase: variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Read.DataBase"),
			Port:     variable.ConfigGormv2Yml.GetInt("Gormv2.Mysql.Read.Port"),
			Prefix:   variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Read.Prefix"),
			User:     variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Read.User"),
			Pass:     variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Read.Pass"),
			Charset:  variable.ConfigGormv2Yml.GetString("Gormv2.Mysql.Read.Charset"),
		},
	}
	return confPrams, nil
}

// 获取数据库驱动, 可以通过options 动态参数连接任意多个数据库
func GetSqlDriver(sqlType string, readDbIsOpen int, dbConf ...ConfigParams) (*gorm.DB, error) {
	var dbDialector gorm.Dialector
	if val, err := getDbDialector(sqlType, "Write", dbConf...); err != nil {
		log.SError("getDbDialector err:", my_errors.ErrorsDialectorDbInitFail+sqlType+err.Error()+"\n")
	} else {
		dbDialector = val
	}

	gormDb, err := gorm.Open(dbDialector, &gorm.Config{})
	if err != nil {
		log.SError("gorm 数据库驱动初始化失败 err:", err.Error()+"\n")
		fmt.Printf("gorm 数据库驱动初始化失败:%v\n", err.Error())
		return nil, err
	}

	fmt.Printf("gorm 数据库驱动初始化成功\n")
	// 如果开启了读写分离，配置读数据库（resource、read、replicas）
	if readDbIsOpen == 1 {
		if val, err := getDbDialector(sqlType, "Read", dbConf...); err != nil {
			log.SError("gorm 数据库驱动初始化失败 err:", my_errors.ErrorsDialectorDbInitFail+sqlType+err.Error()+"\n")
			fmt.Printf("gorm 数据库驱动初始化失败:%v\n", err.Error())
		} else {
			dbDialector = val
		}
		resolverConf := dbresolver.Config{
			Replicas: []gorm.Dialector{dbDialector}, //  读 操作库，查询类
			Policy:   dbresolver.RandomPolicy{},     // sources/replicas 负载均衡策略适用于
		}
		err = gormDb.Use(dbresolver.Register(resolverConf).SetConnMaxIdleTime(time.Second * 30).
			SetConnMaxLifetime(variable.ConfigGormv2Yml.GetDuration("Gormv2."+sqlType+".Read.SetConnMaxLifetime") * time.Second).
			SetMaxIdleConns(variable.ConfigGormv2Yml.GetInt("Gormv2." + sqlType + ".Read.SetMaxIdleConns")).
			SetMaxOpenConns(variable.ConfigGormv2Yml.GetInt("Gormv2." + sqlType + ".Read.SetMaxOpenConns")))
		if err != nil {
			log.SError("gorm 数据库驱动初始化失败 err:", err.Error()+"\n")
			fmt.Printf("gorm 数据库驱动初始化失败:%v\n", err.Error())
			return nil, err
		}
	}

	_ = gormDb.Callback().Create().Before("gorm:before_create").Register("CreateBeforeHook", CreateBeforeHook)
	// 为了完美支持gorm的一系列回调函数
	_ = gormDb.Callback().Update().Before("gorm:before_update").Register("UpdateBeforeHook", UpdateBeforeHook)

	// 为主连接设置连接池
	if rawDb, err := gormDb.DB(); err != nil {
		return nil, err
	} else {
		rawDb.SetConnMaxIdleTime(time.Second * 30)
		rawDb.SetConnMaxLifetime(variable.ConfigGormv2Yml.GetDuration("Gormv2."+sqlType+".Write.SetConnMaxLifetime") * time.Second)
		rawDb.SetMaxIdleConns(variable.ConfigGormv2Yml.GetInt("Gormv2." + sqlType + ".Write.SetMaxIdleConns"))
		rawDb.SetMaxOpenConns(variable.ConfigGormv2Yml.GetInt("Gormv2." + sqlType + ".Write.SetMaxOpenConns"))

		//AutoMigrate自动创建不存在的表：
		err := CreateTables(gormDb)

		if err != nil {
			return nil, err
		}

		// 某些表自动插入一些初始数据：
		// err := insertTables(gormDb)
		// if err !- nil {
		//     return nil, err
		// }

		// 全局sql的debug配置
		if variable.ConfigGormv2Yml.GetBool("Gormv2.SqlDebug") {
			return gormDb.Debug(), nil
		} else {
			return gormDb, nil
		}
	}
}

// 获取一个数据库方言(Dialector),通俗的说就是根据不同的连接参数，获取具体的一类数据库的连接指针
func getDbDialector(sqlType, readWrite string, dbConf ...ConfigParams) (gorm.Dialector, error) {
	var dbDialector gorm.Dialector
	dsn := getDsn(sqlType, readWrite, dbConf...)
	dbDialector = mysql.Open(dsn)
	return dbDialector, nil
}

// 根据配置参数生成数据库驱动 dsn
func getDsn(sqlType, readWrite string, dbConf ...ConfigParams) string {
	Host := variable.ConfigGormv2Yml.GetString("Gormv2." + sqlType + "." + readWrite + ".Host")
	DataBase := variable.ConfigGormv2Yml.GetString("Gormv2." + sqlType + "." + readWrite + ".DataBase")
	Port := variable.ConfigGormv2Yml.GetInt("Gormv2." + sqlType + "." + readWrite + ".Port")
	User := variable.ConfigGormv2Yml.GetString("Gormv2." + sqlType + "." + readWrite + ".User")
	Pass := variable.ConfigGormv2Yml.GetString("Gormv2." + sqlType + "." + readWrite + ".Pass")
	Charset := variable.ConfigGormv2Yml.GetString("Gormv2." + sqlType + "." + readWrite + ".Charset")

	if len(dbConf) > 0 {
		if strings.ToLower(readWrite) == "write" {
			if len(dbConf[0].Write.Host) > 0 {
				Host = dbConf[0].Write.Host
			}
			if len(dbConf[0].Write.DataBase) > 0 {
				DataBase = dbConf[0].Write.DataBase
			}
			if dbConf[0].Write.Port > 0 {
				Port = dbConf[0].Write.Port
			}
			if len(dbConf[0].Write.User) > 0 {
				User = dbConf[0].Write.User
			}
			if len(dbConf[0].Write.Pass) > 0 {
				Pass = dbConf[0].Write.Pass
			}
			if len(dbConf[0].Write.Charset) > 0 {
				Charset = dbConf[0].Write.Charset
			}
		} else {
			if len(dbConf[0].Read.Host) > 0 {
				Host = dbConf[0].Read.Host
			}
			if len(dbConf[0].Read.DataBase) > 0 {
				DataBase = dbConf[0].Read.DataBase
			}
			if dbConf[0].Read.Port > 0 {
				Port = dbConf[0].Read.Port
			}
			if len(dbConf[0].Read.User) > 0 {
				User = dbConf[0].Read.User
			}
			if len(dbConf[0].Read.Pass) > 0 {
				Pass = dbConf[0].Read.Pass
			}
			if len(dbConf[0].Read.Charset) > 0 {
				Charset = dbConf[0].Read.Charset
			}
		}
	}

	switch strings.ToLower(sqlType) {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local", User, Pass, Host, Port, DataBase, Charset)
	case "sqlserver", "mssql":
		return fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=disable", Host, Port, DataBase, User, Pass)
	case "postgresql", "postgre", "postgres":
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable TimeZone=Asia/Shanghai", Host, Port, DataBase, User, Pass)
	}
	return ""
}

func AutoMigrateIfNotExists(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			if err := db.Clauses(dbresolver.Write).AutoMigrate(model); err != nil {
				return fmt.Errorf("自动迁移失败（%T）：%w", model, err)
			}
		}
	}
	return nil
}

func CreateTables(gormDb *gorm.DB) error {

	log.Info("begin to init database......")

	// 自动创建目前还不存在的数据库表
	if err := gormDb.Clauses(dbresolver.Write).AutoMigrate(
		&SGAnswerBacx{},
		&SGAnswerDTX{},
		&SGAnswerSicbo{},
		&SGAnswerSdie{},
		&SGAnswerFsc{},
		&SBetRecord{},
		&SBetAccountRecord{},
		&STipRecord{},
		&SUserAccount{},
		&SMerchantAccount{},
		&SWalletTransPending{},
		&SWalletTransHistory{},
	); err != nil {
		log.SError("createTables失败 err:", err.Error()+"\n")
		return err
	}

	return nil
}

// 以下函数都是gorm的hook函数，拦截一些官方我们认为不合格的操作行为，提升项目整体的完美性
// MaskNotDataError 解决gorm v2 包在查询无数据时，报错问题（record not found），但是官方认为报错是应该是，我们认为查询无数据，代码一切ok，不应该报错
func MaskNotDataError(gormDB *gorm.DB) {
	gormDB.Statement.RaiseErrorOnNotFound = false
}

// InterceptCreatePramsNotPtrError 拦截 create 函数参数如果是非指针类型的错误,新用户最容犯此错误

func CreateBeforeHook(gormDB *gorm.DB) {
	if reflect.TypeOf(gormDB.Statement.Dest).Kind() != reflect.Ptr {
		variable.ZapLog.Warn(my_errors.ErrorsGormDBCreateParamsNotPtr)
	} else {
		destValueOf := reflect.ValueOf(gormDB.Statement.Dest).Elem()
		if destValueOf.Type().Kind() == reflect.Slice || destValueOf.Type().Kind() == reflect.Array {
			inLen := destValueOf.Len()
			for i := 0; i < inLen; i++ {
				row := destValueOf.Index(i)
				if row.Type().Kind() == reflect.Struct {
					if b, column := structHasSpecialField("CreatedAt", row); b {
						destValueOf.Index(i).FieldByName(column).Set(reflect.ValueOf(time.Now().Format(variable.DateFormat)))
					}
					if b, column := structHasSpecialField("UpdatedAt", row); b {
						destValueOf.Index(i).FieldByName(column).Set(reflect.ValueOf(time.Now().Format(variable.DateFormat)))
					}

				} else if row.Type().Kind() == reflect.Map {
					if b, column := structHasSpecialField("created_at", row); b {
						row.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(time.Now().Format(variable.DateFormat)))
					}
					if b, column := structHasSpecialField("updated_at", row); b {
						row.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(time.Now().Format(variable.DateFormat)))
					}
				}
			}
		} else if destValueOf.Type().Kind() == reflect.Struct {
			//  if destValueOf.Type().Kind() == reflect.Struct
			// 参数校验无错误自动设置 CreatedAt、 UpdatedAt
			if b, column := structHasSpecialField("CreatedAt", gormDB.Statement.Dest); b {
				gormDB.Statement.SetColumn(column, time.Now().Format(variable.DateFormat))
			}
			if b, column := structHasSpecialField("UpdatedAt", gormDB.Statement.Dest); b {
				gormDB.Statement.SetColumn(column, time.Now().Format(variable.DateFormat))
			}
		} else if destValueOf.Type().Kind() == reflect.Map {
			if b, column := structHasSpecialField("created_at", gormDB.Statement.Dest); b {
				destValueOf.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(time.Now().Format(variable.DateFormat)))
			}
			if b, column := structHasSpecialField("updated_at", gormDB.Statement.Dest); b {
				destValueOf.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(time.Now().Format(variable.DateFormat)))
			}
		}
	}
}

// UpdateBeforeHook
// InterceptUpdatePramsNotPtrError 拦截 save、update 函数参数如果是非指针类型的错误
// 对于开发者来说，以结构体形式更新数，只需要在 update 、save 函数的参数前面添加 & 即可
// 最终就可以完美兼支持、兼容 gorm 的所有回调函数
// 但是如果是指定字段更新，例如： UpdateColumn 函数则只传递值即可，不需要做校验
func UpdateBeforeHook(gormDB *gorm.DB) {
	if reflect.TypeOf(gormDB.Statement.Dest).Kind() == reflect.Struct {
		//_ = gormDB.AddError(errors.New(my_errors.ErrorsGormDBUpdateParamsNotPtr))
		variable.ZapLog.Warn(my_errors.ErrorsGormDBUpdateParamsNotPtr)
	} else if reflect.TypeOf(gormDB.Statement.Dest).Kind() == reflect.Map {
		// 如果是调用了 gorm.Update 、updates 函数 , 在参数没有传递指针的情况下，无法触发回调函数

	} else if reflect.TypeOf(gormDB.Statement.Dest).Kind() == reflect.Ptr && reflect.ValueOf(gormDB.Statement.Dest).Elem().Kind() == reflect.Struct {
		// 参数校验无错误自动设置 UpdatedAt
		if b, column := structHasSpecialField("UpdatedAt", gormDB.Statement.Dest); b {
			gormDB.Statement.SetColumn(column, time.Now().Format(variable.DateFormat))
		}
	} else if reflect.TypeOf(gormDB.Statement.Dest).Kind() == reflect.Ptr && reflect.ValueOf(gormDB.Statement.Dest).Elem().Kind() == reflect.Map {
		if b, column := structHasSpecialField("updated_at", gormDB.Statement.Dest); b {
			destValueOf := reflect.ValueOf(gormDB.Statement.Dest).Elem()
			destValueOf.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(time.Now().Format(variable.DateFormat)))
		}
	}
}

// structHasSpecialField  检查结构体是否有特定字段
func structHasSpecialField(fieldName string, anyStructPtr interface{}) (bool, string) {
	var tmp reflect.Type
	if reflect.TypeOf(anyStructPtr).Kind() == reflect.Ptr && reflect.ValueOf(anyStructPtr).Elem().Kind() == reflect.Map {
		destValueOf := reflect.ValueOf(anyStructPtr).Elem()
		for _, item := range destValueOf.MapKeys() {
			if item.String() == fieldName {
				return true, fieldName
			}
		}
	} else if reflect.TypeOf(anyStructPtr).Kind() == reflect.Ptr && reflect.ValueOf(anyStructPtr).Elem().Kind() == reflect.Struct {
		destValueOf := reflect.ValueOf(anyStructPtr).Elem()
		tf := destValueOf.Type()
		for i := 0; i < tf.NumField(); i++ {
			if !tf.Field(i).Anonymous && tf.Field(i).Type.Kind() != reflect.Struct {
				if tf.Field(i).Name == fieldName {
					return true, getColumnNameFromGormTag(fieldName, tf.Field(i).Tag.Get("gorm"))
				}
			} else if tf.Field(i).Type.Kind() == reflect.Struct {
				tmp = tf.Field(i).Type
				for j := 0; j < tmp.NumField(); j++ {
					if tmp.Field(j).Name == fieldName {
						return true, getColumnNameFromGormTag(fieldName, tmp.Field(j).Tag.Get("gorm"))
					}
				}
			}
		}
	} else if reflect.Indirect(anyStructPtr.(reflect.Value)).Type().Kind() == reflect.Struct {
		// 处理结构体
		destValueOf := anyStructPtr.(reflect.Value)
		tf := destValueOf.Type()
		for i := 0; i < tf.NumField(); i++ {
			if !tf.Field(i).Anonymous && tf.Field(i).Type.Kind() != reflect.Struct {
				if tf.Field(i).Name == fieldName {
					return true, getColumnNameFromGormTag(fieldName, tf.Field(i).Tag.Get("gorm"))
				}
			} else if tf.Field(i).Type.Kind() == reflect.Struct {
				tmp = tf.Field(i).Type
				for j := 0; j < tmp.NumField(); j++ {
					if tmp.Field(j).Name == fieldName {
						return true, getColumnNameFromGormTag(fieldName, tmp.Field(j).Tag.Get("gorm"))
					}
				}
			}
		}
	} else if reflect.Indirect(anyStructPtr.(reflect.Value)).Type().Kind() == reflect.Map {
		destValueOf := anyStructPtr.(reflect.Value)
		for _, item := range destValueOf.MapKeys() {
			if item.String() == fieldName {
				return true, fieldName
			}
		}
	}
	return false, ""
}

// getColumnNameFromGormTag 从 gorm 标签中获取字段名
// @defaultColumn 如果没有 gorm：column 标签为字段重命名，则使用默认字段名
// @TagValue 字段中含有的gorm："column:created_at" 标签值，可能的格式：1. column:created_at    、2. default:null;  column:created_at  、3.  column:created_at; default:null
func getColumnNameFromGormTag(defaultColumn, TagValue string) (str string) {
	pos1 := strings.Index(TagValue, "column:")
	if pos1 == -1 {
		str = defaultColumn
		return
	} else {
		TagValue = TagValue[pos1+7:]
	}
	pos2 := strings.Index(TagValue, ";")
	if pos2 == -1 {
		str = TagValue
	} else {
		str = TagValue[:pos2]
	}
	return strings.ReplaceAll(str, " ", "")
}
