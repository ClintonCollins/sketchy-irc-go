package sketchyircgo

import (
	"net"
	"sync"
	"time"
)

type IRCInstance struct {
	address    string
	username   string
	password   string
	connected  bool
	conn       *net.TCPConn
	lastActive time.Time
	twitchIRC  bool
	channels map[string]*Channel
	closeChannel chan bool
	sync.RWMutex
}

type Channel struct {
	name       string
	moderators map[string]*User
	users      map[string]*User
	sync.RWMutex
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
