package botctrlservice

import (
	"fmt"
	"time"
	"videogame/common/msg_struct"
)

var CurrentStageNum = 0
var CurrentShuffleNum = 0

const StageNum = 4   //每场4局,每到4局调用breakTime
const ShuffleNum = 2 //每2局洗一次牌，没到2局调用洗牌，请修改程序 if 语句

func init() {

	mapRegisterMsg = make(map[int]*MsgEvent)
	//regMsg(msg.MsgType_NotifyLogout, &msg.RetMsg{}, notifyLogout)
	regMsg(msg_struct.Pong, &msg_struct.MsgPong{}, pong)
	regMsg(msg_struct.LoginResCtrlProtocol, &msg_struct.RetMsg{}, loginRes)
	regMsg(msg_struct.GameDealerJoinProtocolResult, &msg_struct.GameDealerJoinResult{}, retCtrlDealerJoin)

	regMsg(msg_struct.GameBacxConfigProtocol, &msg_struct.GameToCtrlGameConfig{}, retCtrlGameConfig)
	regMsg(msg_struct.GameBacxStatusProtocol, &msg_struct.GameStatusMsg{}, retCtrlGameStatusBacx)
	regMsg(msg_struct.GameBacxOpenAnswerProtocol, &msg_struct.GameOpenAnswerBacx{}, retCtrlGameOpenAnswerBacx)
	regMsg(msg_struct.GameBacxAnswerPeriodicProtocol, &msg_struct.GameBacXAnswersPeriodic{}, retCtrlGameAnswersPerioBacx)

	regMsg(msg_struct.GameDtxConfigProtocol, &msg_struct.GameToCtrlGameConfig{}, retCtrlGameConfig)
	regMsg(msg_struct.GameDtxStatusProtocol, &msg_struct.GameStatusMsg{}, retCtrlGameStatusDTX)
	regMsg(msg_struct.GameDtxOpenAnswerProtocol, &msg_struct.GameOpenAnswerDTX{}, retCtrlGameOpenAnswerDTX)
	regMsg(msg_struct.GameDtxAnswerPeriodicProtocol, &msg_struct.GameDtxAnswersPeriodic{}, retCtrlGameAnswersPerioDTX)

	regMsg(msg_struct.GameSicboConfigProtocol, &msg_struct.GameToCtrlGameConfig{}, retCtrlGameConfig)
	regMsg(msg_struct.GameSicboStatusProtocol, &msg_struct.GameStatusMsg{}, retCtrlGameStatusSicbo)
	regMsg(msg_struct.GameSicboOpenAnswerProtocol, &msg_struct.SicboOpenAnswer{}, retCtrlGameOpenAnswerSicbo)
	regMsg(msg_struct.GameSicboAnswerPeriodicProtocol, &msg_struct.SicboAnswersPeriodic{}, retCtrlGameAnswersPerioSicbo)

	regMsg(msg_struct.GameFtanConfigProtocol, &msg_struct.GameToCtrlGameConfig{}, retCtrlGameConfig)
	regMsg(msg_struct.GameFtanStatusProtocol, &msg_struct.GameStatusMsg{}, retCtrlGameStatusFtan)
	regMsg(msg_struct.GameFtanOpenAnswerProtocol, &msg_struct.FtanOpenAnswer{}, retCtrlGameOpenAnswerFtan)
	regMsg(msg_struct.GameFtanAnswerPeriodicProtocol, &msg_struct.FtanAnswersPeriodic{}, retCtrlGameAnswersPerioFtan)

	regMsg(msg_struct.GameSdieConfigProtocol, &msg_struct.GameToCtrlGameConfig{}, retCtrlGameConfig)
	regMsg(msg_struct.GameSdieStatusProtocol, &msg_struct.GameStatusMsg{}, retCtrlGameStatusSdie)
	regMsg(msg_struct.GameSdieOpenAnswerProtocol, &msg_struct.SdieOpenAnswer{}, retCtrlGameOpenAnswerSdie)
	regMsg(msg_struct.GameSdieAnswerPeriodicProtocol, &msg_struct.SdieAnswersPeriodic{}, retCtrlGameAnswersPerioSdie)

	regMsg(msg_struct.GameFscConfigProtocol, &msg_struct.GameToCtrlGameConfig{}, retCtrlGameConfig)
	regMsg(msg_struct.GameFscStatusProtocol, &msg_struct.GameStatusMsg{}, retCtrlGameStatusFSC)
	regMsg(msg_struct.GameFscOpenAnswerProtocol, &msg_struct.FscOpenAnswer{}, retCtrlGameOpenAnswerFSC)
	regMsg(msg_struct.GameFscAnswerPeriodicProtocol, &msg_struct.FscAnswersPeriodic{}, retCtrlGameAnswersPerioFSC)

}

