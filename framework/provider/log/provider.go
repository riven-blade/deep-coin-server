package log

import (
	framework2 "devops-http/framework"
	"devops-http/framework/contract"
	"devops-http/framework/provider/log/services"
)

// NiceLogServiceProvider 服务提供者
type NiceLogServiceProvider struct {
	framework2.ServiceProvider
}

// Register 注册一个服务实例
func (l *NiceLogServiceProvider) Register(c framework2.Container) framework2.NewInstance {
	return services.NewNiceLog
}

// Boot 启动的时候注入
func (l *NiceLogServiceProvider) Boot(c framework2.Container) error {
	return nil
}

// IsDefer 是否延迟加载
func (l *NiceLogServiceProvider) IsDefer() bool {
	return false
}

// Params 定义要传递给实例化方法的参数
func (l *NiceLogServiceProvider) Params(c framework2.Container) []interface{} {
	return []interface{}{c}
}

// Name 定义对应的服务字符串凭证
func (l *NiceLogServiceProvider) Name() string {
	return contract.LogKey
}
