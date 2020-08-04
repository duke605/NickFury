package commands

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	bolt "github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/route"
)

// Unlink ...
type Unlink struct {
	Section int    `arg:"" optional:"" name:"section" help:"The section to link yourself to"`
	Path    string `arg:"" optional:"" name:"path" help:"The path to link yourself to"`

	User Mention `name:"user" help:"Sets the user that will be linked"`
}

// AfterApply ...
func (u Unlink) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	u.Path = strings.ToUpper(u.Path)

	if u.Section != 0 && (u.Section < 1 || u.Section > 3) {
		return UsageError{
			Param:    "path",
			Message:  "Must be between 1 and 3 (inclusive)",
			Provided: u.Path,
			Footer:   "Type !link --help for command usage",
		}
	}

	if u.Path != "" && len(u.Path) > 1 {
		return UsageError{
			Param:    "path",
			Message:  "Must be a single character",
			Provided: u.Path,
			Footer:   "Type !link --help for command usage",
		}
	}

	if u.Path != "" {
		limit := 4 + u.Section
		min, max := "A"[0], "ABCDEFG"[limit-1 : limit][0]
		if u.Path[0] < min || u.Path[0] > max {
			return UsageError{
				Param:    "path",
				Message:  fmt.Sprintf("Must be between %s and %s (inclusive)", string(min), string(max)),
				Provided: u.Path,
				Footer:   "Type !link --help for command usage",
			}
		}
	}

	return u.checkPermissions(sess, msg)
}

func (u Unlink) checkPermissions(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	// If the command was not invoked with the user flag then the command isn't restricted
	if u.User == "" {
		return nil
	}

	// Checking if the user is authorized
	authUsers := map[string]struct{}{
		"213531207936245761": {}, // Mysterio
		"136856172203474944": {}, // Duke605
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
	if !isAuthorized(mem, authRoles) {
		return PermissionError{
			Message: "You do not have permission to use the `user` flag for this command",
		}
	}

	return nil
}

// Run ...
func (u Unlink) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service) error {
	var newRoute route.Route
	var routes []route.Route
	var err error
	matchingRoutes := map[string]struct{}{}

	u.Path = strings.ToUpper(u.Path)
	userID := msg.Author.ID
	if u.User != "" {
		userID = string(u.User)
	}

	ctx := context.Background()
	err = rs.InTransaction(ctx, true, func(ctx context.Context, _ *bolt.Tx) error {

		// Getting the already selected routes for the channel
		routes, err = rs.GetRoutesInChannel(ctx, msg.ChannelID)
		if err != nil {
			return SystemError{
				error:   err,
				Message: "Something when wrong getting linked routes for channel",
				Stack:   debug.Stack(),
			}
		}

		// Finding all the routes the user is linked to that match the arguments given
		for _, r := range routes {

			// Only looking for routes that are for the user
			if userID != r.UserID {
				continue
			}

			if (u.Section == 0 && u.Path == "") ||
				(u.Section == r.Section && u.Path == "") ||
				(u.Section == r.Section && u.Path == r.Path) {
				rs.DeleteRoute(ctx, r)
				matchingRoutes[string(r.GetID())] = struct{}{}
			}
		}
		if len(matchingRoutes) == 0 {
			m := "You are not currently linked to any routes in this channel"
			if userID != msg.Author.ID {
				m = "User is not currently linked to any routes in this channel"
			}

			return Warning{
				Message: m,
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Creating a map for quick lookup of assigned routes
	idx := map[string][]string{}
	key := fmt.Sprintf("%d:%s", newRoute.Section, newRoute.Path)
	idx[key] = append(idx[key], newRoute.UserID)
	for _, r := range routes {

		// Ignoring routes that were removed
		if _, ok := matchingRoutes[string(r.GetID())]; ok {
			continue
		}

		key = fmt.Sprintf("%d:%s", r.Section, r.Path)
		idx[key] = append(idx[key], r.UserID)
	}

	m, err := sess.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
		Embed: rs.ComposeEmbed(idx),
		Content: func() string {
			m := "Unlinked you from **%d** route(s)"
			if userID != msg.Author.ID {
				m = "Unlinked user from **%d** route(s)"
			}

			return fmt.Sprintf(m, len(matchingRoutes))
		}(),
	})
	if err == nil {
		cleanupPreviousRouteEmbeds(sess, msg.ChannelID, m.ID)
	}

	return nil
}
