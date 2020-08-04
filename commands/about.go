package commands

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

// About ...
type About struct{}

// Run ...
func (About) Run(sess *discordgo.Session, m *discordgo.MessageCreate) error {
	duke, err := sess.User("136856172203474944")
	if err != nil {
		return SystemError{
			error:   err,
			Message: "Something went wrong getting the developer's information",
			Stack:   debug.Stack(),
		}
	}

	imgs := []string{
		"https://gophercises.com/img/gophercises_lifting.gif",
		"https://gophercises.com/img/gophercises_punching.gif",
		"https://gophercises.com/img/gophercises_jumping.gif",
	}

	sess.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    duke.Username,
			IconURL: duke.AvatarURL("32"),
		},
		Description: "Please file an issue at https://github.com/duke605/NickFury if you find any bugs.",
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: imgs[rand.Intn(len(imgs))],
			Text:    fmt.Sprintf("Running Golang v%s", runtime.Version()[2:]),
		},
	})
	return nil
}