func loginRes(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.RetMsg)

	fmt.Printf("botctrl========================= 【游戏=>控端】登录Data=%v \n", *retMsg)

	//登录成功2秒后，通知游戏开始下注
	//	time.Sleep(2 * time.Second)
	//bot.startBet()
}

func retCtrlDealerJoin(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameDealerJoinResult)

	fmt.Printf("botctrl========================= 【荷管加入回复 %v \n", *retMsg)
}

func retCtrlGameConfig(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameToCtrlGameConfig)

	//fmt.Printf("botctrl========================= retCtrlGameConfig =%v \n", *retMsg)

	bot.GameId = retMsg.GameId
	bot.DeskNo = retMsg.DeskNo

	//启动荷管加入
	time.AfterFunc(time.Second*2, bot.dealerJoin)

}

func retCtrlGameStatusBacx(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameStatusMsg)

	fmt.Printf("botctrl========================= retCtrlGameStatus =%v \n", *retMsg)

	bot.GameStatus = retMsg.GameStatus
	bot.Set = retMsg.Set
	bot.Round = retMsg.Round

	//检测游戏状态， 若是收到停止下注，则需要开奖
	if retMsg.GameStatus == msg_struct.GameStatusStopBet {

		fmt.Printf("botctrl=========================停止下注收到。。。set:%d round:%d \n", retMsg.Set, retMsg.Round)

		//若还没有开答，刚发送开奖结果
		if retMsg.IsAnswered == 0 {

			fmt.Println("启动开答。。。")
			time.AfterFunc(time.Second*3, bot.OpenAnswerBacx)

		} else {

			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	} else if retMsg.GameStatus == msg_struct.GameStatusBet { //下注中

		fmt.Printf("服务器下注中。。。。 set :%d, round:%d \n ", retMsg.Set, retMsg.Round)

	} else if retMsg.GameStatus == msg_struct.GameStatusInterMission { //洗牌中

		fmt.Printf("####botctrl=========================洗牌中收到。。。 \n")

		if retMsg.Set > 24*3*3 {

			time.AfterFunc(time.Second*30, bot.SleepTime)

			return
		}

		time.AfterFunc(time.Second*60, bot.StartBet)

	} else if retMsg.GameStatus == msg_struct.GameStatusSleep { //休场中

		fmt.Println("服务器休场中。。。。")

		bot.SleepNum++

		if bot.SleepNum > 1 {
			time.AfterFunc(time.Second*60*10, bot.StartBet)
		} else {
			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	}

}

func retCtrlGameStatusDTX(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameStatusMsg)

	fmt.Printf("botctrl========================= retCtrlGameStatus DTX =%v \n", *retMsg)

	bot.GameStatus = retMsg.GameStatus
	bot.Set = retMsg.Set
	bot.Round = retMsg.Round

	//检测游戏状态， 若是收到停止下注，则需要开奖
	if retMsg.GameStatus == msg_struct.GameStatusStopBet {

		fmt.Printf("botctrl=========================停止下注收到。。。set:%d round:%d \n", retMsg.Set, retMsg.Round)

		//若还没有开答，刚发送开奖结果
		if retMsg.IsAnswered == 0 {

			fmt.Println("启动开答。。。")
			time.AfterFunc(time.Second*3, bot.OpenAnswerDTX)

		} else {

			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	} else if retMsg.GameStatus == msg_struct.GameStatusBet { //下注中

		fmt.Printf("服务器下注中。。。。 set :%d, round:%d \n ", retMsg.Set, retMsg.Round)

	} else if retMsg.GameStatus == msg_struct.GameStatusInterMission { //洗牌中

		fmt.Printf("####botctrl=========================洗牌中收到。。。 \n")

		if retMsg.Set > 24*3*3 {

			time.AfterFunc(time.Second*30, bot.SleepTime)

			return
		}

		time.AfterFunc(time.Second*60, bot.StartBet)

	} else if retMsg.GameStatus == msg_struct.GameStatusSleep { //休场中

		fmt.Println("服务器休场中。。。。")

		//time.AfterFunc(time.Second*60*10, bot.StartBet)

		bot.SleepNum++

		if bot.SleepNum > 1 {
			time.AfterFunc(time.Second*60*10, bot.StartBet)
		} else {
			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	}

}

func retCtrlGameOpenAnswerBacx(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameOpenAnswerBacx)

	fmt.Printf("botctrl===========*******============== retCtrlGameOpenAnswer =%v \n", *retMsg)

	if bot.Round >= 60 {
		//若检测超出局数了，则进入洗牌状态
		bot.Shuffle()
		return
	}

	time.AfterFunc(time.Second*5, bot.StartBet)

}

func retCtrlGameOpenAnswerDTX(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameOpenAnswerDTX)

	fmt.Printf("botctrl===========*******============== retCtrlGameOpenAnswer =%v \n", *retMsg)

	if bot.Round >= 60 {
		//若检测超出局数了，则进入洗牌状态
		bot.Shuffle()
		return
	}

	time.AfterFunc(time.Second*5, bot.StartBet)

}

