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
func (Purge) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate) error {

	// Checking if the user is authorized
	authUsers := map[string]struct{}{
		"136856172203474944": {}, // Duke605
		"213531207936245761": {}, // Mysterio
	}
	if _, ok := authUsers[msg.Author.ID]; ok {
		return nil
	}

	// Getting role information about user
	mem, err := getMember(sess, msg.GuildID, msg.Author.ID)
	if err != nil {
		return SystemError{
			error:   err,
			Message: "Something went wrong getting your role information",
			Stack:   debug.Stack(),
		}
	}

	// Checking if the user has any authorized roles
	authRoles := map[string]struct{}{
		"735605674343661639": {}, // Sith Load Officers
		"735605674343661644": {}, // officers
		"735605674343661645": {}, // leaders
	}
	if isAuthorized(mem, authRoles) {
		return nil
	}
	return PermissionError{
		Message: "You do not have permission to run use this command",
	}
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
