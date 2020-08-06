package commands

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/route"
	"github.com/spf13/viper"
)

type _map struct {
	Sections byte     `arg:"" name:"sections" help:"The number of sections the map has"`
	Paths    []string `arg:"" name:"max_paths" help:"The max letter each section goes to. If sections was 4 then there should be 4 letters"`
}

func (m *_map) AfterApply(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	cmdPrefix := viper.GetString("COMMAND_PREFIX")

	if err := m.checkPermissions(sess, msg); err != nil {
		return err
	}

	// Checking that section is a number above 0
	if m.Sections < 1 {
		return UsageError{
			Param:   "sections",
			Message: "Must be greater than 0",
			Footer:  fmt.Sprintf("Type %smap --help for command usage", cmdPrefix),
		}
	}

	// Checking if each path provided is only one char and is between A-Z
	for i, p := range m.Paths {
		p = strings.ToUpper(p)
		m.Paths[i] = p

		// Checking if the path is only one character
		if len(p) != 1 {
			return UsageError{
				Param:   fmt.Sprintf("max_paths[%d]", i),
				Message: fmt.Sprintf(`Invalid argument "%s". Max path element must be 1 letter`, p),
				Footer:  fmt.Sprintf("Type %smap --help for command usage", cmdPrefix),
			}
		}

		// Checking that the path is between A and Z
		if p[0] < 'A' || p[0] > 'Z' {
			return UsageError{
				Param:   fmt.Sprintf("max_paths[%d]", i),
				Message: fmt.Sprintf(`Invalid argument "%s". Max path elements must be between A and Z`, p),
				Footer:  fmt.Sprintf("Type %smap --help for command usage", cmdPrefix),
			}
		}
	}

	// Checking that the correct number of paths was given
	if len(m.Paths) != int(m.Sections) {
		return UsageError{
			Param:   "max_paths",
			Message: fmt.Sprintf("Not enough max path elements provided for sections speficied (have %d, need %d)", len(m.Paths), m.Sections),
			Footer:  fmt.Sprintf("Type %smap --help for command usage", cmdPrefix),
		}
	}

	return nil
}

func (_map) checkPermissions(sess *discordgo.Session, m *discordgo.MessageCreate) error {
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

func (m *_map) Run(sess *discordgo.Session, msg *discordgo.MessageCreate, rs *route.Service) error {
	return rs.InTransaction(context.Background(), true, func(ctx context.Context, tx *bolt.Tx) error {

		// Clearing all linked routes for the channel
		err := rs.DeleteAllRoutesForChannel(ctx, msg.ChannelID)
		if err != nil {
			return SystemError{
				error:   err,
				Message: "Something went wrong when purging all routes for the channel",
				Stack:   debug.Stack(),
			}
		}

		// Inserting the map into the database
		err = rs.InsertMap(ctx, route.Map{
			ID:       msg.ChannelID,
			Sections: m.Sections,
			MaxPaths: m.Paths,
		})
		if err != nil {
			return SystemError{
				error:   err,
				Message: "Something went wrong when saving the map",
				Stack:   debug.Stack(),
			}
		}

		info := newInfoEmbed()
		info.Description = "All routes have been purged and a new map has been saved for this channel"
		sess.ChannelMessageSendEmbed(msg.ChannelID, info)
		return nil
	})
}
