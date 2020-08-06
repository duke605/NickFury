package commands

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"

	"github.com/alecthomas/kong"
	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/route"
	"github.com/spf13/viper"
)

type show struct{}

func (show) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, k *kong.Kong) error {
	cmdPrefix := viper.GetString("COMMAND_PREFIX")

	// Checking if a map exists for the channel
	m, err := rs.GetMapForChannel(context.Background(), msg.ChannelID)
	if err != nil && err != sql.ErrNoRows {
		return SystemError{
			error:   err,
			Message: "Something went wrong getting the map for this channel",
			Stack:   debug.Stack(),
		}
	} else if err == sql.ErrNoRows {
		return Warning{
			Message: fmt.Sprintf("There is no map configured for this channel. Use the `%smap` command to configure one", cmdPrefix),
		}
	}

	return kong.Bind(m).Apply(k)
}

func (show) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, m route.Map) error {
	var routes []route.Route
	var err error

	ctx := context.Background()
	routes, err = rs.GetRoutesInChannel(ctx, msg.ChannelID)
	if err != nil {
		return SystemError{
			error:   err,
			Message: "Something went wrong getting linked routes for channel",
			Stack:   debug.Stack(),
		}
	}

	// Creating a map for quick lookup of assigned routes
	idx := map[string][]string{}
	for _, r := range routes {
		key := fmt.Sprintf("%d:%s", r.Section, r.Path)
		idx[key] = append(idx[key], r.UserID)
	}

	// Sending route list to channel and then cleaning up previous lists
	newMsg, err := sess.ChannelMessageSendEmbed(msg.ChannelID, rs.ComposeEmbed(m, idx))
	if err == nil {
		cleanupPreviousRouteEmbeds(sess, newMsg.ChannelID, newMsg.ID)
	}

	return nil
}
