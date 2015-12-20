package cqrs

import (
	"github.com/berkaroad/ddd"
	// "github.com/berkaroad/ddd/infrastructure/mq"
	"github.com/berkaroad/util"
	"github.com/berkaroad/uuid"
	"reflect"
	"sync"
	"time"
)

type commandHandlerContext struct {
	commandName      string
	handlerName      string
	stopByMiddleware CommandMiddleware
	isStop           bool
	isSuccess        bool
}

func (self *commandHandlerContext) CommandName() string {
	return self.commandName
}

func (self *commandHandlerContext) HandlerName() string {
	return self.handlerName
}

func (self *commandHandlerContext) StopBy(middleware CommandMiddleware) {
	self.isStop = true
	self.stopByMiddleware = middleware
}

func (self *commandHandlerContext) IsSuccess() bool {
	return self.isSuccess
}

// CommandProcessor
// 用于注册 CommandHandler ，然后执行命令
type CommandProcessor struct {
	adapter         CommandMQAdapter
	handlers        map[string]CommandHandler
	middlewareTypes []reflect.Type
}

// 创建一个CommandProcessor
func NewCommandProcessor(adapter CommandMQAdapter) *CommandProcessor {
	processor := CommandProcessor{adapter: adapter, handlers: make(map[string]CommandHandler, 0)}
	return &processor
}

// 注册 CommandHandler
func (self *CommandProcessor) RegisterHandler(handler CommandHandler) {
	handlerName := util.FromPtrTypeOf(handler).String()
	for _, commandType := range handler.CommandTypes() {
		commandName := util.FromPtrType(commandType).Name()
		self.handlers[commandName] = handler
		consoleLog.Printf("Register handler:%s for %s.\n", handlerName, commandName)
	}
}

// 注册 CommandMiddleware
func (self *CommandProcessor) RegisterMiddleware(middleware CommandMiddleware) {
	middlewareType := util.FromPtrTypeOf(middleware)
	middlewareName := middlewareType.String()
	self.middlewareTypes = append(self.middlewareTypes, middlewareType)
	consoleLog.Printf("Register middleware:%s.\n", middlewareName)
}

// 创建已注册的中间件实例
func (self *CommandProcessor) createMiddlewares() []CommandMiddleware {
	middlewares := make([]CommandMiddleware, len(self.middlewareTypes))
	for i := 0; i < len(middlewares); i++ {
		if middleware, ok := reflect.New(self.middlewareTypes[i]).Interface().(CommandMiddleware); ok {
			middlewares[i] = middleware
		}
	}
	return middlewares
}

// 在 Command 执行前，先执行中间件
func (self *CommandProcessor) handleBefore(middlewares []CommandMiddleware, command Command, context ExecuteBeforeContext) {
	for _, middleware := range middlewares {
		if middleware != nil {
			middleware.Before(command, context)
		}
	}
}

// 在 Command 执行后，执行中间件
func (self *CommandProcessor) handleAfter(middlewares []CommandMiddleware, command Command, context ExecuteAfterContext) {
	for _, middleware := range middlewares {
		if middleware != nil {
			middleware.After(command, context)
		}
	}
}

// 执行单个 Command
func (self *CommandProcessor) executeCommand(command Command, commandId uuid.UUID) {
	commandName := util.FromPtrTypeOf(command).Name()
	var handler CommandHandler
	var handlerName string

	for key, val := range self.handlers {
		if key == commandName {
			handler = val
			handlerName = util.FromPtrTypeOf(val).String()
			break
		}
	}

	if handler != nil {
		context := &commandHandlerContext{commandName: commandName, handlerName: handlerName, isStop: false, isSuccess: false}
		middlewares := self.createMiddlewares()
		self.handleBefore(middlewares, command, ExecuteBeforeContext(context))
		if !context.isStop {
			startTime := time.Now().UnixNano()
			handler.Execute(command)
			endTime := time.Now().UnixNano()
			if ddd.DEBUG {
				consoleLog.Printf("Command '%s_%s' execute in %vms.\n", commandName, commandId, float64(endTime-startTime)/float64(time.Millisecond))
			}
			context.isSuccess = true
			self.handleAfter(middlewares, command, ExecuteAfterContext(context))
		} else {
			consoleLog.Printf("Command '%s_%s' is stopped by %s.\n", commandName, commandId, util.FromPtrTypeOf(context.stopByMiddleware).String())
		}
	} else {
		consoleLog.Panicf("[panic] CommandHandler for Command %s not found!", commandName)
	}
}

// 运行
// [command]
// mq_addr=127.0.0.1:5001
// mq_queue_name=command_queue
// concurrent_num=2
func (self *CommandProcessor) Run() {
	// mq_addr := config.Get("command", "mq_addr")
	// mq_queue_name := config.Get("command", "mq_queue_name")
	concurrent_num := config.GetInt("command", "concurrent_num")
	if concurrent_num <= 0 {
		concurrent_num = 1
	}
	defer self.adapter.Close()

	self.adapter.ReceiveCommand(func(commandStream chan Command, wg *sync.WaitGroup) {
		locker := make(chan bool, 4)
		defer close(locker)

		for command := range commandStream {
			if command == nil {
				continue
			}
			cmdChan := make(chan Command)
			defer close(cmdChan)
			go func(cmdChan chan Command, locker chan bool) {
				command := <-cmdChan
				self.executeCommand(command, uuid.New())
				<-locker
				wg.Done()
			}(cmdChan, locker)
			cmdChan <- command
			locker <- true
		}
	})
	// if err := self.mqProvider.Dial(mq_addr); err == nil {
	// 	defer self.mqProvider.Close()
	// 	if subscri, err := self.mqProvider.Subscribe(mq_queue_name); err == nil {
	// subscri.Receive(func(messageStream chan mq.Message, wg *sync.WaitGroup) {
	// 	locker := make(chan bool, 4)
	// 	defer close(locker)

	// 	for message := range messageStream {
	// 		if message == nil || message.GetBody() == nil {
	// 			continue
	// 		}
	// 		msgchan := make(chan mq.Message)
	// 		defer close(msgchan)
	// 		go func(msgchan chan mq.Message, locker chan bool) {
	// 			msg := <-msgchan

	// 			if command, ok := util.FromPtrValueOf(msg.GetBody()).Interface().(Command); ok {
	// 				self.executeCommand(command, msg.GetId())
	// 			}
	// 			<-locker
	// 			wg.Done()
	// 		}(msgchan, locker)
	// 		msgchan <- message
	// 		locker <- true
	// 	}
	// })
	// }
	// }

	suspend := make(chan bool)
	<-suspend
}
