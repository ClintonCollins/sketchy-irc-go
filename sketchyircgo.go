package sketchyircgo

import (
	"strings"
	"time"
	"sync"
)

var (
	IRCInstanceLock sync.RWMutex
)

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
