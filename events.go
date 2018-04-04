package sketchyircgo

import "sync"

var (
	messageEventHandlers        []func(*IRCInstance, *Message)
	messageEventHandlersLock    sync.RWMutex
	serverJoinEventHandlers     []func(*IRCInstance)
	serverJoinEventHandlersLock sync.RWMutex
	modeChangeEventHandlers     []func(*IRCInstance, *ModeChange)
	modeChangeEventHandlersLock sync.RWMutex
	userJoinEventHandlers       []func(*IRCInstance, *UserJoin)
	userJoinEventHandlersLock   sync.RWMutex
	userPartEventHandlers       []func(*IRCInstance, *UserPart)
	userPartEventHandlersLock   sync.RWMutex
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

func (Instance *IRCInstance) NewUserJoinListener(function func(*IRCInstance, *UserJoin)) {
	userJoinEventHandlersLock.Lock()
	userJoinEventHandlers = append(userJoinEventHandlers, function)
	userJoinEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) sendUserJoinListener(instance *IRCInstance, userJoin *UserJoin) {
	userJoinEventHandlersLock.Lock()
	for _, listener := range userJoinEventHandlers {
		listener(instance, userJoin)
	}
	userJoinEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) NewUserPartListener(function func(*IRCInstance, *UserPart)) {
	userPartEventHandlersLock.Lock()
	userPartEventHandlers = append(userPartEventHandlers, function)
	userPartEventHandlersLock.Unlock()
}

func (Instance *IRCInstance) sendUserJPartListener(instance *IRCInstance, userPart *UserPart) {
	userPartEventHandlersLock.Lock()
	for _, listener := range userPartEventHandlers {
		listener(instance, userPart)
	}
	userPartEventHandlersLock.Unlock()
}
