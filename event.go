package sketchyircgo

import "sync"

var (
	messageEventHandlers        []func(*IRCInstance, *Message)
	messageEventHandlersLock    sync.RWMutex
	serverJoinEventHandlers     []func(*IRCInstance)
	ServerJoinEventHandlersLock sync.RWMutex
)

func (Instance *IRCInstance) NewMessageListener(function func(*IRCInstance, *Message)) {
	messageEventHandlersLock.Lock()
	messageEventHandlers = append(messageEventHandlers, function)
	messageEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) sendMessageListener(instance *IRCInstance, message *Message) {
	messageEventHandlersLock.Lock()
	for _, listener := range messageEventHandlers {
		listener(instance, message)
	}
	messageEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) NewServerReadyListener(function func(*IRCInstance)) {
	ServerJoinEventHandlersLock.Lock()
	serverJoinEventHandlers = append(serverJoinEventHandlers, function)
	ServerJoinEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) sendServerReadyListener(instance *IRCInstance) {
	ServerJoinEventHandlersLock.Lock()
	for _, listener := range serverJoinEventHandlers {
		listener(instance)
	}
	ServerJoinEventHandlersLock.Unlock()
}