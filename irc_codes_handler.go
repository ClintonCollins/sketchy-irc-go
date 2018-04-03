package sketchyircgo

import (
	"github.com/prometheus/common/log"
)

func (Instance *IRCInstance) ircPRIVMSG(rawMessage string) {
	if Instance.TwitchIRC {
		message, err := parseTwitchPrivMsg(rawMessage)
		if err != nil {
			log.Error("Error parsing Twitch message.")
			return
		}
		Instance.sendMessageListener(Instance, message)
	} else {
		message, err := parseIRCPrivMsg(rawMessage)
		if err != nil {
			log.Error("Error parsing IRC message.")
			return
		}
		Instance.sendMessageListener(Instance, message)
	}
}
