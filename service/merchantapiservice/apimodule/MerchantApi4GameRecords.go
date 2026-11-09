//*****************************************************************
// OBLiveApi-4-Game.go里，放的是API文档里描述的数据接口(目前一共6个)，包括：
// 查询下注历史、对账接口(按代理)、对账接口(按会员)等供商户调用的API接口。
//*****************************************************************

package apimodule

import (
	"encoding/json"
	"strings"

	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	"time"
	models "videogame/common/model"
	"strconv"
	"videogame/common/msg_struct"
	"videogame/common/util"
)

// //龙虎的Winner编码
var (

	betAreas map[string]string

	dtxWinnerMap = map[int]string{
		0:  	"Dragon",
		1:   	"Tiger",
		2:     	"Tie",
		3: 		"TiePair",
	}

	bacxWinnerMap = map[int]string{
		0:      "Player",
		1:      "Banker",
		2:      "Tie",
		3:      "PlayerPair",
		4:      "BankerPair",
		5:		"BankerNoComm",
		6:    	"SuperSix",
		7:     	"AnyPair",
		8: 		"PerfectPair",
		9:      "Big",
		10:     "Small",
		11: 	"PlayerBonus",
		12: 	"BankerBonus",
	}

)

func init() {
	if betAreas == nil {
		betAreas = make(map[string]string)
		currs, err := models.GetBetAreas()
		if err != nil {
			fmt.Printf("获取下注编码列表失败")
			return
		}else{
			for _, area := range currs {
				betAreas[area.AreaId] = area.AreaKey
			}		
		}

		fmt.Printf("GetBetAreas, betAreas=%v\n", betAreas)
	}
}

type AnswerDTX struct {
	DCard msg_struct.Card `json:"dcard"`
	TCard msg_struct.Card `json:"tcard"`
}

func convertCard(card msg_struct.Card) string {
	suitMap := map[byte]string{
		1: "S", // 黑桃 Spades
		2: "H", // 红心 Hearts
		3: "C", // 梅花 Clubs
		4: "D", // 方块 Diamonds
	}
	valueMap := map[byte]string{
		1:  "A", 2: "2", 3: "3", 4: "4", 5: "5",
		6:  "6", 7: "7", 8: "8", 9: "9", 10: "T",
		11: "J", 12: "Q", 13: "K",
	}

	valStr := valueMap[card.Number]
	suitStr := suitMap[card.Color]

	return valStr + suitStr
}

func ConvertAnswerDTX(raw json.RawMessage) (map[string]string, error) {
	var dto AnswerDTX
	err := json.Unmarshal(raw, &dto)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"dcard": convertCard(dto.DCard),
		"tcard": convertCard(dto.TCard),
	}, nil

}

func ConvertAnswerBACX(raw json.RawMessage, ppair int, bpair int) (map[string]interface{}, error) {
	var dto msg_struct.BacxAnswerCard

	err := json.Unmarshal(raw, &dto)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]interface{})

	//付加的对子检测
	ret["is_ppair"] = uint32(ppair)
	ret["is_bpair"] = uint32(bpair)
	ret["is_any_pair"] = uint32(util.ConverBoolToInt(dto.IsAnyPair()))
	ret["is_big"] = uint32(util.ConverBoolToInt(dto.IsBig()))
	ret["is_perfect_pair"] = uint32(util.ConverBoolToInt(dto.IsPerfectPair()))
	ok := dto.IsBankerBonus()
	ret["is_bank_bonus"] = uint32(util.ConverBoolToInt(ok))
	ok = dto.IsPlayerBonus()
	ret["is_player_bonus"] = uint32(util.ConverBoolToInt(ok))

	convertCards := func(cards [3]msg_struct.Card) []string {
		var result []string
		for _, card := range cards {
			if card.Used == 0 {
				continue
			}
			result = append(result, convertCard(card))
		}
		return result
	}

	ret["bcard"] = convertCards(dto.PlayerCard)
	ret["pcard"] = convertCards(dto.BankerCard)

	return ret, nil

}

type BetRecord struct {
	RecordId		string			`json:"record_id"`
	GameTypeId		string			`json:"game_type_id"`
	GameTableId		int				`json:"game_table_id"`
	GameRoundId		string			`json:"game_round_id"`
	Username		string			`json:"username"`
	Code			string			`json:"currency"`
	BetTime			string			`json:"bet_time"`
	BetAreaCode		string			`json:"bet_area_code"`
	BetAmount		FixedOutput 	`json:"bet_amount"`
	ValidBetAmount 	FixedOutput		`json:"valid_bet_amount"`
	NetAmount		FixedOutput		`json:"net_amount"`
	Winner			interface{}		`json:"winner"`
	Value1			int				`json:"-"`
	Value2			int				`json:"-"`
	Value3			int				`json:"-"`
	AnswerTime		string			`json:"answer_time"`
	Answer			interface{}		`json:"answer"`
}

type DTXAnswer struct {
	Dcard			string			`json:"dcard"`
	Tcard			string			`json:"tcard"`
}

