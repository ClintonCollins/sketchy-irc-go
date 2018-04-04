package sketchyircgo

import (
	"net"
	"sync"
	"time"
)

type IRCInstance struct {
	Address      string
	Username     string
	Password     string
	Connected    bool
	Conn         *net.TCPConn
	LastActive   time.Time
	TwitchIRC    bool
	ChannelsLock sync.RWMutex
	Channels     map[string]*Channel
	SafetyLock   sync.RWMutex
	CloseChannel chan bool
}

type Channel struct {
	Name       string
	Moderators map[string]*User
	Users      map[string]*User
}

type Message struct {
	Channel *Channel
	Message string
	Time    time.Time
	Author  *User
	Type    string
}

type User struct {
	Name      string
	Moderator bool
	// For Twitch Only Below
	Subscriber  bool
	Broadcaster bool
	Turbo       bool
	DisplayName string
	GlobalMod   bool
	Staff       bool
	Admin       bool
}

type ModeChange struct {
	Channel  *Channel
	Message  string
	Time     time.Time
	Sender   string
	Receiver string
	Mode     string
}

type UserJoin struct {
	Channel *Channel
	Time    time.Time
	User    *User
}

type UserPart struct {
	Channel *Channel
	Time    time.Time
	User    *User
}
