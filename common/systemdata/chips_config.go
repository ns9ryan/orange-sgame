package systemdata

import (
	models "videogame/common/model"
)

// 抽码的列表信息
type ChipItem struct {
	Id         int
	ChipValue  uint
	ChipDigits uint
	ChipFValue float32
	ChipName   string
	Status     int
}

var SystemChips map[uint]*ChipItem

// 初始化系统抽码列表，从数据表读取全部
func InitChipsConfig() {

	SystemChips = make(map[uint]*ChipItem)

	var chipsTable models.SChipsConfig

	chipList, err := chipsTable.GetGhipConfig()

	if err != nil {
		return
	}

	for _, chipItem := range chipList {

		SystemChips[chipItem.ID] = new(ChipItem)

		SystemChips[chipItem.ID].ChipValue = chipItem.Value
		SystemChips[chipItem.ID].ChipDigits = chipItem.Digits
		SystemChips[chipItem.ID].ChipName = chipItem.Name
	}

}

// 读取抽码值
func SystemChipsGetChip(key uint) (*ChipItem, bool) {

	item, ok := SystemChips[key]

	if !ok {
		return nil, false
	}

	return item, true

}
