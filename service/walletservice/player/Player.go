package player

import (
	"encoding/json"
	"github.com/duanhf2012/origin/v2/log"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
	"videogame/common/global/variable"
	models "videogame/common/model"
	"videogame/common/proto/msg"
	redisId "videogame/common/redis"
	"videogame/service/interfacedef"
)

type Player struct {
	interfacedef.IPlayerTimer
	interfacedef.IGSService
	interfacedef.IMsgSender
	interfacedef.IMsgReceiver
	interfacedef.IPlayer

	PoolObj
	DataInfo
	UserInfo

	pingTime time.Time //ping时间

	optMutex sync.RWMutex //用于对点数的操作

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

// 初始化余额，从redis读取
func (p *Player) InitBalance() bool {

	var wallet redisId.RedisUserWallet

	walletkey := redisId.RedisId_PlayerWallet + p.Id
	bytes, err := variable.RedisStorage.Get(walletkey)

	//读取redis的钱包
	if err == nil {

		err = json.Unmarshal([]byte(bytes), &wallet)

		if err == nil {

			p.Digits = wallet.Digits
			p.Balance = uint64(wallet.Credit)
			p.Currency = wallet.Currency

			if p.Digits == 0 {
				p.Digits = 1
			}
			return true
		}
	}

	//从数据库读取
	playerInfo, err1 := models.GetUserWalletByUid(p.UserId)

	if err1 != nil {
		return false
	}

	if playerInfo.Digits < 1 {
		playerInfo.Digits = 1
	}

	p.Balance = uint64(playerInfo.Balance)
	p.Digits = playerInfo.Digits
	p.Currency = playerInfo.Currency

	p.SaveWalletToRedis()

	return false

}

func (p *Player) SaveWalletToRedis() bool {

	walletKey := redisId.RedisId_PlayerWallet + p.UserId

	var userWallet redisId.RedisUserWallet

	userWallet.UserId = p.UserId
	userWallet.Credit = uint(p.Balance)
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

// 检测余额是否够
func (p *Player) BalanceIsEnough(point uint64) bool {

	p.optMutex.RLock()
	defer p.optMutex.RUnlock()

	if p.Balance >= point {
		return true
	}

	return false
}

func (p *Player) BalanceAdd(point uint64) uint64 {

	p.optMutex.Lock()
	defer p.optMutex.Unlock()

	p.Balance += point

	return p.Balance
}

func (p *Player) BalanceDec(point uint64) uint64 {

	p.optMutex.Lock()
	defer p.optMutex.Unlock()

	p.Balance -= point

	return p.Balance
}

func (p *Player) GetBalance() uint64 {

	p.optMutex.RLock()
	defer p.optMutex.RUnlock()

	return p.Balance
}

func (p *Player) GetBalanceF() float64 {

	p.optMutex.RLock()
	defer p.optMutex.RUnlock()

	if p.Digits < 1 {
		p.Digits = 1
	}

	return float64(p.Balance) / float64(p.Digits)
}
