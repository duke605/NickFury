package commands

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kong"
	bolt "github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/route"
	"github.com/spf13/viper"
)

type unlink struct {
	Section int    `arg:"" optional:"" name:"section" help:"The section to link yourself to"`
	Path    string `arg:"" optional:"" name:"path" help:"The path to link yourself to"`

	User Mention `name:"user" help:"Sets the user that will be linked"`
}

func (u *unlink) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, k *kong.Kong) error {
	cmdPrefix := viper.GetString("COMMAND_PREFIX")
	u.Path = strings.ToUpper(u.Path)

	// Checking if the user is allowed to use the command
	err := u.checkPermissions(sess, msg)
	if err != nil {
		return err
	}

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
	kong.Bind(m).Apply(k)

	// Checking that the section param is in the acceptable range
	if u.Section < 1 || u.Section > int(m.Sections) {
		return UsageError{
			Param:    "section",
			Message:  fmt.Sprintf("Must be between 1 and %d (inclusive)", m.Sections),
			Provided: u.Path,
			Footer:   fmt.Sprintf("Type %sunlink --help for command usage", cmdPrefix),
		}
	}

	// Checking that the path param is a valid character
	if len(u.Path) > 1 {
		return UsageError{
			Param:    "path",
			Message:  "Must be a single character",
			Provided: u.Path,
			Footer:   fmt.Sprintf("Type %sunlink --help for command usage", cmdPrefix),
		}
	}

	// Checking that the path param is a valid path for the map
	if !m.IsValidPath(u.Section, u.Path) {
		paths := m.Paths(u.Section)
		return UsageError{
			Param:    "path",
			Message:  fmt.Sprintf("Must be between A and %s (inclusive)", paths[len(paths)-1]),
			Provided: u.Path,
			Footer:   fmt.Sprintf("Type %sunlink --help for command usage", cmdPrefix),
		}
	}

	return nil
}

func (u *unlink) checkPermissions(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	// If the command was not invoked with the user flag then the command isn't restricted
	if u.User == "" {
		return nil
	}

	auth, err := isUserOrRole(sess, msg.GuildID, msg.Author.ID, trustedUsers, trustedRoles)
	if err != nil {
		return SystemError{
			error:   err,
			Message: "Something went wrong getting your role information",
			Stack:   debug.Stack(),
		}
	}

	if !auth {
		return PermissionError{
			Message: "You do not have permission to use this command with the `user` flag",
		}
	}

	return nil
}

func (u *unlink) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, m route.Map) error {
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

			sectionsMatch := u.Section == 0 || u.Section == r.Section
			pathsMatch := u.Path == "" || u.Path == r.Path
			if sectionsMatch && pathsMatch {
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

	newMsg, err := sess.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
		Embed: rs.ComposeEmbed(m, idx),
		Content: func() string {
			m := "Unlinked you from **%d** route(s)"
			if userID != msg.Author.ID {
				m = "Unlinked user from **%d** route(s)"
			}

			return fmt.Sprintf(m, len(matchingRoutes))
		}(),
	})
	if err == nil {
		cleanupPreviousRouteEmbeds(sess, msg.ChannelID, newMsg.ID)
	}

	return nil
}
