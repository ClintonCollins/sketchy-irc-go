package sketchyircgo

import (
	"strings"
	"net"
	"time"
	"fmt"
	"sync"
	"errors"
)

type IRCInstance struct {
	Address    string
	Username   string
	Password   string
	Connected  bool
	Conn       *net.TCPConn
	LastActive time.Time
	TwitchIRC  bool
	Channels   []*Channel
}

type Channel struct {
	Name       string
	Moderators []*User
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

var (
	IRCInstanceLock sync.RWMutex
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

func parseTwitchPrivMsg(rawMessage string) (*Message, error) {
	rawMessageSplit := strings.Split(rawMessage, " ")
	usernameSplit := rawMessageSplit[1]
	username := usernameSplit[1:]
	if len(rawMessageSplit) < 5 {
		return &Message{}, errors.New("couldn't parse message")
	}
	channel := rawMessageSplit[3]
	messageString := strings.Join(rawMessageSplit[4:], " ")
	message := strings.TrimPrefix(messageString, ":")
	badgesSplit := strings.Split(rawMessageSplit[0], ";")
	moderator := false
	turbo := false
	subscriber := false
	broadcaster := false
	staff := false
	globalMod := false
	displayName := ""
	for _, badge := range badgesSplit {
		switch badge {
		case "mod=1":
			moderator = true
			continue
		case "turbo=1":
			turbo = true
			continue
		case "subscriber=1":
			subscriber = true
			continue
		}
		if strings.HasPrefix(badge, "display-name=") {
			displayName = strings.TrimPrefix(badge, "display-name=")
		}
	}
	if strings.Contains(badgesSplit[0], "broadcaster") {
		broadcaster = true
	}
	if strings.Contains(badgesSplit[0], "global_mod") {
		globalMod = true
	}
	if strings.Contains(badgesSplit[0], "staff") {
		staff = true
	}
	user := User{
		Name:        username,
		Moderator:   moderator,
		Subscriber:  subscriber,
		Turbo:       turbo,
		DisplayName: displayName,
		Broadcaster: broadcaster,
		Staff:       staff,
		GlobalMod:   globalMod,
	}
	newMessage := Message{
		Message: message,
		Author:  &user,
		Channel: &Channel{Name: channel},
	}
	return &newMessage, nil
}

func parseIRCPrivMsg(rawMessage string) (*Message, error) {
	rawMessageSplit := strings.Split(rawMessage, " ")
	usernameSplit := rawMessageSplit[1]
	username := usernameSplit[1:]
	if len(rawMessageSplit) < 4 {
		return &Message{}, errors.New("couldn't parse message")
	}
	channel := rawMessageSplit[3]
	messageString := strings.Join(rawMessageSplit[3:], " ")
	message := strings.TrimPrefix(messageString, ":")
	user := User{
		Name: username,
	}
	newMessage := Message{
		Message: message,
		Author:  &user,
		Channel: &Channel{Name: channel},
	}
	return &newMessage, nil
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

func (Instance *IRCInstance) JoinChannel(channelName string) {
	if !strings.HasPrefix(channelName, "#") {
		channelName = "#" + channelName
	}
	Instance.send("JOIN " + channelName)
	if Instance.TwitchIRC {
		Instance.send("CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands")
	}
}

func New(address, username, password string) *IRCInstance {
	return &IRCInstance{Address: address, Username: username, Password: password}
}

func (Instance *IRCInstance) RunIRC() {
	Instance.connect(Instance.Address, Instance.Username, Instance.Password)
	go connWatchdog(Instance)
	for {
		buf := make([]byte, 8192)
		l, err := Instance.Conn.Read(buf)
		if err != nil {
			IRCInstanceLock.Lock()
			Instance.Connected = false
			IRCInstanceLock.Unlock()
			// TODO add reconnect
		}
		if l < 1 {
			continue
		}
		IRCInstanceLock.Lock()
		Instance.LastActive = time.Now()
		IRCInstanceLock.Unlock()
		rawMessageSplit := strings.Split(string(buf[:l]), "\r\n")
		for i := 0; i < len(rawMessageSplit); i++ {
			parsedMessageSplit := strings.Split(rawMessageSplit[i], " ")
			// Empty message.
			if len(parsedMessageSplit) <= 1 || parsedMessageSplit == nil {
				continue
			}
			switch parsedMessageSplit[0] { // Special packet processing
			case "PING":
				if len(parsedMessageSplit) < 2 {
					continue
				}
				Instance.send("PONG " + parsedMessageSplit[1])
				continue
			case "ERROR":
				writeLog("Connection to server closed. Reason: " + rawMessageSplit[0])
				return
			}
			if len(parsedMessageSplit) < 2 {
				continue
			}

			// Parse IRC Responses
			switch parsedMessageSplit[1] {
			case "001":
				/// Welcome message from server.
				Instance.sendServerReadyListener(Instance)
			case "353":
				//_, rawNames := parseMsg(recv[i])
				//names := strings.Split(rawNames, " ")
				continue
			case "366":
				//blankOwners = true
				continue
			case "JOIN":
				user, _ := parseSender(rawMessageSplit[i])
				writeLog("*** " + user + " has joined the channel.")
				writeLog(rawMessageSplit[i])
				continue
			case "PART":
				user, _ := parseSender(rawMessageSplit[i])
				writeLog("* " + user + " has left the channel.")
				continue
			case "QUIT":
				user, msg := parseMsg(rawMessageSplit[i])
				writeLog("* " + user + " has quit the chat. (" + msg + ")")
				continue
			case "MODE":
				user, _ := parseSender(rawMessageSplit[i])
				packet := strings.Split(rawMessageSplit[i], " ")
				if len(packet) < 5 {
					continue
				}
				if packet[4] == "" {
					//packet[4] = Bot.Channel
				}
				writeLog("*** " + user + " set mode " + packet[3] + " for " + packet[4])
				continue
			case "KICK":
				user, msg := parseMsg(rawMessageSplit[i])
				packet := strings.Split(rawMessageSplit[i], " ")
				if len(packet) < 4 {
					continue
				}
				writeLog("*** " + packet[3] + " was kicked from the channel by " + user + " [" + msg + "]")
				continue
			case "PRIVMSG":
				Instance.ircPRIVMSG(rawMessageSplit[i])
				continue
			case "NOTICE":
				user, msg := parseMsg(rawMessageSplit[i])
				if strings.ToUpper(msg) == "\001VERSION\001" {
					Instance.send("NOTICE " + user + " :\001VERSION SketchyIRCGo version 1.0 \001")
					writeLog("*** Version check from " + user)
					continue
				}
				writeLog("*** " + user + ": " + msg)
				continue
			}
			// Parse Twitch IRC Messages
			// Twitch is enabled so Twitch TAGS come before the IRC message.
			if Instance.TwitchIRC {
				switch parsedMessageSplit[2] {
				case "PRIVMSG":
					Instance.ircPRIVMSG(rawMessageSplit[i])
					continue
				case "USERSTATE":
					continue
				}
			}
		}
	}
}
