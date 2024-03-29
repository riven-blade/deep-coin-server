package gin

import (
	framework2 "devops-http/framework"
)

// SetContainer 为Engine设置container
func (engine *Engine) SetContainer(container framework2.Container) {
	engine.container = container
}

// GetContainer 从Engine中获取container
func (engine *Engine) GetContainer() framework2.Container {
	return engine.container
}

// Bind engine实现container的绑定封装
func (engine *Engine) Bind(provider framework2.ServiceProvider) error {
	return engine.container.Bind(provider)
}

// IsBind 关键字凭证是否已经绑定服务提供者
func (engine *Engine) IsBind(key string) bool {
	return engine.container.IsBind(key)
}
