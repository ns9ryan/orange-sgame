package walletService

import (
	"github.com/duanhf2012/origin/v2/log"
	"time"
	"videogame/common/proto/rpc"
)

// 调用这个RPC，等于设置是否正在读取这个表的数据，用于结算服的通知，表示此表正在结算中
func (self *WalletService) RPCSetReadingTableFlag(setStatus *rpc.Bool, ret *rpc.Int) error {

	//是否需要使用原子操作
	self.ReadingTableStatus = setStatus.Value

	ret.Value = 1

	nowTime := time.Now()

	log.Infof("RPCSetReadingTableFlag :Time:%s", nowTime.String())

	return nil
}
