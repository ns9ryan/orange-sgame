package templatelimit

import (
	"sync"

	"github.com/duanhf2012/origin/v2/log"
	models "videogame/common/model"
)

/************************************
 功能:用于读取数据库中的限红模板的限红方案中的ID列表,并保存在内存中,方案查询
*************************************/

type TemplatePlanId struct {
	PlanId string
	Name   string
	MinBet uint // 最小投注
	MaxBet uint // 最大投注
	Digits uint // 小数点
}

var (
	isInitTemplatePlan bool // 是否初始化标志
	TemplatePlans      map[string]*TemplatePlanId
	templatePlanLock   sync.RWMutex // 读写锁
)

func init() {
	TemplatePlans = make(map[string]*TemplatePlanId)
}

// 初始化限红模板
func InitTemplatePlans() bool {

	var templatePlanId models.STemplatePlan

	templatePlanLock.Lock()
	defer templatePlanLock.Unlock()

	if isInitTemplatePlan {
		return true
	}

	TemplatePlans = make(map[string]*TemplatePlanId)

	items, err := templatePlanId.GetAllsData()
	if err != nil {
		return false
	}

	for _, item := range items {
		TemplatePlans[item.PlanId] = &TemplatePlanId{
			PlanId: item.PlanId,
			Name:   item.Name,
			MinBet: item.MinValue,
			MaxBet: item.MaxValue,
			Digits: item.Digits,
		}
	}

	log.Infof("读取限红ID表个数:%d", len(items))

	isInitTemplatePlan = true

	return true
}

// 读取指定的限红ID的数据
func GetTemplatePlanData(planeId string) (*TemplatePlanId, bool) {
	templatePlanLock.RLock()
	defer templatePlanLock.RUnlock()

	item, ok := TemplatePlans[planeId]
	if !ok {
		return nil, false
	}
	return item, true
}

// 设置指定的限红ID的数据
func SetTemplatePlanData(planeId string) (*TemplatePlanId, bool) {

	var templatePlanId models.STemplatePlan

	item, err := templatePlanId.GetPlanData(planeId)

	if err != nil {
		return nil, false
	}

	templatePlan := &TemplatePlanId{
		PlanId: item.PlanId,
		Name:   item.Name,
		MinBet: item.MinValue,
		MaxBet: item.MaxValue,
		Digits: item.Digits,
	}

	templatePlanLock.Lock()
	defer templatePlanLock.Unlock()

	TemplatePlans[item.PlanId] = templatePlan
	return templatePlan, true
}

// 重新载入模板ID
func ReGetTemplatePlanData(planeId string) (*TemplatePlanId, bool) {

	templatePlan, ok := GetTemplatePlanData(planeId)

	if ok == false {
		return SetTemplatePlanData(planeId)
	}

	var templatePlanId models.STemplatePlan

	item, err := templatePlanId.GetPlanData(planeId)

	if err != nil {
		return nil, false
	}

	templatePlan.MinBet = item.MinValue
	templatePlan.MaxBet = item.MaxValue
	templatePlan.Digits = item.Digits

	return templatePlan, true
}
