package sketchy_irc_go

import (
	"strings"
	"net"
	"time"
	"fmt"
)

type IRCInstance struct {
	Address string
	Username string
	Password string
	Connected bool
	Conn *net.TCPConn
	LastActive time.Time
}

func send(c *net.TCPConn, s string) {
	c.Write([]byte(s + "\r\n"))
	fmt.Println("[" + time.Now().Format("2006/01/02 15:04:05") + "] -> " + s)
}

// Wrapper to easily connect to an IRC server
func connect(Address, Username, OAuth string) (socket *net.TCPConn, error bool) {
	raddr, err := net.ResolveTCPAddr("tcp", Address)
	if err != nil {
		panic(err)
	}
	s, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return s, true
	}
	send(s, "PASS "+OAuth)
	send(s, "NICK "+Username)
	send(s, "USER "+Username+Username+Address+" :"+Username)
	return s, false
}

// Gracefully exit server and throw a stack trace when closing
func handle(s *net.TCPConn, e error) {
	send(s, "QUIT :Segfault >:O")
	panic(e)
}

// Parses most IRC packets
func parseTagsMsg(s string) (user, msg string) {
	user, _ = parseTagsSender(s)
	msg = s
	chop := msg[strings.Index(s, " :")+2:]
	msg = chop[strings.Index(chop, " :")+2:]
	return user, msg
}

// Parses sender from IRC packets
func parseTagsSender(s string) (user, host string) {
	temp := strings.Split(s, " ")
	user = temp[1]
	user = user[1:]
	host = ""
	if strings.Contains(user, "!") {
		host = user[strings.Index(temp[1], "!"):]
		user = user[:strings.Index(temp[1], "!")-1]
	}
	return user, host
}

// Parses most IRC packets
func parseMsg(s string) (user, msg string) {
	user, _ = parseSender(s)
	msg = s
	msg = msg[strings.Index(s, " :")+2:]
	return user, msg
}

// Parses sender from IRC packets
func parseSender(s string) (user, host string) {
	temp := strings.Split(s, " ")
	user = temp[0]
	user = user[1:]
	host = ""
	if strings.Contains(user, "!") {
		host = user[strings.Index(temp[0], "!"):]
		user = user[:strings.Index(temp[0], "!")-1]
	}
	return user, host
}

