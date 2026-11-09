package factory

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/sync"
	"reflect"
	"runtime"
	syncc "sync"
	"time"
	"videogame/service/games/player"
	"videogame/service/interfacedef"
)

var playerPool = sync.NewPoolEx(make(chan sync.IPoolData, 1000), func() sync.IPoolData {
	return &player.Player{}
})

type PlayerFactoryModule struct {
	service.Module

	playerMutex syncc.RWMutex

	mapPlayer map[string]*player.Player // map

	//Analyzer  *performance.PerformanceAnalyzer
}

func NewPlayerFactoryImp() *PlayerFactoryModule {
	return &PlayerFactoryModule{}
}

func (m *PlayerFactoryModule) OnInit() error {

	m.mapPlayer = make(map[string]*player.Player, 2048)

	return nil
}

func (m *PlayerFactoryModule) NewPlayer(id string, sender interfacedef.IMsgSender, gsService interfacedef.IGSService, receive interfacedef.IMsgReceiver) *player.Player {

	m.playerMutex.Lock()
	defer m.playerMutex.Unlock()

	p := playerPool.Get().(*player.Player)

	p.Init(id, sender, gsService, m, receive)

	m.mapPlayer[id] = p

	return p
}

func (m *PlayerFactoryModule) ReleasePlayer(object *player.Player) {
	//log.SInfo("player[", object.GetId(), "] release")

	m.playerMutex.Lock()
	defer m.playerMutex.Unlock()

	delete(m.mapPlayer, object.GetId())
	playerPool.Put(object)

}

func (m *PlayerFactoryModule) SafeTimerAfter(timerId *uint64, objectId string, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	id := *timerId

	m.SafeAfterFunc(timerId, d, AdditionData, func(uint64, interface{}) {
		_, ok := m.mapPlayer[objectId]
		if ok == true {
			cb(id, AdditionData)
		} else {
			funName := runtime.FuncForPC(reflect.ValueOf(cb).Pointer()).Name()
			log.SError("GameObjectFactoryModule TimerAfter obj[", objectId, "] not be found", ",funName", funName)
		}
	})
}

func (m *PlayerFactoryModule) SafeTimerTicker(tickerId *uint64, objectId string, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	m.SafeNewTicker(tickerId, d, AdditionData, func(uint64, interface{}) {
		_, ok := m.mapPlayer[objectId]
		if ok == true {
			cb(*tickerId, AdditionData)
		} else {
			funName := runtime.FuncForPC(reflect.ValueOf(cb).Pointer()).Name()
			log.SError("GameObjectFactoryModule TimerTicker obj[", objectId, "] not be found", ",funName", funName)
			if m.CancelTimerId(tickerId) == false {
				log.StackError(fmt.Sprint("GameObjectFactoryModule SafeTimerTicker -> CancelTimerId[", tickerId, "] failed"))
			}
		}
	})
}

func (m *PlayerFactoryModule) SafeCancelTimer(timerId *uint64) {
	if m.CancelTimerId(timerId) == false {
		log.StackError(fmt.Sprint("GameObjectFactoryModule SafeCancelTimer[", timerId, "] failed"))
	}
}

func (m *PlayerFactoryModule) GetPlayer(id string) *player.Player {

	m.playerMutex.RLock()
	defer m.playerMutex.RUnlock()

	return m.mapPlayer[id]

}

func (m *PlayerFactoryModule) GetPlayerNum() int {

	m.playerMutex.RLock()
	defer m.playerMutex.RUnlock()

	return len(m.mapPlayer)
}

func (m *PlayerFactoryModule) GetAllPlayer() map[string]*player.Player {

	m.playerMutex.RLock()
	defer m.playerMutex.RUnlock()

	return m.mapPlayer
}

func (m *PlayerFactoryModule) ReportPlayerInfo(gameId string) {

	num := len(m.mapPlayer)

	//则是没有玩家，则不打出
	if num == 0 {
		return
	}

	fmt.Printf("=================%s:在线玩家=== 人数:%d===============\n", gameId, num)

	idx := 0

	for _, player := range m.mapPlayer {

		if player.GetOnline() == false {
			continue
		}

		fmt.Printf("在线玩家:%d  user:%s ,clientId:%s \n", idx, player.UserId, player.ClientId)

		idx++
	}

	fmt.Println("=====================================================\n")

}
