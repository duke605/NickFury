package commands

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/route"
)

// Show ...
type Show struct{}

// Run ...
func (Show) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, t time.Time) error {
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

	m, err := sess.ChannelMessageSendEmbed(msg.ChannelID, rs.ComposeEmbed(idx))
	if err == nil {
		cleanupPreviousRouteEmbeds(sess, m.ChannelID, m.ID)
	}

	return nil
}
