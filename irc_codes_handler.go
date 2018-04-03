package sketchyircgo

func (Instance *IRCInstance) ircPRIVMSG(rawMessage string) {
	if Instance.TwitchIRC {
		message, err := parseTwitchPrivMsg(rawMessage)
		if err != nil {
			writeLog("Error parsing Twitch message.")
			return
		}
		Instance.sendMessageListener(Instance, message)
	} else {
		message, err := parseIRCPrivMsg(rawMessage)
		if err != nil {
			writeLog("Error parsing IRC message.")
			return
		}
		Instance.sendMessageListener(Instance, message)
	}
}

func (Instance *IRCInstance) ircMODE(rawMessage string) {
	modeChange, err := parseModeChange(rawMessage)
	if err != nil {
		writeLog("Error parsing mode change.")
		return
	}
	Instance.sendModeChangeListener(Instance, modeChange)
}
