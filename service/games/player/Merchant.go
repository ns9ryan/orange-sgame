package player

import (
	"fmt"
	"sync"
	models "videogame/common/model"
	"videogame/common/util"
)

const (
	APIWallet     = 1 //API调用钱包， 需要调用远程的API服务器
	ClusterWallet = 0 //本地集群钱包
)

var isInitMerchant bool //是否初始化商户了，用于启动载入数据

var MerchantGroupMutex sync.Mutex           //商户读写的锁
var MerchantGroups map[string]*MerchantNode //商户代理列表

type MerchantNode struct {
	MerchantId  string //商户的ID
	MerchantKey uint   //商户的Key

	MerchantName  string //代理姓名
	MerchantCode  string //商户码
	TopMerchantId string //顶级的总商户

	ParentId string //它的上级ID

	Currency string //此商户的主币种
	Digits   uint   //币种的小数位

	WalletType uint //钱包类型( 1单钱包  2 多钱包)
	LevelNo    int  //层级

	Blocked   bool //是否锁了
	BetPaused bool //投注是否暂停

	Commission_rate uint //成数实际要除100
	Water_rate      uint //洗码比，实际要除1000

	ParentNode *MerchantNode //它的上级商户

	ChildMerchants    map[string]*MerchantNode      //它的子商户
	MerchantCurrencys map[string]*MerchantCurrrency //它的币种线,现阶一个商户一个币种线

	//API的回调列表
}

// 商户下面的币种代理
type MerchantCurrrency struct {
	MerchantId string //它自已的ID
	ParentId   string //商户的ID

	Currency string //币种
	Digits   uint   //小数位

	MinTip uint //最小打赏金额
	MaxTip uint //最大打赏金额

	TipItems []uint //打赏金额的列表

	ParentNode *MerchantNode //它归属的商户Node

	Player map[string]*Player //属于它的玩家
}

// 初始化商户代理链
func InitMerchantGroups() bool {

	if isInitMerchant {

		fmt.Printf("111 \n")
		return true
	}

	//读取商户列表
	groups, err := models.GetMerchants("root")

	if err != nil {

		fmt.Printf("222 \n")
		return false
	}

	MerchantGroups = make(map[string]*MerchantNode, len(groups))

	for _, group := range groups {

		mapkey := group.Username

		ParentNode := MerchantNode{}

		ParentNode.ParentId = group.ParentId
		ParentNode.WalletType = group.WalletType
		ParentNode.Blocked = !(util.ConvertIntToBool(int(group.Status)))
		ParentNode.BetPaused = util.ConvertIntToBool(int(group.BetPaused))
		ParentNode.ChildMerchants = nil //后面处理

		ParentNode.MerchantId = group.Username
		ParentNode.MerchantName = group.Nickname
		ParentNode.MerchantCode = group.MerchantCode
		ParentNode.TopMerchantId = group.Username
		ParentNode.MerchantKey = group.RootKey

		//成数,洗码
		ParentNode.Commission_rate = group.CommissionRate
		ParentNode.Water_rate = group.RollingRate

		ParentNode.ParentNode = nil //它没有上级了

		MerchantGroups[mapkey] = &ParentNode

		//去读取商户下面的币种线，现在每个
		childs, ok := ReadMerchantCurrency(mapkey)

		if ok == true {

			MerchantGroups[mapkey].MerchantCurrencys = make(map[string]*MerchantCurrrency, len(childs))

			for _, child := range childs {

				child.ParentNode = &ParentNode //它的归属的商户的NODE

				MerchantGroups[mapkey].MerchantCurrencys[child.MerchantId] = child

				ParentNode.Currency = child.Currency
				ParentNode.Digits = child.Digits
			}

		}

		//再读取它下面商户的，只有二级，因此不需要递归处理
		subgroups, err := models.GetMerchants(mapkey)

		if err != nil {

			fmt.Printf("333 \n")
			continue
		}

		ParentNode.ChildMerchants = make(map[string]*MerchantNode, len(subgroups))

		//分别读取下面子商户的信息
		for _, subgroup := range subgroups {

			mapkey := subgroup.Username

			subnode := MerchantNode{}

			MerchantGroups[mapkey] = &subnode

			MerchantGroups[mapkey].ParentId = subgroup.ParentId
			MerchantGroups[mapkey].MerchantKey = subgroup.RootKey

			MerchantGroups[mapkey].WalletType = subgroup.WalletType
			MerchantGroups[mapkey].Blocked = !(util.ConvertIntToBool(int(subgroup.Status)))
			MerchantGroups[mapkey].BetPaused = util.ConvertIntToBool(int(subgroup.BetPaused))
			MerchantGroups[mapkey].ChildMerchants = nil //后面处理

			MerchantGroups[mapkey].MerchantId = subgroup.Username
			MerchantGroups[mapkey].MerchantName = subgroup.Nickname
			MerchantGroups[mapkey].MerchantCode = subgroup.MerchantCode
			MerchantGroups[mapkey].TopMerchantId = subgroup.ParentId

			//成数和洗码要跟上级一样
			MerchantGroups[mapkey].Commission_rate = ParentNode.Commission_rate
			MerchantGroups[mapkey].Water_rate = ParentNode.Water_rate

			MerchantGroups[mapkey].ParentNode = &ParentNode

			//它有儿子
			ParentNode.ChildMerchants[mapkey] = &subnode

			//去读取商户下面的币种线，现在每个商户只支持
			childs, ok := ReadMerchantCurrency(mapkey)

			if ok == true {

				MerchantGroups[mapkey].MerchantCurrencys = make(map[string]*MerchantCurrrency, len(childs))

				for _, child := range childs {

					child.ParentNode = &subnode
					MerchantGroups[mapkey].MerchantCurrencys[child.MerchantId] = child

					subnode.Currency = child.Currency
					subnode.Digits = child.Digits
				}
			}
		}
	}

	isInitMerchant = true

	fmt.Println("初始化商户成功！！！")

	return true
}

