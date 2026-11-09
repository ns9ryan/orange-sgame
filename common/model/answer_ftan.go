package models

/******************************************
  番摊的开答表

******************************************/

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
	"videogame/common/msg_struct"
)

type SGAnswerFtan struct {
	SGAnswerTable
}

// TableName 重写 GORM 的默认表名
func (SGAnswerFtan) TableName() string {
	return "or_answer_ftan" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SGAnswerFtan) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SGAnswerFtan) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *SGAnswerFtan) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 追加一条记录
func (dbTable *SGAnswerFtan) AppendToTable(roundId string, subcode string, answer *msg_struct.FtanOpenAnswer, dealer string) error {

	nowTime := time.Now()
	dbTable.AnswerTime = &nowTime

	dbTable.Type = subcode
	dbTable.GameId = answer.GameId

	dbTable.DeskNo = answer.DeskNo
	dbTable.Set = answer.Set
	dbTable.Round = answer.Round

	dbTable.Win = answer.Winner
	dbTable.Dealer = dealer
	dbTable.RoundId = roundId

	err := dbTable.Create()

	return err

}
