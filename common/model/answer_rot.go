package models

/******************************************
  龙虎的开答表

******************************************/

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
	"videogame/common/msg_struct"
)

type SGAnswerRot struct {
	SGAnswerTable
}

// TableName 重写 GORM 的默认表名
func (SGAnswerRot) TableName() string {
	return "or_answer_rot" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SGAnswerRot) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SGAnswerRot) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *SGAnswerRot) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 追加一条记录
func (dbTable *SGAnswerRot) AppendToTable(roundId string, subcode string, answer *msg_struct.GameOpenAnswerRot, dealer string) error {

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
