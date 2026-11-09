package player

import (
	"encoding/json"
	"github.com/duanhf2012/origin/v2/log"
	"google.golang.org/protobuf/proto"
	"time"
	"videogame/common/global/variable"
	models "videogame/common/model"
	"videogame/common/proto/msg"
	"videogame/service/interfacedef"
	"fmt"
	"github.com/duanhf2012/origin/v2/service"
	redisId "videogame/common/redis"
)

type Player struct {
	service.Module	

	interfacedef.IPlayerTimer
	interfacedef.IGSService
	interfacedef.IMsgSender
	interfacedef.IMsgReceiver
	interfacedef.IPlayer

	PoolObj
	DataInfo
	UserInfo

	pingTime time.Time //ping时间
}

func (p *Player) Init(id string, sender interfacedef.IMsgSender, gsService interfacedef.IGSService, playerTimer interfacedef.IPlayerTimer, receive interfacedef.IMsgReceiver) {

	p.IMsgSender = sender
	p.IGSService = gsService
	p.IPlayerTimer = playerTimer
	p.IMsgReceiver = receive
	p.Id = id
	p.UserId = id

}
func (p *Player) Reset() {
	*p = Player{}
}

func (p *Player) SendMsg(msgType msg.MsgType, message proto.Message) int {
	return p.IMsgSender.SendToClient(p.GetClientId(), msgType, message)
}

func (p *Player) Destroy() {
}

// 初始化玩家配置信息，保存在内存里
func (p *Player) InitBalance() bool {

	if p.MerchantKey == "" {

		fmt.Printf("=======大厅RPCWalletInit 调用了单钱包InitBalance()=========\n")

		//从数据库读取玩家配置，保存在内存中，后面要用MerchantCode获取Token和Url等
		playerInfo, err1 := models.GetUserWalletByUid(p.UserId)
		if err1 != nil {
			return false
		}

		if playerInfo.Digits < 1 {
			playerInfo.Digits = 1
		}

		p.Digits = playerInfo.Digits
		p.Code = playerInfo.Currency
		p.MerchantKey = playerInfo.MerchantId

		fmt.Printf("=======大厅调用了单钱包InitBalance, p.Code=%s, p.MerchantKey=%s\n", p.Code, p.MerchantKey)
	}

	return true

}

func (p *Player) SaveWalletToRedis() bool {

	walletKey := redisId.RedisId_PlayerWallet + p.UserId

	var userWallet redisId.RedisUserWallet

	userWallet.UserId = p.UserId
	userWallet.Credit = p.Credit
	userWallet.Digits = p.Digits
	userWallet.Currency = p.Currency

	bytes, err := json.Marshal(userWallet)

	if err != nil {
		return false
	}

	// 尝试第一次设置
	if err = variable.RedisStorage.Set(walletKey, bytes, 0); err != nil {

		log.SError("-------->钱包服第一次写入redis失败，err:", err.Error())

		// 再尝试第二次
		if err = variable.RedisStorage.Set(walletKey, bytes, 0); err != nil {
			log.SError("-------->钱包服第二次写入redis失败，准备删除，err: ", err.Error())

			// 删除 Redis 数据
			if res, delErr := variable.RedisStorage.Del(walletKey); delErr != nil {
				log.SError("-------->钱包服删除redis失败，err:", delErr.Error())

				return false

			} else {
				if res <= 0 {
					log.SWarn("-------->钱包服删除redis失败，userId:", userWallet.UserId, " ,walletKey:", walletKey, "不存在")
				}
			}
		}
	}

	return true

}

func (p *Player) GetBalanceF() float64 {

	if p.Digits < 1 {
		p.Digits = 1
	}

	return float64(p.Balance) / float64(p.Digits)
}

func (p *Player) BalanceAdd(point uint64) uint64 {

	p.Balance += point

	return p.Balance
}

func (p *Player) BalanceDec(point uint64) uint64 {

	p.Balance -= point

	return p.Balance
}
