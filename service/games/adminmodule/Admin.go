package adminmodule

import (
	"time"
	"videogame/service/interfacedef"
)

type Admin struct {
	interfacedef.IGSService
	PoolObj
	DataInfo
	AdminInfo

	UserId   string //荷官的ID
	UserName string //荷官的名字

	pingTime time.Time //ping时间

	loginTime time.Time //登陆的时间
}

func (p *Admin) Init(id string, gsService interfacedef.IGSService) {

	p.IGSService = gsService
	p.Id = id

	p.loginTime = time.Now()

}

func (p *Admin) Reset() {
	*p = Admin{}
}

func (p *Admin) Destroy() {
}