// 重置商户代理读取标志
func ResetReadMerchant() {
	isInitMerchant = false
}

func PrintMerchantGroups() {

	ind := 0

	for k, v := range MerchantGroups {

		fmt.Printf("ind:%d key:%s  merchantId:%s parentId:%s   top:%s \n", ind, k, v.MerchantId, v.ParentId, v.TopMerchantId)

		ind++
	}
}

// 读取商户下面的币种代理线
func ReadMerchantCurrency(parentId string) ([]*MerchantCurrrency, bool) {

	groups, err := models.GetChildMerchantCurrencys(parentId)

	if err != nil {
		return nil, false
	}

	currencys := make([]*MerchantCurrrency, 0)

	for _, group := range groups {

		item := MerchantCurrrency{}

		item.Currency = group.Currency
		item.MerchantId = group.MerchantCurrencyId
		item.ParentId = parentId
		item.Digits = group.Digits
		item.MaxTip = group.MaxTip
		item.MinTip = group.MinTip

		item.TipItems = make([]uint, 0)

		nums := util.GetUintSliceByStrings(group.TipOptions, ",")

		//若有设置打赏金额列表
		if len(nums) > 0 {

			item.TipItems = append(item.TipItems, nums...)

		} else {

			tipitem := []uint{1, 2, 5, 10, 20, 50, 100, 200}

			item.TipItems = append(item.TipItems, tipitem...)
		}

		currencys = append(currencys, &item)
	}

	return currencys, true
}

// 读取商户的信息
func GetMerchantNode(merchantId string) (*MerchantNode, bool) {

	MerchantGroupMutex.Lock()
	defer MerchantGroupMutex.Unlock()

	node, ok := MerchantGroups[merchantId]

	return node, ok
}

