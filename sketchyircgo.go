package sketchyircgo

import (
	"errors"
	"strings"
	"time"
)

func (Instance *IRCInstance) JoinChannel(channelName string) {
	if !strings.HasPrefix(channelName, "#") {
		channelName = "#" + channelName
	}
	Instance.send("JOIN " + channelName)
	if Instance.TwitchIRC {
		Instance.send("CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands")
	}
	Instance.Lock()
	newChannel := &Channel{}
	_, exists := Instance.Channels[channelName]
	if !exists {
		newChannel.Moderators = make(map[string]*User)
		newChannel.Users = make(map[string]*User)
		newChannel.Name = channelName
		Instance.Channels[channelName] = newChannel
	}
	Instance.Unlock()
}

func (Instance *IRCInstance) PartChannel(channelName string) {
	if !strings.HasPrefix(channelName, "#") {
		channelName = "#" + channelName
	}
	Instance.send("PART " + channelName)

	Instance.Lock()
	channel, exists := Instance.Channels[channelName]
	if exists {
		delete(Instance.Channels, channel.Name)
	}
	Instance.Unlock()
}

func New(address, username, password string) *IRCInstance {
	return &IRCInstance{Address: address,
		Username:     username,
		Password:     password,
		Connected:    false,
		CloseChannel: make(chan bool),
		Channels:     make(map[string]*Channel),
	}
}

func (Instance *IRCInstance) SendMessage(channelName, message string) {
	if !strings.HasPrefix(channelName, "#") {
		channelName = "#" + channelName
	}
	Instance.send("PRIVMSG " + channelName + " :" + message)
}

func (Instance *IRCInstance) Close() {
	Instance.CloseChannel <- true
}

func (Instance *IRCInstance) RunIRC() error {
	if err := Instance.connect(Instance.Address, Instance.Username, Instance.Password, -1); err != nil {
		return err
	}
	if !Instance.Connected { // connect bailed with no error, just exit
		return nil
	}
	go connWatchdog(Instance)
	for {
		buf := make([]byte, 8192)
		l, err := Instance.Conn.Read(buf)
		if err != nil {
			if !Instance.Connected { // disconnect was intentional, just exit
				return nil
			}
			if err := Instance.connect(Instance.Address, Instance.Username, Instance.Password, -1); err != nil {
				return err
			}
			if !Instance.Connected { // connect bailed with no error, just exit
				return nil
			}
			go connWatchdog(Instance)
			continue
		}
		if l < 1 {
			continue
		}
		Instance.Lock()
		Instance.LastActive = time.Now()
		Instance.Unlock()
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
				return errors.New("connection to server closed with reason " + rawMessageSplit[0])
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
				//user, _ := parseSender(rawMessageSplit[i])
				//writeLog("*** " + user + " has joined the channel.")
				//writeLog(rawMessageSplit[i])
				Instance.ircJOIN(rawMessageSplit[i])
				continue
			case "PART":
				//user, _ := parseSender(rawMessageSplit[i])
				//writeLog("* " + user + " has left the channel.")
				//writeLog(rawMessageSplit[i])
				Instance.ircPART(rawMessageSplit[i])
				continue
			case "QUIT":
				user, msg := parseMsg(rawMessageSplit[i])
				writeLog("* " + user + " has quit the chat. (" + msg + ")")
				continue
			case "MODE":
				Instance.ircMODE(rawMessageSplit[i])
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