type BACXAnswer struct {
	Dcard			[]string			`json:"dcard"`
	Tcard			[]string			`json:"tcard"`
	IsPpair			int					`json:"is_ppair"`
	IsBpair			int					`json:"is_bpair"`
	IsSupersix		int					`json:"is_supersix"`
	IsAnyPair		int					`json:"is_any_pair"`
	IsPerfectPair	int					`json:"is_perfect_pair"`
}

// 游戏下注明细查询，定义请求结构体
type MBetRecordsApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	StartTime 		string  `json:"start_time"`
	EndTime 		string  `json:"end_time"`
	PageIndex		int		`json:"page_index"`

	GameTypeId 		string  `json:"game_type_id"`
	GameTableId		int		`json:"game_table_id"`
	GameRoundId		string	`json:"game_round_id"`
	Username		string  `json:"username"`
}

// 游戏下注明细查询，定义返回结构体
type MBetRecordsApiRet struct {
	PageIndex		int		    `json:"page_index"`
	PageSize		int		    `json:"page_size"`
	TotalRecord		int64		`json:"total_record"`
	TotalPage		int64		`json:"total_page"`

	Records			[]BetRecord 	`json:"records"`
}

// 下注明细查询
func (h *MerchantApiModule) GameRecordsBet(c *gin.Context) {
	var req MBetRecordsApiReq
	var ret MBetRecordsApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	_, err := models.GetMerchantByCode(req.MerchantCode)
	if err != nil {
		ReturnHttpApi(c, 10001, "没有找到商户信息", "")
		return
	}

	username := strings.ToLower(fmt.Sprintf("%s_%s", req.MerchantCode, req.Username))
	if req.Username == "" {
		username = ""
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		ReturnHttpApi(c, 10001, "解析开始时间失败", "")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		ReturnHttpApi(c, 10001, "解析开始时间失败", "")
		return
	}

	betRecords, err := models.GetBetRecords(req.MerchantCode, username, req.GameTypeId, req.GameTableId, req.GameRoundId, startTime, endTime, req.PageIndex)
	if err != nil {
		ReturnHttpApi(c, 10001, "商户订单号不存在", "")
		return
	}

	var records = make([]BetRecord, 0)
	for _, br := range betRecords {
		var record BetRecord

		record.RecordId = br.RecordId
		record.GameTypeId = br.Type
		record.GameTableId = br.DeskNo
		record.GameRoundId = br.RoundId
		record.Username = br.UserId
		record.Code = br.Currency

		var betTimeStr string
		if br.BetTime != nil {
		    betTimeStr = br.BetTime.Format("2006-01-02T15:04:05-07:00")
		} else {
		    betTimeStr = "" // 处理 nil 值，避免 panic
		}
		record.BetTime = betTimeStr

		record.BetAreaCode = strconv.Itoa(br.BetAreaId)
		betAreaName, ok := betAreas[record.BetAreaCode]
		if ok {
			record.BetAreaCode = betAreaName
		}

		if br.Digits == 0 {
			br.Digits = 1
		}

		record.BetAmount = WrapFixedFloat(float64(br.BetAmount) / float64(br.Digits), br.Currency)
		record.ValidBetAmount = WrapFixedFloat(float64(br.BetEffective) / float64(br.Digits), br.Currency)
		record.NetAmount = WrapFixedFloat(float64(br.WinLost) / float64(br.Digits), br.Currency)

		record.Winner = br.Winner

		var answerTimeStr string
		if br.BetTime != nil {
		    answerTimeStr = br.AnswerTime.Format("2006-01-02T15:04:05-07:00")
		} else {
		    answerTimeStr = "" // 处理 nil 值，避免 panic
		}

		record.AnswerTime = answerTimeStr
		record.Answer = br.Answer

		switch br.Type {
		case "DTX":
			name, ok := dtxWinnerMap[br.Winner]
			if ok {
				record.Winner = name
			}

			if len(br.Answer) > 0 {
				converted, err := ConvertAnswerDTX(br.Answer)
				if err != nil {
					log.Error("DTX Answer 解析失败", log.ErrorField("err", err))
				} else {
					record.Answer = converted
				}
			}

		case "BACX":
			name, ok := bacxWinnerMap[br.Winner]
			if ok {
				record.Winner = name
			}

			if len(br.Answer) > 0 {
				converted, err := ConvertAnswerBACX(br.Answer, br.Value1, br.Value2)
				if err != nil {
					log.Error("DTX Answer 解析失败", log.ErrorField("err", err))
				} else {
					record.Answer = converted
				}
			}

		default:

			if len(br.Answer) > 0{
				record.Answer = br.Answer
			}else{
				record.Answer = br.Winner
			}

		}

		records = append(records, record)
	}

	ret.PageIndex = req.PageIndex
	ret.PageSize = 500

	count, err := models.CountBetRecords(req.MerchantCode, username, req.GameTypeId, req.GameTableId, req.GameRoundId, startTime, endTime,)

	if err != nil {
		ReturnHttpApi(c, 10001, "获取查询总条目失败", "")
		return	
	}

	ret.TotalRecord = count
	ret.TotalPage = (count + 500 - 1) / 500

	ret.Records = records

	//fmt.Printf("GameRecordsBet ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

type TipRecord struct {
	TipId			string			`json:"tip_id"`
	GameTypeId		string			`json:"game_type_id"`
	GameTableId		int				`json:"game_table_id"`
	GameRoundId		string			`json:"game_round_id"`
	TipTime			string			`json:"tip_time"`
	Username		string			`json:"username"`
	Dealername		string			`json:"dealername"`
	Code			string			`json:"currency"`
	Amount			any 			`json:"amount"`
}

// 荷官小费明细查询，定义请求结构体
type MTipRecordsApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	StartTime 		string  `json:"start_time"`
	EndTime 		string  `json:"end_time"`
	PageIndex		int		`json:"page_index"`

	GameTypeId 		string  `json:"game_type_id"`
	GameTableId		int		`json:"game_table_id"`
	GameRoundId		string	`json:"game_round_id"`
	Username		string  `json:"username"`
	Dealername		string  `json:"dealername"`
}

