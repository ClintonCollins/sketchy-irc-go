package sketchyircgo

import (
	"strings"
	"errors"
)

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

func parseModeChange(rawmessage string) (*ModeChange, error) {
	writeLog(rawmessage)
	rawMessageSplit := strings.Split(rawmessage, " ")
	if len(rawMessageSplit) < 5 {
		return &ModeChange{}, errors.New("couldn't parse mode change")
	}
	mode := rawMessageSplit[3]
	sender := strings.TrimPrefix(rawMessageSplit[0], ":")
	receiver := rawMessageSplit[4]
	channelName := rawMessageSplit[2]
	newMode := ModeChange{
		Channel: &Channel{Name: channelName},
		Mode: mode,
		Receiver: receiver,
		Sender: sender,
	}
	return &newMode, nil
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