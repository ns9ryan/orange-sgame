package models

/******************************************
  龙虎的开答表

******************************************/

import (
	"encoding/json"
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
	"videogame/common/msg_struct"
)

type SGAnswerDTX struct {
	SGAnswerTable
}

// TableName 重写 GORM 的默认表名
func (SGAnswerDTX) TableName() string {
	return "or_answer_dtx" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SGAnswerDTX) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SGAnswerDTX) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *SGAnswerDTX) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 追加一条记录
func (dbTable *SGAnswerDTX) AppendToTable(roundId string, subcode string, answer *msg_struct.GameOpenAnswerDTX, dealer string) error {

	bytes, err := json.Marshal(&answer.DTXAnswerCard)

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

	dbTable.Win = answer.Winner
	dbTable.Dealer = dealer
	dbTable.Cards = bytes
	dbTable.RoundId = roundId

	err = dbTable.Create()

	return err

}
