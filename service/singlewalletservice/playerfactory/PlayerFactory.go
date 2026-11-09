package factory

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/sync"
	"reflect"
	"runtime"
	"videogame/service/interfacedef"
	"videogame/service/singlewalletservice/player"

	"time"
)

var playerPool = sync.NewPoolEx(make(chan sync.IPoolData, 1000), func() sync.IPoolData {
	return &player.Player{}
})

type PlayerFactoryModule struct {
	service.Module

	// mapPlayerMutex ssync.RWMutex

	mapPlayer map[string]*player.Player

	//Analyzer  *performance.PerformanceAnalyzer
}

func NewPlayerFactoryModule() *PlayerFactoryModule {
	return &PlayerFactoryModule{}
}

func (m *PlayerFactoryModule) OnInit() error {
	m.mapPlayer = make(map[string]*player.Player, 2048)
	return nil
}

func (m *PlayerFactoryModule) NewPlayer(id string, sender interfacedef.IMsgSender, gsService interfacedef.IGSService, receive interfacedef.IMsgReceiver) *player.Player {

	p := playerPool.Get().(*player.Player)

	p.Init(id, sender, gsService, m, receive)

	m.mapPlayer[id] = p

	return p
}

func (m *PlayerFactoryModule) ReleasePlayer(object *player.Player) {

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
				log.StackErrorf(fmt.Sprint("GameObjectFactoryModule SafeTimerTicker -> CancelTimerId[", tickerId, "] failed"))
			}
		}
	})
}

func (m *PlayerFactoryModule) SafeCancelTimer(timerId *uint64) {
	if m.CancelTimerId(timerId) == false {
		log.StackErrorf(fmt.Sprint("GameObjectFactoryModule SafeCancelTimer[", timerId, "] failed"))
	}
}

func (m *PlayerFactoryModule) GetPlayer(id string) *player.Player {
	return m.mapPlayer[id]
}

func (m *PlayerFactoryModule) GetPlayerNum() int {
	return len(m.mapPlayer)
}

func (m *PlayerFactoryModule) GetAllPlayer() map[string]*player.Player {
	return m.mapPlayer
}