func retCtrlGameAnswersPerioBacx(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameBacXAnswersPeriodic)

	fmt.Printf("botctrl========================= retCtrlGameAnswersPerio DTX =%v \n", *retMsg)

}

func retCtrlGameAnswersPerioDTX(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameDtxAnswersPeriodic)

	fmt.Printf("botctrl========================= retCtrlGameAnswersPerio DTX  =%v \n", *retMsg)

}

func retCtrlGameStatusSicbo(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameStatusMsg)

	fmt.Printf("botctrl========================= retCtrlGameStatus =%v \n", *retMsg)

	bot.GameStatus = retMsg.GameStatus
	bot.Set = retMsg.Set
	bot.Round = retMsg.Round

	//检测游戏状态， 若是收到停止下注，则需要开奖
	if retMsg.GameStatus == msg_struct.GameStatusStopBet {

		fmt.Printf("botctrl=========================停止下注收到。。。set:%d round:%d \n", retMsg.Set, retMsg.Round)

		//若还没有开答，刚发送开奖结果
		if retMsg.IsAnswered == 0 {

			fmt.Println("启动开答。。。")
			time.AfterFunc(time.Second*3, bot.OpenAnswerSicbo)

		} else {

			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	} else if retMsg.GameStatus == msg_struct.GameStatusBet { //下注中

		fmt.Printf("服务器下注中。。。。 set :%d, round:%d \n ", retMsg.Set, retMsg.Round)

	} else if retMsg.GameStatus == msg_struct.GameStatusInterMission { //洗牌中

		fmt.Printf("####botctrl=========================洗牌中收到。。。轮盘是不可能收到的 \n")

	} else if retMsg.GameStatus == msg_struct.GameStatusSleep { //休场中

		fmt.Println("服务器休场中。。。。")

		//		time.AfterFunc(time.Second*60*10, bot.StartBet)
		bot.SleepNum++

		if bot.SleepNum > 1 {
			time.AfterFunc(time.Second*60*10, bot.StartBet)
		} else {
			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	}

}

func retCtrlGameOpenAnswerSicbo(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.SicboOpenAnswer)

	fmt.Printf("botctrl===========*******============== retCtrlGameOpenAnswer =%v \n", *retMsg)

	if bot.Round >= 1000 {

		time.AfterFunc(time.Second*10, bot.SleepTime)

	} else {
		time.AfterFunc(time.Second*60*10, bot.StartBet)
	}

}

