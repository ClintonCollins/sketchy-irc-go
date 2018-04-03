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
	IRCInstanceLock.Lock()
	Instance.Conn = s
	IRCInstanceLock.Unlock()
	if err != nil {
		IRCInstanceLock.Lock()
		Instance.Connected = false
		IRCInstanceLock.Unlock()
		panic(err)
	} else {
		IRCInstanceLock.Lock()
		Instance.Connected = true
		IRCInstanceLock.Unlock()
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
	IRCInstanceLock.Lock()
	Instance.LastActive = time.Now()
	IRCInstanceLock.Unlock()
	for {
		time.Sleep(1 * time.Second)
		IRCInstanceLock.RLock()
		timeSinceActive := time.Since(Instance.LastActive)
		IRCInstanceLock.RUnlock()
		if timeSinceActive > 300*time.Second {
			IRCInstanceLock.Lock()
			Instance.Conn.Close()
			IRCInstanceLock.Unlock()
			return
		}
	}
}