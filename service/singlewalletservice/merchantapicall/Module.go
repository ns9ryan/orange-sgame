package merchantapicall

import (
	"github.com/duanhf2012/origin/v2/service"
)

type MerchantApiModule struct {
	service.Module

	mapMerchant map[string]*Merchant
	RoundAnswer map[string]int32
}

func NewMerchantApiModule() *MerchantApiModule {
	return &MerchantApiModule{}
}

func (m *MerchantApiModule) OnInit() error {
	m.OpenConcurrent(100, 500, 1000000)

	m.mapMerchant = make(map[string]*Merchant, 2048)
	m.RoundAnswer = make(map[string]int32, 2048)
	return nil
}

// type Merchant struct {
//     MerchantCode       string
//     MerchantName       string

//     WalletUrl          string
//     OrToken            string
// }
