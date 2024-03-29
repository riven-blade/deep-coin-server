package role

import (
	"devops-http/app/module/base"
)

// DevopsSysRole 角色表
type DevopsSysRole struct {
	base.DevopsModel
	ParentId string `gorm:"column:parent_id;type:int;default:0;" json:"parentId"`                 // 角色
	Name     string `gorm:"unique;column:name;type:varchar(200);not null;default:''" json:"name"` // 名称
	Status   int    `gorm:"column:status;type:int;default:null;default:1" json:"status"`          // 状态//radio/2,隐藏,1,显示
	Sort     int    `gorm:"column:sort;type:int;default:null;default:1" json:"sort"`              // 排序
	Remark   string `gorm:"column:remark;type:text;default:null" json:"remark"`                   // 说明//textarea
	Enable   int    `gorm:"column:enable;type:tinyint(1);default:null;default:1" json:"enable"`   // 是否启用//radio/1,启用,2,禁用
	Domain   int    `gorm:"column:domain;type:int;default:0;" json:"domain"`                      // 域
}
