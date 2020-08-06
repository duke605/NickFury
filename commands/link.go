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

// Link ...
type Link struct {
	Section int    `arg:"" name:"section" help:"The section to link yourself to"`
	Path    string `arg:"" name:"path" help:"The path to link yourself to"`

	User Mention `name:"user" help:"Sets the user that will be linked"`
}

// AfterApply ...
func (l Link) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, k *kong.Kong) error {
	cmdPrefix := viper.GetString("COMMAND_PREFIX")
	l.Path = strings.ToUpper(l.Path)

	// Checking if the user is allowed to use the command
	err := l.checkPermissions(sess, msg)
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

	// Checking if the section is in the acceptable range
	if l.Section < 1 || l.Section > int(m.Sections) {
		return UsageError{
			Param:    "section",
			Message:  fmt.Sprintf("Must be between 1 and %d (inclusive)", m.Sections),
			Provided: l.Path,
			Footer:   fmt.Sprintf("Type %slink --help for command usage", cmdPrefix),
		}
	}

	// Checking that the path param is a valid character
	if len(l.Path) > 1 {
		return UsageError{
			Param:    "path",
			Message:  "Must be a single character",
			Provided: l.Path,
			Footer:   fmt.Sprintf("Type %slink --help for command usage", cmdPrefix),
		}
	}

	// Checking that the path param is a valid path for the map
	if !m.IsValidPath(l.Section, l.Path) {
		paths := m.Paths(l.Section)
		return UsageError{
			Param:    "path",
			Message:  fmt.Sprintf("Must be between A and %s (inclusive)", paths[len(paths)-1]),
			Provided: l.Path,
			Footer:   fmt.Sprintf("Type %slink --help for command usage", cmdPrefix),
		}
	}

	return nil
}

func (l Link) checkPermissions(sess *discordgo.Session, m *discordgo.MessageCreate) error {
	// If the command was not invoked with the user flag then the command isn't restricted
	if l.User == "" {
		return nil
	}

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
func (l Link) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service, m route.Map) error {
	var newRoute route.Route
	var routes []route.Route
	var err error

	l.Path = strings.ToUpper(l.Path)
	userID := msg.Author.ID
	if l.User != "" {
		userID = string(l.User)
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

		// Checking if the user is already linked to the section
		for _, r := range routes {
			if r.Section == l.Section && r.Path == l.Path && r.UserID == userID {
				m := "You are already linked to path %s in section %d"
				if userID != msg.Author.ID {
					m = "User is already linked to path %s in section %d"
				}
				return Warning{
					Message: fmt.Sprintf(m, r.Path, l.Section),
				}
			}
		}

		// Inserting section into db
		newRoute = route.Route{
			ID:        msg.ID,
			UserID:    userID,
			ChannelID: msg.ChannelID,
			Path:      strings.ToUpper(l.Path),
			Section:   l.Section,
		}
		err = rs.InsertRoute(ctx, newRoute)
		if err != nil {
			return SystemError{
				error:   err,
				Message: "Something went wrong linking you to the new route",
				Stack:   debug.Stack(),
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
		key = fmt.Sprintf("%d:%s", r.Section, r.Path)
		idx[key] = append(idx[key], r.UserID)
	}

	newMsg, err := sess.ChannelMessageSendEmbed(msg.ChannelID, rs.ComposeEmbed(m, idx))
	if err == nil {
		cleanupPreviousRouteEmbeds(sess, newMsg.ChannelID, newMsg.ID)
	}

	return nil
}
