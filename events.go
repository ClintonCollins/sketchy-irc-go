package sketchyircgo

import "sync"

var (
	messageEventHandlers        []func(*IRCInstance, *Message)
	messageEventHandlersLock    sync.RWMutex
	serverJoinEventHandlers     []func(*IRCInstance)
	serverJoinEventHandlersLock sync.RWMutex
	modeChangeEventHandlers     []func(*IRCInstance, *ModeChange)
	modeChangeEventHandlersLock sync.RWMutex
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

func (Instance *IRCInstance) NewModeChangeListener(function func(*IRCInstance, *ModeChange)) {
	modeChangeEventHandlersLock.Lock()
	modeChangeEventHandlers = append(modeChangeEventHandlers, function)
	modeChangeEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) sendModeChangeListener(instance *IRCInstance, userMode *ModeChange) {
	modeChangeEventHandlersLock.Lock()
	for _, listener := range modeChangeEventHandlers {
		listener(instance, userMode)
	}
	modeChangeEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) NewServerReadyListener(function func(*IRCInstance)) {
	serverJoinEventHandlersLock.Lock()
	serverJoinEventHandlers = append(serverJoinEventHandlers, function)
	serverJoinEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) sendServerReadyListener(instance *IRCInstance) {
	serverJoinEventHandlersLock.Lock()
	for _, listener := range serverJoinEventHandlers {
		listener(instance)
	}
	serverJoinEventHandlersLock.Unlock()
}
