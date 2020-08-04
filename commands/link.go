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

// Link ...
type Link struct {
	Section int    `arg:"" name:"section" help:"The section to link yourself to"`
	Path    string `arg:"" name:"path" help:"The path to link yourself to"`

	User Mention `name:"user" help:"Sets the user that will be linked"`
}

// AfterApply ...
func (l Link) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	l.Path = strings.ToUpper(l.Path)

	if l.Section < 1 || l.Section > 3 {
		return UsageError{
			Param:    "path",
			Message:  "Must be between 1 and 3 (inclusive)",
			Provided: l.Path,
			Footer:   "Type !link --help for command usage",
		}
	}

	if len(l.Path) > 1 {
		return UsageError{
			Param:    "path",
			Message:  "Must be a single character",
			Provided: l.Path,
			Footer:   "Type !link --help for command usage",
		}
	}

	limit := 4 + l.Section
	min, max := "A"[0], "ABCDEFG"[limit-1 : limit][0]
	if l.Path[0] < min || l.Path[0] > max {
		return UsageError{
			Param:    "path",
			Message:  fmt.Sprintf("Must be between %s and %s (inclusive)", string(min), string(max)),
			Provided: l.Path,
			Footer:   "Type !link --help for command usage",
		}
	}

	return l.checkPermissions(sess, msg)
}

func (l Link) checkPermissions(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	// If the command was not invoked with the user flag then the command isn't restricted
	if l.User == "" {
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
func (l Link) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service) error {
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

	m, err := sess.ChannelMessageSendEmbed(msg.ChannelID, rs.ComposeEmbed(idx))
	if err == nil {
		cleanupPreviousRouteEmbeds(sess, msg.ChannelID, m.ID)
	}

	return nil
}
