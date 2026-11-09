package models

/******************************************
  百家乐的开答表

******************************************/
import (
	"github.com/goccy/go-json"
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
	"videogame/common/msg_struct"
)

type SGAnswerBacx struct {
	SGAnswerTable
}

// TableName 重写 GORM 的默认表名
func (SGAnswerBacx) TableName() string {
	return "or_answer_bacx" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SGAnswerBacx) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SGAnswerBacx) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *SGAnswerBacx) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 追加一条记录
func (dbTable *SGAnswerBacx) AppendToTable(roundId string, subcode string, answer *msg_struct.GameOpenAnswerBacx, dealer string) error {

	bytes, err := json.Marshal(&answer.BacxAnswerCard)

	if err != nil {
		return err
	}

	nowTime := time.Now()
	dbTable.AnswerTime = &nowTime

	dbTable.Type = subcode

	dbTable.GameId = answer.GameId
	dbTable.DeskNo = answer.DeskNo
	dbTable.Set = answer.Set
	dbTable.Round = answer.Round
	dbTable.Win = answer.Winner.Winner

	dbTable.Value1 = answer.Winner.PPair
	dbTable.Value2 = answer.Winner.BPair

	dbTable.Dealer = dealer
	dbTable.Cards = bytes
	dbTable.RoundId = roundId

	err = dbTable.Create()

	if err != nil {

	}

	return err

}