// 根据id 得到商户Node
func BuildMerchantNode(merchantId string) (*MerchantNode, bool) {

	var merchantTable models.MerchantGroup

	//到数据表读取信息
	merchantGroup, err := merchantTable.GetMerchantGroup(merchantId)

	if err != nil {
		return nil, false
	}

	merchantNode := MerchantNode{}

	merchantNode.MerchantId = merchantId
	merchantNode.MerchantName = merchantGroup.MerchantName
	merchantNode.TopMerchantId = merchantGroup.ParentId
	merchantNode.WalletType = merchantGroup.WalletType
	merchantNode.MerchantKey = merchantGroup.RootKey
	merchantNode.ParentId = merchantGroup.ParentId
	merchantNode.Commission_rate = merchantGroup.CommissionRate
	merchantNode.Water_rate = merchantGroup.RollingRate
	merchantNode.Blocked = !(util.ConvertIntToBool(int(merchantGroup.Status)))
	merchantNode.BetPaused = util.ConvertIntToBool(int(merchantGroup.BetPaused))

	//去读取商户下面的币种线，现在每个商户只有一条币种
	childs, ok := ReadMerchantCurrency(merchantId)

	if ok == true {

		//系统现在每个商户对应一条币种线,但支持多币种
		merchantNode.MerchantCurrencys = make(map[string]*MerchantCurrrency, len(childs))

		for _, child := range childs {

			child.ParentNode = &merchantNode //它的归属的商户的NODE

			merchantNode.MerchantCurrencys[child.MerchantId] = child

			merchantNode.Currency = child.Currency
			merchantNode.Digits = child.Digits
		}
	} else {

		//没有找到也不能建此线
		return nil, false
	}

	if merchantGroup.ParentId == "root" {
		merchantNode.ParentNode = nil
	} else {

		MerchantGroupMutex.Lock()
		parentNode, ok := MerchantGroups[merchantGroup.ParentId]
		MerchantGroupMutex.Unlock()

		//没有找到上级,说明系统出错了
		if ok == false {
			//需要重新建立上级商户

			//到数据表去读取这个商户信息
			merchantGroup, err = merchantTable.GetMerchantGroup(merchantGroup.ParentId)

			//找不到这个商户
			if err != nil {
				return nil, false
			}

			parentNode = &MerchantNode{}

			parentNode.MerchantId = merchantGroup.Username
			parentNode.MerchantName = merchantGroup.MerchantName
			parentNode.TopMerchantId = merchantGroup.Username
			parentNode.WalletType = merchantGroup.WalletType
			parentNode.MerchantKey = merchantGroup.RootKey
			parentNode.ParentId = merchantGroup.ParentId
			parentNode.Commission_rate = merchantGroup.CommissionRate
			parentNode.Water_rate = merchantGroup.RollingRate
			parentNode.Blocked = !(util.ConvertIntToBool(int(merchantGroup.Status)))
			parentNode.BetPaused = util.ConvertIntToBool(int(merchantGroup.BetPaused))

			parentNode.ChildMerchants = make(map[string]*MerchantNode)

			//把这个加入结点
			MerchantGroupMutex.Lock()
			MerchantGroups[merchantGroup.Username] = parentNode
			MerchantGroupMutex.Unlock()
		}
		merchantNode.ParentNode = parentNode

		if parentNode.ChildMerchants != nil {
			parentNode.ChildMerchants[merchantId] = &merchantNode
		}
	}

	MerchantGroupMutex.Lock()
	MerchantGroups[merchantId] = &merchantNode
	MerchantGroupMutex.Unlock()

	return &merchantNode, true

}

// 重新载入商户的数据
func ReLoadMerchantNode(merchantId string) (*MerchantNode, bool) {

	var merchantTable models.MerchantGroup

	merchantNode, ok := GetMerchantNode(merchantId)

	//内存原先就没有这个商户
	if ok == false {
		return BuildMerchantNode(merchantId)
	}

	//到数据表重新读取数据
	merchantGroup, err := merchantTable.GetMerchantGroup(merchantId)

	if err != nil {
		return nil, false
	}

	merchantNode.MerchantName = merchantGroup.MerchantName
	merchantNode.TopMerchantId = merchantGroup.ParentId
	merchantNode.WalletType = merchantGroup.WalletType
	merchantNode.MerchantKey = merchantGroup.RootKey
	merchantNode.ParentId = merchantGroup.ParentId
	merchantNode.Commission_rate = merchantGroup.CommissionRate
	merchantNode.Water_rate = merchantGroup.RollingRate
	merchantNode.Blocked = !(util.ConvertIntToBool(int(merchantGroup.Status)))
	merchantNode.BetPaused = util.ConvertIntToBool(int(merchantGroup.BetPaused))

	//去读取商户下面的币种线，现在每个商户只有一条币种
	childs, ok := ReadMerchantCurrency(merchantId)

	if ok == true {

		//系统现在每个商户对应一条币种线,但支持多币种

		for _, child := range childs {

			subNode, rok := merchantNode.MerchantCurrencys[child.MerchantId]

			if rok == false {
				continue
			}

			subNode.TipItems = child.TipItems
			subNode.MinTip = child.MinTip
			subNode.MaxTip = child.MaxTip

		}
	}

	return merchantNode, true

}

// 读取商户代码
func (merchant *MerchantNode) GetMerchantCode() string {
	return merchant.MerchantCode
}

// 读取商户的ID
func (merchant *MerchantNode) GetMerchantId() string {
	return merchant.MerchantId
}

func (merchant *MerchantNode) GetRootKey() uint {

	return merchant.MerchantKey
}
