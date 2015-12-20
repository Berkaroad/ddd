package cqrs

import (
	"github.com/berkaroad/ddd/domain"
	"github.com/berkaroad/util"
	"reflect"
	"sync"
)

// Command 接口
type Command interface {
	AggregateId() domain.AggregateIdentity
}

// CommandType
type CommandType reflect.Type

// 获取 Command 的 CommandType
func CommandTypeOf(c Command) CommandType {
	return CommandType(util.FromPtrTypeOf(c))
}

// 一个 CommandHandler 对应一组 CommandType
// 单实例，非线程安全
type CommandHandler interface {
	// 支持的CommandType列表
	CommandTypes() []CommandType
	// 执行 Command
	Execute(c Command)
}

// CommandMiddleware 接口
// 临时对象，线程安全
type CommandMiddleware interface {
	// 执行 Command 前切入
	Before(c Command, ctx ExecuteBeforeContext)
	// 执行 Command 后切入
	After(c Command, ctx ExecuteAfterContext)
}

// ExecuteBeforeContext 接口
// 用于中间件
type ExecuteBeforeContext interface {
	// Command 的类型名
	CommandName() string
	// CommandHandler 名字
	HandlerName() string
	// 停止即将执行的 Command
	StopBy(middleware CommandMiddleware)
}

// ExecuteAfterContext 接口
// 用于中间件
type ExecuteAfterContext interface {
	// Command 的类型名
	CommandName() string
	// CommandHandler 名字
	HandlerName() string
	// 是否 Command 执行成功
	IsSuccess() bool
}

// Command 消息队列适配器
type CommandMQAdapter interface {
	util.Initializer
	ReceiveCommand(receiveFunc func(commandStream chan Command, wg *sync.WaitGroup)) error
	Close()
}

// 用于发送Command的服务
type CommandService interface {
	SendCommand(c Command)
}
