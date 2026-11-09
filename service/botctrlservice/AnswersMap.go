package botctrlservice

import "videogame/common/msg_struct"

var AnswersMapBacx []*msg_struct.GameOpenAnswerBacx
var AnswersMapDTX []*msg_struct.GameOpenAnswerDTX
var AnswersMapSICBO []*msg_struct.SicboOpenAnswer

func InitBacxPairA() *msg_struct.GameOpenAnswerBacx {

	openan := msg_struct.GameOpenAnswerBacx{}

	openan.Winner = msg_struct.BacxWinner{0, 0, 0}

	openan.BacxAnswerCard.PlayerCard[0].Used = 1
	openan.BacxAnswerCard.PlayerCard[0].Number = 10
	openan.BacxAnswerCard.PlayerCard[0].Color = 3

	openan.BacxAnswerCard.PlayerCard[1].Used = 1
	openan.BacxAnswerCard.PlayerCard[1].Number = 7
	openan.BacxAnswerCard.PlayerCard[1].Color = 3

	openan.BacxAnswerCard.BankerCard[0].Used = 1
	openan.BacxAnswerCard.BankerCard[0].Number = 1
	openan.BacxAnswerCard.BankerCard[0].Color = 4

	openan.BacxAnswerCard.BankerCard[1].Used = 1
	openan.BacxAnswerCard.BankerCard[1].Number = 2
	openan.BacxAnswerCard.BankerCard[1].Color = 2

	return &openan
}

func InitBacxPairB() *msg_struct.GameOpenAnswerBacx {

	openan := msg_struct.GameOpenAnswerBacx{}

	openan.Winner = msg_struct.BacxWinner{1, 0, 0}

	openan.BacxAnswerCard.PlayerCard[0].Used = 1
	openan.BacxAnswerCard.PlayerCard[0].Number = 10
	openan.BacxAnswerCard.PlayerCard[0].Color = 3

	openan.BacxAnswerCard.PlayerCard[1].Used = 1
	openan.BacxAnswerCard.PlayerCard[1].Number = 2
	openan.BacxAnswerCard.PlayerCard[1].Color = 2

	openan.BacxAnswerCard.PlayerCard[2].Used = 1
	openan.BacxAnswerCard.PlayerCard[2].Number = 3
	openan.BacxAnswerCard.PlayerCard[2].Color = 3

	openan.BacxAnswerCard.BankerCard[0].Used = 1
	openan.BacxAnswerCard.BankerCard[0].Number = 4
	openan.BacxAnswerCard.BankerCard[0].Color = 4

	openan.BacxAnswerCard.BankerCard[1].Used = 1
	openan.BacxAnswerCard.BankerCard[1].Number = 1
	openan.BacxAnswerCard.BankerCard[1].Color = 2

	openan.BacxAnswerCard.BankerCard[2].Used = 1
	openan.BacxAnswerCard.BankerCard[2].Number = 4
	openan.BacxAnswerCard.BankerCard[2].Color = 3

	return &openan
}

func InitBacxPairC() *msg_struct.GameOpenAnswerBacx {

	openan := msg_struct.GameOpenAnswerBacx{}

	openan.Winner = msg_struct.BacxWinner{2, 0, 0}

	openan.BacxAnswerCard.PlayerCard[0].Used = 1
	openan.BacxAnswerCard.PlayerCard[0].Number = 5
	openan.BacxAnswerCard.PlayerCard[0].Color = 3

	openan.BacxAnswerCard.PlayerCard[1].Used = 1
	openan.BacxAnswerCard.PlayerCard[1].Number = 2
	openan.BacxAnswerCard.PlayerCard[1].Color = 3

	openan.BacxAnswerCard.BankerCard[0].Used = 1
	openan.BacxAnswerCard.BankerCard[0].Number = 3
	openan.BacxAnswerCard.BankerCard[0].Color = 4

	openan.BacxAnswerCard.BankerCard[1].Used = 1
	openan.BacxAnswerCard.BankerCard[1].Number = 4
	openan.BacxAnswerCard.BankerCard[1].Color = 2

	return &openan
}

func InitAnswersMapBacx() {

	AnswersMapBacx = make([]*msg_struct.GameOpenAnswerBacx, 0)

	//开奖数组1
	AnswersMapBacx = append(AnswersMapBacx, InitBacxPairA())
	AnswersMapBacx = append(AnswersMapBacx, InitBacxPairB())
	AnswersMapBacx = append(AnswersMapBacx, InitBacxPairC())

}
