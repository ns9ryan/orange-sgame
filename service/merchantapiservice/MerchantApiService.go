package merchantapiservice

import (
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"videogame/service/merchantapiservice/apimodule"
	"time"
)

func init() {
	node.Setup(&MerchantApiService{})
}

type MerchantApiService struct {
	service.Service
	MerchantApiModule apimodule.MerchantApiModule
}

func (apiserv *MerchantApiService) OnInit() error {
	//添加模块
	apiserv.MerchantApiModule = apimodule.MerchantApiModule{}
	apiserv.AddModule(&apiserv.MerchantApiModule)

	//性能监控
	apiserv.OpenProfiler()
	apiserv.GetProfiler().SetOverTime(time.Millisecond * 100)
	apiserv.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}
