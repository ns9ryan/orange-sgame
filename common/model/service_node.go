package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

type ServiceNode struct {
	ID              uint                  `gorm:"primarykey;type:int unsigned;comment:ID"`
	Ip              string                `gorm:"size:50;comment:IP地址"`
	Port            uint                  `gorm:"type:int unsigned;not null;comment:端口"`
	OnlineCount     uint                  `gorm:"type:int unsigned;not null;default:0;comment:当前在线玩家数（针对连接数）"`
	Disabled        uint                  `gorm:"type:tinyint unsigned;not null;default:0;comment:是否禁用:0否,1是"`
	Note            string                `gorm:"not null;default:'';comment:备注"`
	ServiceKey      NodeServiceKey        `gorm:"size:64;not null;default:'login_ws';comment:服务标识，例如 login_ws 或 wallet_tcp 或 chat_ws等"`
	ServiceConnType ServiceConnectionType `gorm:"size:10;not null;default:'ws';comment:服务类型，ws=WebSocket，tcp=TCP"`
	BaseModel
}

func (ServiceNode) TableName() string {
	return "or_service_node"
}

func GetServiceNodeList() ([]*ServiceNode, error) {
	var serviceNode []*ServiceNode

	result := variable.GormDbMysql.Clauses(dbresolver.Read).
		Select("ip,port,note").
		Find(&serviceNode)

	return serviceNode, result.Error
}

func GetServiceNodeByNodeId(nodeId string) (*ServiceNode, error) {
	var serviceNode ServiceNode

	result := variable.GormDbMysql.Clauses(dbresolver.Read).
		Where("note = ? AND disabled = ?", nodeId, 0).
		Select("ip,port").
		First(&serviceNode)

	return &serviceNode, result.Error
}
