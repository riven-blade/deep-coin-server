package host

import (
	"devops-http/app/module/base/request"
	"devops-http/app/module/base/response"
	"devops-http/app/module/sys/service/host"
	"devops-http/framework"
	"devops-http/framework/gin"
)

type ApiHost struct {
	service *host.Service
}

func Register(r *gin.Engine) error {
	api := NewHostApi(r.GetContainer())
	cmdbGroup := r.Group("/cmdb/", func(c *gin.Context) {
	})
	cmdbGroup.POST("host/list", api.GetHostList)
	cmdbGroup.GET("host/group/tree", api.GetHostGroupTree)
	return nil
}

func NewHostApi(c framework.Container) *ApiHost {
	return &ApiHost{service: host.NewService(c)}
}

// GetHostGroupTree godoc
// @Summary 获得主机分组树接口
// @Security ApiKeyAuth
// @Description 获得主机分组树接口
// @accept application/json
// @Produce application/json
// @Tags Menu
// @Success 200 {object}  response.Response
// @Router /cmdb/host/group/tree [get]
func (a *ApiHost) GetHostGroupTree(c *gin.Context) {
	res := response.Response{Code: 1, Msg: "查询成功", Data: nil}
	result, err := a.service.GetHostGroupTree()
	if err != nil {
		res.Msg = err.Error()
		c.DJson(res)
		return
	}
	res.Data = map[string]interface{}{"list": result}
	c.DJson(res)
}

// GetHostList godoc
// @Summary 获得主机列表接口
// @Security ApiKeyAuth
// @Description 获得主机列表接口
// @accept application/json
// @Produce application/json
// @Param data body request.SearchHostParams true "页数，页大小，筛选条件"
// @Tags Menu
// @Success 200 {object}  response.Response
// @Router /cmdb/host/list [post]
func (a *ApiHost) GetHostList(c *gin.Context) {
	var param request.SearchHostParams
	err := c.ShouldBindJSON(&param)
	res := response.Response{Code: 1, Msg: "查询成功", Data: nil}
	if err != nil {
		res.Msg = err.Error()
		c.DJson(res)
		return
	}
	result, err := a.service.GetHostList(param)
	if err != nil {
		res.Msg = err.Error()
		c.DJson(res)
		return
	}
	res.Data = &result
	c.DJson(res)
}
