package redisId

/****************************************************
此文件是用于定义同redis打交道的各类定义
*****************************************************/

const (
	RedisId_PlayerWallet  = "Or_PlayerWallet:"  //玩家的钱包
	RedisId_PlayerAccount = "Or_PlayerAccount:" //玩家的帐目

	RedisHashId_PlayersKey = "Or_PlayersInfo:"

	RedisHashId_MerchantKey = "Or_MerchantNode:" //商户的信息

	RedisId_GameInfo    = "Or_GameInfo:" //游戏的当前信息
	RedisId_GameAnswers = "Or_GameAnswers:"
)
