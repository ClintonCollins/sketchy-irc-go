package sketchyircgo

import (
	"fmt"
	"time"
	"net"
)

func (Instance *IRCInstance) send(message string) {
	Instance.Conn.Write([]byte(message + "\r\n"))
	fmt.Println("[" + time.Now().Format("2006/01/02 15:04:05") + "] -> " + message)
}

// Wrapper to easily connect to an IRC server
func (Instance *IRCInstance) connect(Address, Username, OAuth string) {
	raddr, err := net.ResolveTCPAddr("tcp", Address)
	if err != nil {
		panic(err)
	}
	s, err := net.DialTCP("tcp", nil, raddr)
	Instance.SafetyLock.Lock()
	Instance.Conn = s
	Instance.SafetyLock.Unlock()
	if err != nil {
		Instance.SafetyLock.Lock()
		Instance.Connected = false
		Instance.SafetyLock.Unlock()
		panic(err)
	} else {
		Instance.SafetyLock.Lock()
		Instance.Connected = true
		Instance.SafetyLock.Unlock()
		Instance.send("PASS " + OAuth)
		Instance.send("NICK " + Username)
		Instance.send("USER " + Username + Username + Address + " :" + Username)
	}
}

// Gracefully exit server and throw a stack trace when closing
func (Instance *IRCInstance) handle(s *net.TCPConn, e error) {
	Instance.send("QUIT :Segfault >:O")
	panic(e)
}

// Wrapper to write a message to the bot's log
func writeLog(s string) {
	fmt.Println("[" + time.Now().Format("2006/01/02 15:04:05") + "] <- " + s)
}

func connWatchdog(Instance *IRCInstance) {
	Instance.SafetyLock.Lock()
	Instance.LastActive = time.Now()
	Instance.SafetyLock.Unlock()
	for {
		time.Sleep(1 * time.Second)
		Instance.SafetyLock.RLock()
		timeSinceActive := time.Since(Instance.LastActive)
		Instance.SafetyLock.RUnlock()
		if timeSinceActive > 300*time.Second {
			Instance.SafetyLock.Lock()
			Instance.Conn.Close()
			Instance.SafetyLock.Unlock()
			return
		}
	}
}