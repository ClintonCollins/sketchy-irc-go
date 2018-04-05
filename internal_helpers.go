package sketchyircgo

import (
	"errors"
	"fmt"
	"net"
	"time"
)

func (Instance *IRCInstance) send(message string) {
	Instance.Conn.Write([]byte(message + "\r\n"))
	fmt.Println("[" + time.Now().Format("2006/01/02 15:04:05") + "] -> " + message)
}

// Wrapper to easily connect to an IRC server
func (Instance *IRCInstance) connect(Address, Username, Password string, MaxAttempts int) error {
	Instance.Lock()
	if Instance.Connected {
		Instance.Connected = false
		Instance.Conn.Close()
	}
	Instance.Unlock()
	rt := 1
	rc := 0
	raddr, err := net.ResolveTCPAddr("tcp", Address)
	if err != nil {
		return errors.New("failed to resolve remote address")
	}
	for {
		s, err := net.DialTCP("tcp", nil, raddr)
		if err != nil {
			rc++
			if rc > MaxAttempts && MaxAttempts > -1 {
				writeLog("Reconnect attempt limit exceeded")
				return errors.New("reconnect attempt limit exceeded")
			}
			rt *= 2
			if rt > 60 {
				rt = 60
			}
			writeLog(fmt.Sprintf("Failed to connect to server (attempt %d of %d). Reason: %s", rc, MaxAttempts, err.Error()))
			writeLog(fmt.Sprintf("Retrying in %d seconds", rt))
			tick := time.After(time.Duration(rt) * time.Second)
			select {
			case <- Instance.CloseChannel:
				writeLog("Connect aborted")
				return nil
			case <- tick:
				writeLog("Reconnecting...")
				continue
			}
			return errors.New("unknown error occurred while attempting to reconnect")
		} else {
			Instance.Lock()
			Instance.Conn = s
			Instance.Connected = true
			Instance.Unlock()
			if Password != "" {
				Instance.send(fmt.Sprintf("PASS %s", Password))
			}
			Instance.send(fmt.Sprintf("NICK %s", Username))
			Instance.send(fmt.Sprintf("USER %s %s %s :%s", Username, Username, Address, Username))
			return nil
		}
	}
}

// Gracefully exit server and throw a stack trace when closing
func (Instance *IRCInstance) handle(s *net.TCPConn, e error) {
	Instance.send("QUIT :Segfault >:(")
	panic(e)
}

// Wrapper to write a message to the bot's log
func writeLog(s string) {
	fmt.Println("[" + time.Now().Format("2006/01/02 15:04:05") + "] <- " + s)
}

func connWatchdog(Instance *IRCInstance) {
	Instance.Lock()
	Instance.LastActive = time.Now()
	Instance.Unlock()
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-Instance.CloseChannel:
			if Instance.Connected {
				Instance.send("QUIT :Closing")
				Instance.Lock()
				Instance.Connected = false
				Instance.Conn.Close()
				Instance.Unlock()
				return
			}
		case <-tick:
			Instance.RLock()
			timeSinceActive := time.Since(Instance.LastActive)
			if !Instance.Connected {
				Instance.RUnlock()
				return
			}
			Instance.RUnlock()
			if timeSinceActive > 300*time.Second {
				writeLog("Connection appears dead, attempting reconnect")
				Instance.Lock()
				Instance.Connected = false
				Instance.Conn.Close()
				Instance.Unlock()
				return
			}
		}
	}
}