func retCtrlGameAnswersPerioSicbo(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.SicboAnswersPeriodic)

	fmt.Printf("botctrl========================= retCtrlGameAnswersPerio =%v \n", *retMsg)

}

func pong(bot *Bot, msg interface{}) {
	msgPong := msg.(*msg_struct.MsgPong)
	_ = msgPong
}

func retCtrlGameStatusFtan(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameStatusMsg)

	fmt.Printf("-------Ftan----->botctrl=========================游戏状态收到 retMsg=%v \n", *retMsg)

	bot.GameStatus = retMsg.GameStatus
	bot.Set = retMsg.Set
	bot.Round = retMsg.Round

	//检测游戏状态， 若是收到停止下注，则需要开奖
	if retMsg.GameStatus == msg_struct.GameStatusStopBet {

		fmt.Printf("botctrl=========================停止下注收到。。。set:%d round:%d \n", retMsg.Set, retMsg.Round)

		//若还没有开答，刚发送开奖结果
		if retMsg.IsAnswered == 0 {

			fmt.Println("启动开答。。。")
			time.AfterFunc(time.Second*3, bot.OpenAnswerFtan)

		} else {

			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	} else if retMsg.GameStatus == msg_struct.GameStatusBet { //下注中

		fmt.Printf("服务器下注中。。。。 set :%d, round:%d \n ", retMsg.Set, retMsg.Round)

	} else if retMsg.GameStatus == msg_struct.GameStatusInterMission { //洗牌中

		fmt.Printf("####botctrl=========================洗牌中收到。。。反滩是不可能收到的 \n")

	} else if retMsg.GameStatus == msg_struct.GameStatusSleep { //休场中

		fmt.Println("服务器休场中。。。。")

		//		time.AfterFunc(time.Second*60*10, bot.StartBet)
		bot.SleepNum++

		if bot.SleepNum > 1 {
			time.AfterFunc(time.Second*60*10, bot.StartBet)
		} else {
			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	}

}

func retCtrlGameOpenAnswerFtan(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.FtanOpenAnswer)
	_ = retMsg

	//fmt.Printf("-------Ftan----->botctrl=========================开达收到。。。 retMsg=%v \n", *retMsg)

	if bot.Round >= 1000 {

		time.AfterFunc(time.Second*10, bot.SleepTime)

	} else {
		time.AfterFunc(time.Second*60*10, bot.StartBet)
	}

}

func retCtrlGameAnswersPerioFtan(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.FtanAnswersPeriodic)

	//fmt.Printf("-------Ftan----->botctrl=========================历史开达收到。。。 retMsg==%v \n", *retMsg)
	_ = retMsg

}

func retCtrlGameStatusSdie(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameStatusMsg)

	//fmt.Printf("-------Sdie----->botctrl=========================游戏状态收到 retMsg=%v \n", *retMsg)
	bot.GameStatus = retMsg.GameStatus
	bot.Set = retMsg.Set
	bot.Round = retMsg.Round

	//检测游戏状态， 若是收到停止下注，则需要开奖
	if retMsg.GameStatus == msg_struct.GameStatusStopBet {

		fmt.Printf("botctrl=========================停止下注收到。。。set:%d round:%d \n", retMsg.Set, retMsg.Round)

		//若还没有开答，刚发送开奖结果
		if retMsg.IsAnswered == 0 {

			fmt.Println("启动开答。。。")
			time.AfterFunc(time.Second*3, bot.OpenAnswerSdie)

		} else {

			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	} else if retMsg.GameStatus == msg_struct.GameStatusBet { //下注中

		fmt.Printf("服务器下注中。。。。 set :%d, round:%d \n ", retMsg.Set, retMsg.Round)

	} else if retMsg.GameStatus == msg_struct.GameStatusInterMission { //洗牌中

		fmt.Printf("####botctrl=========================洗牌中收到。。。反滩是不可能收到的 \n")

	} else if retMsg.GameStatus == msg_struct.GameStatusSleep { //休场中

		fmt.Println("服务器休场中。。。。")

		//		time.AfterFunc(time.Second*60*10, bot.StartBet)
		bot.SleepNum++

		if bot.SleepNum > 1 {
			time.AfterFunc(time.Second*60*10, bot.StartBet)
		} else {
			time.AfterFunc(time.Second*3, bot.StartBet)
		}
	}

}

