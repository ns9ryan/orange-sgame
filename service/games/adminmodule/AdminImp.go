package adminmodule

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/sync"
	"time"
	"videogame/service/interfacedef"
)

var adminPool = sync.NewPoolEx(make(chan sync.IPoolData, 10), func() sync.IPoolData {
	return &Admin{}
})

type AdminImp struct {
	service.Module

	ServerName string //归属哪个服务名

	mapAdmin map[string]*Admin
}

func NewAdminImp(serverName string) *AdminImp {
	return &AdminImp{ServerName: serverName}
}

func (m *AdminImp) OnInit() error {
	m.mapAdmin = make(map[string]*Admin, 10)

	return nil
}

func (m *AdminImp) NewAdmin(id string, gsService interfacedef.IGSService, ip string) *Admin {

	//m.GsService = gsService

	p := adminPool.Get().(*Admin)

	p.Init(id, gsService)

	m.mapAdmin[id] = p

	p.IpAddr = ip

	return p
}

func (m *AdminImp) ReleaseAdmin(userId string) bool {

	admin, ok := m.mapAdmin[userId]

	if ok == false {
		return false
	}

	delete(m.mapAdmin, userId)

	adminPool.Put(admin)

	return true
}

func (m *AdminImp) GetAdmin(id string) *Admin {
	return m.mapAdmin[id]
}

func (m *AdminImp) GetAdminNum() int {
	return len(m.mapAdmin)
}

func (m *AdminImp) GetAllAdmin() map[string]*Admin {
	return m.mapAdmin
}

// 检测帐号是否存在
func (m *AdminImp) AdminIsExist(id string) bool {

	_, ok := m.mapAdmin[id]

	return ok
}

// 报告此通道的信息
func (m *AdminImp) ReportInfo() {

	//控端没有数量，则不输出
	if len(m.mapAdmin) == 0 {
		return
	}

	ntime := time.Now()

	gsname := m.ServerName

	fmt.Printf("======%s ===控端在线列表==%s===================\n", gsname, ntime.Format("2006-01-02 15:04:05"))

	for key, admin := range m.mapAdmin {

		fmt.Printf("*** admin:%s ,ip:%s   clintId:%s  loginTime:%s  \n ", key, admin.IpAddr, admin.ClientId, admin.loginTime.Format("2006-01-02 15:04:05"))

	}

	fmt.Printf("******************************************************\n")

}
