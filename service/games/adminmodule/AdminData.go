package adminmodule

import (
	"strings"
)

type PoolObj struct {
	ref bool
}

// DataInfo 用于存于不需要持久化的数据，需要序列化的数据首字母大写
type DataInfo struct {
	ClientId string //clientId

	IpAddr string

	isOnline bool //是否在线状态

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

func (df *DataInfo) SetOnline(onLine bool) {
	df.isOnline = onLine
}

func (df *DataInfo) AttachConn(clientId string, ip string, os string) {

	ipStr := strings.Split(ip, ":")

	if len(ipStr) > 0 {
		ip = ipStr[0]
	}

	df.ClientId = clientId

	if clientId == "" {
		df.isOnline = false
	}

}
