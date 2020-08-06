package commands

import (
	"fmt"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

var trustedUsers = map[string]struct{}{
	"213531207936245761": {}, // Mysterio
	"136856172203474944": {}, // Duke605
}

var trustedRoles = map[string]struct{}{
	"735605674343661639": {}, // Sith Load Officers
	"735605674343661644": {}, // officers
	"735605674343661645": {}, // leaders
}

// Root ...
type Root struct {
	Link   Link   `cmd:"" help:"Links yourself to a section and path"`
	Unlink unlink `cmd:"" help:"Unlinks yourself from a path and/or section"`
	Show   show   `cmd:"" help:"Shows all assigned and unassigned routes for the channel"`
	Map    _map   `cmd:"" help:"Configures the map for the channel"`
	Purge  Purge  `cmd:"" help:"Clears all linked routes for the channel"`
	Ping   Ping   `cmd:"" help:"Diagnostics command"`
	About  About  `cmd:"" help:"Shows information about this bot"`
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

// hasRole determines if the provided member has a role in the authorizedRoles set provided
func hasRole(m *discordgo.Member, autorizedRoles map[string]struct{}) bool {
	for _, r := range m.Roles {
		if _, ok := autorizedRoles[r]; ok {
			return true
		}
	}

	return false
}

// isUserOrRole detemines if the user is in the authUserIDs set or if the user has a role
// in the authRoles set.
func isUserOrRole(sess *discordgo.Session, guildID, userID string, authUserIDs, authRoles map[string]struct{}) (bool, error) {
	if _, ok := authUserIDs[userID]; ok {
		return true, nil
	}

	mem, err := getMember(sess, guildID, userID)
	if err != nil {
		return false, err
	}

	return hasRole(mem, authRoles), nil
}

// newInfoEmbed creates a new info embed with some field prefilled
func newInfoEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color: 0x3C9EFF,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: "https://i.imgur.com/EPHiUNu.png",
			Name:    "Info",
		},
	}
}