// 游戏下注明细查询，定义返回结构体
type MTipRecordsApiRet struct {
	PageIndex		int		    `json:"page_index"`
	PageSize		int		    `json:"page_size"`
	TotalRecord		int64		`json:"total_record"`
	TotalPage		int64		`json:"total_page"`

	Records			[]TipRecord 	`json:"records"`
}

// 小费明细查询
func (h *MerchantApiModule) GameRecordsTip(c *gin.Context) {
	var req MTipRecordsApiReq
	var ret MTipRecordsApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	_, err := models.GetMerchantByCode(req.MerchantCode)
	if err != nil {
		ReturnHttpApi(c, 10001, "没有找到商户信息", "")
		return
	}

	username := strings.ToLower(fmt.Sprintf("%s_%s", req.MerchantCode, req.Username))
	if req.Username == "" {
		username = ""
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		ReturnHttpApi(c, 10001, "解析开始时间失败", "")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		ReturnHttpApi(c, 10001, "解析开始时间失败", "")
		return
	}

	tipRecords, err := models.GetTipRecords(req.MerchantCode, username, req.Dealername, req.GameTypeId, req.GameTableId, req.GameRoundId, startTime, endTime, req.PageIndex)
	if err != nil {
		ReturnHttpApi(c, 10001, "商户订单号不存在", "")
		return
	}

	var records = make([]TipRecord, 0)
	for _, tr := range tipRecords {
		var record TipRecord

		record.TipId = tr.RecordId
		record.GameTypeId = tr.SubType
		record.GameTableId = tr.DeskNo
		record.GameRoundId = tr.RoundId
		record.Username = tr.UserName
		record.Dealername = tr.DealerId
		record.Code = tr.Currency

		var tipTimeStr string
		if tr.TipTime != nil {
		    tipTimeStr = tr.TipTime.Format("2006-01-02T15:04:05-07:00")
		} else {
		    tipTimeStr = "" // 处理 nil 值，避免 panic
		}
		record.TipTime = tipTimeStr

		if tr.Digits == 0 {
			tr.Digits = 1
		}

		record.Amount = WrapFixedFloat(float64(tr.TipAmount) / float64(tr.Digits), tr.Currency)

		records = append(records, record)
	}

	ret.PageIndex = req.PageIndex
	ret.PageSize = 500

	count, err := models.CountTipRecords(req.MerchantCode, username, req.Dealername, req.GameTypeId, req.GameTableId, req.GameRoundId, startTime, endTime)

	if err != nil {
		ReturnHttpApi(c, 10001, "获取查询总条目失败", "")
		return	
	}

	ret.TotalRecord = count
	ret.TotalPage = (count + 500 - 1) / 500
	ret.Records = records

	fmt.Printf("GameRecordsTip ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

func validateTimeRange(reqStartTime, reqEndTime string) error {
	// 定义时间格式，根据你的时间格式调整
	const timeFormat = "2006-01-02 15:04:05"

	// 解析 StartTime 和 EndTime
	startTime, err := time.Parse(timeFormat, reqStartTime)
	if err != nil {
		return fmt.Errorf("解析 StartTime 失败: %v", err)
	}

	endTime, err := time.Parse(timeFormat, reqEndTime)
	if err != nil {
		return fmt.Errorf("解析 EndTime 失败: %v", err)
	}

	// 检查 EndTime 是否大于当前时间减去 40 秒
	currentTime := time.Now()
	if endTime.After(currentTime.Add(-40 * time.Second)) {
		return fmt.Errorf("EndTime 不能大于当前时间减去 40 秒")
	}

	// 检查 StartTime 和 EndTime 是否在 30 分钟之内
	diff := endTime.Sub(startTime)
	if diff > 30*time.Minute {
		return fmt.Errorf("时间范围超过 30 分钟")
	}

	if diff < 0 {
		return fmt.Errorf("EndTime 早于 StartTime")
	}

	return nil
}
