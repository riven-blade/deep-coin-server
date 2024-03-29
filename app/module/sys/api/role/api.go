package role

import (
	"devops-http/app/module/sys/service/role"
	"devops-http/framework"
	"devops-http/framework/gin"
)

type ApiRole struct {
	service *role.Service
}

func Register(r *gin.Engine) error {
	api := NewSysApi(r.GetContainer())
	sysGroup := r.Group("/sys/", func(c *gin.Context) {
	})

	// 用户角色相关接口
	sysGroup.GET("roles/:id", api.GetRole)
	sysGroup.POST("roles/list", api.ListRoles)
	sysGroup.GET("roles/tree", api.TreeRoles)
	sysGroup.POST("roles/add", api.AddRole)
	sysGroup.PUT("roles/modify", api.ModifyRole)
	sysGroup.POST("roles/copy", api.CopyRole)
	sysGroup.DELETE("roles/delete", api.DeleteRole)

	return nil
}

func NewSysApi(c framework.Container) *ApiRole {
	return &ApiRole{service: role.NewService(c)}
}
