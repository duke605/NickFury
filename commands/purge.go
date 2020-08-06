package commands

import (
	"context"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/route"
)

// Purge ...
type Purge struct{}

// AfterApply ...
func (Purge) AfterApply(sess *discordgo.Session, m *discordgo.MessageCreate) error {
	auth, err := isUserOrRole(sess, m.GuildID, m.Author.ID, trustedUsers, trustedRoles)
	if err != nil {
		return SystemError{
			error:   err,
			Message: "Something went wrong getting your role information",
			Stack:   debug.Stack(),
		}
	}

	if !auth {
		return PermissionError{
			Message: "You do not have permission to use this command",
		}
	}

	return nil
}

// Run ...
func (Purge) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service) error {
	var err error

	err = rs.DeleteAllRoutesForChannel(context.Background(), msg.ChannelID)
	if err != nil {
		return SystemError{
			error:   err,
			Message: "Something went wrong purging all linked routes for the channel",
			Stack:   debug.Stack(),
		}
	}

	sess.ChannelMessageSend(msg.ChannelID, "All routes have been purged for this channel")
	return nil
}
