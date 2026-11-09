package player

type PoolObj struct {
	ref bool
}

// DataInfo 用于存于不需要持久化的数据，需要序列化的数据首字母大写
type DataInfo struct {
	FromGateId string //关联网关nodeId
	ClientId   string //clientId
	isOnline   bool   //是否在线状态

	Currency string //币货

	Balance uint64 //余额

	Digits uint //余额的小数位

	//Balance decimal.Decimal //玩家的余额
}

func (po *PoolObj) IsRef() bool {
	return po.ref
}

func (po *PoolObj) Ref() {
	po.ref = true
}

func (po *PoolObj) UnRef() {
	po.ref = false
}

func (po *PoolObj) Reset() {
	po.ref = false
}

func (df *DataInfo) GetClientId() string {
	return df.ClientId
}

func (df *DataInfo) GetGateNodeId() string {
	return df.FromGateId
}

func (df *DataInfo) SetOnline(onLine bool) {
	df.isOnline = onLine
}

func (df *DataInfo) GetOnline() bool {
	return df.isOnline
}

func (df *DataInfo) AttachConn(fromGateNodeId string, clientId string, sessionId string) {

	df.FromGateId = fromGateNodeId
	df.ClientId = clientId

	//ipStr := strings.Split(ip, ":")
	//if len(ipStr) > 0 {
	//	ip = ipStr[0]
	//}
}