func retCtrlGameOpenAnswerSdie(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.SdieOpenAnswer)

	fmt.Printf("-------Sdie----->botctrl=========================开达收到。。。 Winner=%v \n", retMsg.Winner)

	if bot.Round >= 1000 {

		time.AfterFunc(time.Second*10, bot.SleepTime)

	} else {
		time.AfterFunc(time.Second*3, bot.StartBet)
	}

}

func retCtrlGameAnswersPerioSdie(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.SdieAnswersPeriodic)

	//fmt.Printf("-------Sdie----->botctrl=========================历史开达收到。。。 retMsg==%v \n", *retMsg)
	_ = retMsg

}

func retCtrlGameStatusFSC(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.GameStatusMsg)

	//fmt.Printf("-------Sdie----->botctrl=========================游戏状态收到 retMsg=%v \n", *retMsg)
	bot.GameStatus = retMsg.GameStatus
	bot.Set = retMsg.Set
	bot.Round = retMsg.Round

	//检测游戏状态， 若是收到停止下注，则需要开奖
	if retMsg.GameStatus == msg_struct.GameStatusStopBet {

		fmt.Printf("botctrl=========================停止下注收到。。。set:%d round:%d \n", retMsg.Set, retMsg.Round)

		//若还没有开答，刚发送开奖结果
		if retMsg.IsAnswered == 0 {

			fmt.Println("启动开答。。。")
			time.AfterFunc(time.Second*3, bot.OpenAnswerFSC)

		} else {

			time.AfterFunc(time.Second*3, bot.StartBet)
		}

	} else if retMsg.GameStatus == msg_struct.GameStatusBet { //下注中

		fmt.Printf("服务器下注中。。。。 set :%d, round:%d \n ", retMsg.Set, retMsg.Round)

	} else if retMsg.GameStatus == msg_struct.GameStatusInterMission { //洗牌中

		fmt.Printf("####botctrl=========================洗牌中收到。。。反滩是不可能收到的 \n")

	} else if retMsg.GameStatus == msg_struct.GameStatusSleep { //休场中

		fmt.Println("服务器休场中。。。。")

		//		time.AfterFunc(time.Second*60*10, bot.StartBet)
		bot.SleepNum++

		if bot.SleepNum > 1 {
			time.AfterFunc(time.Second*60*10, bot.StartBet)
		} else {
			time.AfterFunc(time.Second*3, bot.StartBet)
		}
	}

}

func retCtrlGameOpenAnswerFSC(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.FscOpenAnswer)

	fmt.Printf("-------FSC----->botctrl=========================开达收到。。。 Winner=%v \n", retMsg.Winner)

	if bot.Round >= 1000 {

		time.AfterFunc(time.Second*10, bot.SleepTime)

	} else {
		time.AfterFunc(time.Second*3, bot.StartBet)
	}

}

func retCtrlGameAnswersPerioFSC(bot *Bot, msg interface{}) {

	retMsg := msg.(*msg_struct.FscAnswersPeriodic)

	//fmt.Printf("-------Sdie----->botctrl=========================历史开达收到。。。 retMsg==%v \n", *retMsg)
	_ = retMsg

}
