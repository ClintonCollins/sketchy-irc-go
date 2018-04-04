package sketchyircgo

func (Instance *IRCInstance) ircPRIVMSG(rawMessage string) {
	if Instance.TwitchIRC {
		message, err := Instance.parseTwitchPrivMsg(rawMessage)
		if err != nil {
			writeLog("Error parsing Twitch message.")
			return
		}
		Instance.sendMessageListener(Instance, message)
	} else {
		message, err := Instance.parseIRCPrivMsg(rawMessage)
		if err != nil {
			writeLog("Error parsing IRC message.")
			return
		}
		Instance.sendMessageListener(Instance, message)
	}
}

func (Instance *IRCInstance) ircMODE(rawMessage string) {
	modeChange, err := Instance.parseModeChange(rawMessage)
	if err != nil {
		writeLog("Error parsing mode change.")
		return
	}

	// Let's handle adding and removing moderators to channels.
	if modeChange.Mode == "+o" {
		// Find the user if they exist already in the channel, otherwise create a new one.
		Instance.ChannelsLock.Lock()
		newUser := &User{}
		user, exists := modeChange.Channel.Users[modeChange.Receiver]
		if !exists {
			newUser.Name = modeChange.Receiver
			newUser.DisplayName = modeChange.Receiver
			user = newUser
		}
		modeChange.Channel.Moderators[user.Name] = user
		Instance.ChannelsLock.Unlock()

	} else if modeChange.Mode == "-o" { // If mode is -o that means moderator was taken away.

	Instance.ChannelsLock.Lock()
		// Find the user if they exist already in moderators, otherwise create a new one.
		_, exists := modeChange.Channel.Moderators[modeChange.Receiver]
		if exists {
			delete(modeChange.Channel.Moderators, modeChange.Receiver)
		}
		Instance.ChannelsLock.Unlock()
	}

	Instance.sendModeChangeListener(Instance, modeChange)
}

func (Instance *IRCInstance) ircJOIN(rawMessage string) {
	userJoin, err := Instance.parseIRCJoin(rawMessage)
	if err != nil {
		writeLog("Error parsing user join.")
		return
	}
	Instance.ChannelsLock.Lock()
	channel := userJoin.Channel
	newUser := &User{}
	user, exists := userJoin.Channel.Users[userJoin.User.Name]
	if !exists {
		newUser.Name = userJoin.User.Name
		newUser.DisplayName = userJoin.User.Name
		user = newUser
	}
	channel.Users[user.Name] = user
	Instance.ChannelsLock.Unlock()
	Instance.sendUserJoinListener(Instance, userJoin)
}

func (Instance *IRCInstance) ircPART(rawMessage string) {
	userPart, err := Instance.parseIRCPart(rawMessage)
	if err != nil {
		writeLog("Error parsing user join.")
		return
	}
	Instance.ChannelsLock.Lock()
	channel := userPart.Channel
	user, exists := userPart.Channel.Users[userPart.User.Name]
	if exists {
		delete(channel.Users, user.Name)
	}
	Instance.ChannelsLock.Unlock()
	Instance.sendUserJPartListener(Instance, userPart)
}