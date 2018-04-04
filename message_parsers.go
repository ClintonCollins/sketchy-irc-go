package sketchyircgo

import (
	"errors"
	"strings"
	"time"
)

func (Instance *IRCInstance) parseIRCPrivMsg(rawMessage string) (*Message, error) {
	rawMessageSplit := strings.Split(rawMessage, " ")
	usernameSplit := rawMessageSplit[1]
	username := usernameSplit[1:]
	if len(rawMessageSplit) < 4 {
		return &Message{}, errors.New("couldn't parse message")
	}
	channelName := rawMessageSplit[3]
	messageString := strings.Join(rawMessageSplit[3:], " ")
	message := strings.TrimPrefix(messageString, ":")

	// Find the channel if it exists already for the instance, otherwise create a new one.
	Instance.Lock()
	newChannel := &Channel{}
	channel, exists := Instance.Channels[channelName]
	if !exists {
		newChannel.Moderators = make(map[string]*User)
		newChannel.Users = make(map[string]*User)
		newChannel.Name = channelName
		channel = newChannel
	}
	Instance.Unlock()

	// Find the user if they exist already in the channel, otherwise create a new one.
	Instance.Lock()
	newUser := &User{}
	user, exists := channel.Users[username]
	if !exists {
		newUser.Name = username
		newUser.DisplayName = username
		user = newUser
	}
	Instance.Unlock()

	newMessage := Message{
		Message: message,
		Author:  user,
		Channel: channel,
	}
	return &newMessage, nil
}

func (Instance *IRCInstance) parseModeChange(rawmessage string) (*ModeChange, error) {
	rawMessageSplit := strings.Split(rawmessage, " ")
	if len(rawMessageSplit) < 5 {
		return &ModeChange{}, errors.New("couldn't parse mode change")
	}
	mode := rawMessageSplit[3]
	sender := strings.TrimPrefix(rawMessageSplit[0], ":")
	receiver := rawMessageSplit[4]
	channelName := rawMessageSplit[2]

	// Find the channel if it exists already for the instance, otherwise create a new one.
	Instance.Lock()
	newChannel := &Channel{}
	channel, exists := Instance.Channels[channelName]
	if !exists {
		newChannel.Moderators = make(map[string]*User)
		newChannel.Users = make(map[string]*User)
		newChannel.Name = channelName
		channel = newChannel
	}
	Instance.Unlock()

	newMode := ModeChange{
		Channel:  channel,
		Mode:     mode,
		Receiver: receiver,
		Sender:   sender,
	}
	return &newMode, nil
}

func (Instance *IRCInstance) parseIRCJoin(rawMessage string) (*UserJoin, error) {
	rawMessageSplit := strings.Split(rawMessage, " ")
	if len(rawMessageSplit) < 3 {
		return &UserJoin{}, errors.New("unable to parse user join")
	}
	usernameSplit := strings.Split(rawMessageSplit[0], "!")
	username := strings.TrimPrefix(usernameSplit[0], ":")
	channelName := rawMessageSplit[2]

	// Find the channel if it exists already for the instance, otherwise create a new one.
	Instance.Lock()
	newChannel := &Channel{}
	channel, exists := Instance.Channels[channelName]
	if !exists {
		newChannel.Moderators = make(map[string]*User)
		newChannel.Users = make(map[string]*User)
		newChannel.Name = channelName
		channel = newChannel
	}
	Instance.Unlock()

	// Find the user if they exist already in the channel, otherwise create a new one.
	Instance.Lock()
	newUser := &User{}
	user, exists := channel.Users[username]
	if !exists {
		newUser.Name = username
		newUser.DisplayName = username
		user = newUser
	}
	Instance.Unlock()

	// Create new user join object.
	newUserJoin := UserJoin{
		Channel: channel,
		User:    user,
		Time:    time.Now(),
	}
	return &newUserJoin, nil
}

func (Instance *IRCInstance) parseIRCPart(rawMessage string) (*UserPart, error) {
	rawMessageSplit := strings.Split(rawMessage, " ")
	if len(rawMessageSplit) < 3 {
		return &UserPart{}, errors.New("unable to parse user part")
	}
	usernameSplit := strings.Split(rawMessageSplit[0], "!")
	username := strings.TrimPrefix(usernameSplit[0], ":")
	channelName := rawMessageSplit[2]

	// Find the channel if it exists already for the instance, otherwise create a new one.
	Instance.Lock()
	newChannel := &Channel{}
	channel, exists := Instance.Channels[channelName]
	if !exists {
		newChannel.Moderators = make(map[string]*User)
		newChannel.Users = make(map[string]*User)
		newChannel.Name = channelName
		channel = newChannel
	}
	Instance.Unlock()

	// Find the user if they exist already in the channel, otherwise create a new one.
	Instance.Lock()
	newUser := &User{}
	user, exists := channel.Users[username]
	if !exists {
		newUser.Name = username
		newUser.DisplayName = username
		user = newUser
	}
	Instance.Unlock()

	// Create new user part object.
	newUserPart := UserPart{
		Channel: channel,
		User:    user,
		Time:    time.Now(),
	}
	return &newUserPart, nil
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

func (Instance *IRCInstance) parseTwitchPrivMsg(rawMessage string) (*Message, error) {
	rawMessageSplit := strings.Split(rawMessage, " ")
	usernameSplit := rawMessageSplit[1]
	username := usernameSplit[1:]
	if len(rawMessageSplit) < 5 {
		return &Message{}, errors.New("couldn't parse message")
	}
	channelName := rawMessageSplit[3]
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

	// Find the channel if it exists already for the instance, otherwise create a new one.
	Instance.Lock()
	newChannel := &Channel{}
	channel, exists := Instance.Channels[channelName]
	if !exists {
		newChannel.Moderators = make(map[string]*User)
		newChannel.Users = make(map[string]*User)
		newChannel.Name = channelName
		channel = newChannel
	}
	Instance.Unlock()

	// Find the user if they exist already in the channel, otherwise create a new one.
	Instance.Lock()
	newUser := &User{}
	user, exists := channel.Users[username]
	if !exists {
		newUser.Name = username
		newUser.Moderator = moderator
		newUser.Subscriber = subscriber
		newUser.Turbo = turbo
		newUser.DisplayName = displayName
		newUser.Broadcaster = broadcaster
		newUser.Staff = staff
		newUser.GlobalMod = globalMod
		user = newUser
	} else {
		user.Name = username
		user.Moderator = moderator
		user.Subscriber = subscriber
		user.Turbo = turbo
		user.DisplayName = displayName
		user.Broadcaster = broadcaster
		user.Staff = staff
		user.GlobalMod = globalMod

		channel.Users[user.Name] = user
	}
	Instance.Unlock()

	newMessage := Message{
		Message: message,
		Author:  user,
		Channel: channel,
	}
	return &newMessage, nil
}