// Some string parsing functions to make life easy
func stringLeft(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func stringTrimLeft(s string, n int) string {
	if len(s) <= n {
		return ""
	}
	return s[n:]
}

func stringRight(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func stringTrimRight(s string, n int) string {
	if len(s) <= n {
		return ""
	}
	return s[:len(s)-n]
}

// End string parsing functions

// Wrapper to write a message to the bot's log
func writeLog(s string) {
	fmt.Println("[" + time.Now().Format("2006/01/02 15:04:05") + "] <- " + s)
}

func connWatchdog(Instance *IRCInstance) {
	Instance.LastActive = time.Now()
	for {
		time.Sleep(1 * time.Second)
		if time.Since(Instance.LastActive) > 300 * time.Second {
			Instance.Conn.Close()
			return
		}
	}
}

func New(address, username, password string) *IRCInstance {
	return &IRCInstance{Address: address, Username: username, Password: password, Connected: false}
}

func (Instance *IRCInstance) Run() {
	var sock *net.TCPConn
	for {
		tempSock, err := connect(Instance.Address, Instance.Username, Instance.Password)
		if !err {
			sock = tempSock
			Instance.Connected = true
			break
		}
		writeLog("*** ERROR: Failed to connect to server, retrying in 10 seconds")
		time.Sleep(10 * time.Second)
	}
	go connWatchdog(Instance)
	for {
		buf := make([]byte, 8192)
		l, err := sock.Read(buf)
		if err != nil {
			Instance.Connected = false
			handle(sock, err) // TODO: Make this not just crash and burn on disconnect
		}
		if l < 1 {
			continue
		}
		Instance.LastActive = time.Now()
		recv := strings.Split(string(buf[:l]), "\r\n")
		for i := 0; i < len(recv); i++ {
			temp := strings.Split(recv[i], " ")
			switch temp[0] { // Special packet processing
			case "PING":
				if len(temp) < 2 {
					continue
				}
				send(sock, "PONG "+temp[1])
				continue
			case "ERROR":
				writeLog("Connection to server closed. Reason: " + recv[0])
				return
			}
			if len(temp) < 2 {
				continue
			}
			switch temp[1] {
			case "001":
				/// DISABLED FOR NOW
				//TODO change this up
				//send(sock, "JOIN "+Bot.Channel)
				//send(sock, "CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands")
			case "353":
				//_, rawNames := parseMsg(recv[i])
				//names := strings.Split(rawNames, " ")
				continue
			case "366":
				//blankOwners = true
				continue
			case "JOIN":
				user, _ := parseSender(recv[i])
				writeLog("*** " + user + " has joined the channel.")
				writeLog(recv[i])
				continue
			case "PART":
				user, _ := parseSender(recv[i])
				writeLog("* " + user + " has left the channel.")
				continue
			case "QUIT":
				user, msg := parseMsg(recv[i])
				writeLog("* " + user + " has quit the chat. (" + msg + ")")
				continue
			case "MODE":
				user, _ := parseSender(recv[i])
				packet := strings.Split(recv[i], " ")
				if len(packet) < 5 {
					continue
				}
				if packet[4] == "" {
					//packet[4] = Bot.Channel
				}
				writeLog("*** " + user + " set mode " + packet[3] + " for " + packet[4])
				continue
			case "KICK":
				user, msg := parseMsg(recv[i])
				packet := strings.Split(recv[i], " ")
				if len(packet) < 4 {
					continue
				}
				writeLog("*** " + packet[3] + " was kicked from the channel by " + user + " [" + msg + "]")
				continue
			case "PRIVMSG":
				user, msg := parseMsg(recv[i])
				if temp[2] == strings.ToLower(Instance.Username) {
					temp[2] = user
					if strings.ToUpper(stringLeft(msg, 7)) == "\001ACTION" && stringRight(msg, 1) == "\001" {
						writeLog("* " + user + " " + stringTrimLeft(stringTrimRight(msg, 1), 8))
					} else {
						writeLog("From " + user + ": " + msg)
					}
				} else {
					if strings.ToUpper(stringLeft(msg, 7)) == "\001ACTION" && stringRight(msg, 1) == "\001" {
						writeLog("* " + user + " " + stringTrimLeft(stringTrimRight(msg, 1), 8))
					} else {
						writeLog("<" + user + "> " + msg)
					}
				}
				writeLog(recv[i])
				continue
			case "NOTICE":
				user, msg := parseMsg(recv[i])
				if strings.ToUpper(msg) == "\001VERSION\001" {
					send(sock, "NOTICE "+user+" :\001VERSION SketchyIRCGo version 1.0 \001")
					writeLog("*** Version check from " + user)
					continue
				}
				writeLog("*** " + user + ": " + msg)
				continue
			}
			// This switch if for Twitch.tv Tags Enabled //
			switch temp[2] {
			case "PRIVMSG":
				user, msg := parseTagsMsg(recv[i])
				if user == strings.ToLower(Instance.Username) {
					if strings.ToUpper(stringLeft(msg, 7)) == "\001ACTION" && stringRight(msg, 1) == "\001" {
						writeLog("* " + user + " " + stringTrimLeft(stringTrimRight(msg, 1), 8))
					} else {
						writeLog("From " + user + ": " + msg)
					}
				} else {
					if strings.ToUpper(stringLeft(msg, 7)) == "\001ACTION" && stringRight(msg, 1) == "\001" {
						writeLog("* " + user + " " + stringTrimLeft(stringTrimRight(msg, 1), 8))
					} else {
						writeLog("<" + user + "> " + msg)
					}
				}

				continue
			case "USERSTATE":
				continue
			}
			writeLog(recv[i])
		}
	}
}
