package commands

import (
	"fmt"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

// Root ...
type Root struct {
	Link   Link   `cmd:"" help:"Links yourself to a section and path"`
	Unlink Unlink `cmd:"" help:"Unlinks yourself from a path and/or section"`
	Show   Show   `cmd:"" help:"Shows all assigned and unassigned routes for the channel"`
	Purge  Purge  `cmd:"" help:"Clears all linked routes for the channel"`
	Ping   Ping   `cmd:"" help:"Diagnostics command"`
}

// cleanupPreviousRouteEmbeds deletes messages from the bot that are route embeds that come
// before the provided message id on a channel
func cleanupPreviousRouteEmbeds(sess *discordgo.Session, channelID, messageID string) {
	msgs, err := sess.ChannelMessages(channelID, 10, messageID, "", "")
	if err != nil {
		fmt.Println("Error occured cleaning up messages: ", err)
		return
	}

	// Finding route messages and deleting them
	for _, m := range msgs {
		if m.Author.ID != sess.State.User.ID || len(m.Embeds) == 0 {
			continue
		}

		// Checking if embed is for routes
		e := m.Embeds[0]
		if e.Author != nil && e.Author.Name != "Routes" {
			continue
		}

		// Message is for routes
		sess.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

// getMember returns a member. Function will attempt to get the member from the state
// first before making a request to discord.
func getMember(sess *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	mem, err := sess.State.Member(guildID, userID)
	if err == discordgo.ErrStateNotFound {
		mem, err = sess.GuildMember(guildID, userID)
	}
	if err != nil {
		return nil, SystemError{
			error:   err,
			Message: "Something went wrong getting your role information",
			Stack:   debug.Stack(),
		}
	}

	return mem, nil
}

// isAutorized determines if the provided member has one of the authorized roles
// and returns true if they do.
func isAuthorized(m *discordgo.Member, autorizedRoles map[string]struct{}) bool {
	for _, r := range m.Roles {
		if _, ok := autorizedRoles[r]; ok {
			return true
		}
	}

	return false
}
