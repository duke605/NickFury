package main

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/duke605/NickFury/commands"
)

var usageErrorPattern = regexp.MustCompile(`^<(.+)>: (.+)$`)

// sendEphemeralMessage send a message to a channel and deletes the message
// after the duration has elapsed
func sendEphemeralMessage(channelID, msg string, d time.Duration) {
	m, err := bot.ChannelMessageSend(channelID, msg)
	if err != nil {
		return
	}

	time.Sleep(d)
	bot.ChannelMessageDelete(m.ChannelID, m.ID)
}

// errorToEmbed attempts to transform the error into a message embed if the error
// provided is one that can be handled by this function. Returns nil if the error
// is not handled by this function
func errorToEmbed(err error) *discordgo.MessageEmbed {
	err = getActualError(err)
	embed := &discordgo.MessageEmbed{
		Color: 0xFF0033,
	}

	// Formatting an error message for usage errors
	if ue := new(commands.UsageError); errors.As(err, ue) {
		embed.Author = &discordgo.MessageEmbedAuthor{IconURL: "https://i.imgur.com/6qVd0MG.png", Name: "Usage Error"}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  ue.Param,
			Value: ue.Message,
		})

		// Adding footer if provided
		if ue.Footer != "" {
			embed.Footer = &discordgo.MessageEmbedFooter{Text: ue.Footer}
		}

		return embed
	}

	// Formatting an error message for a warning
	if w := new(commands.Warning); errors.As(err, w) {
		embed.Author = &discordgo.MessageEmbedAuthor{IconURL: "https://i.imgur.com/U30Ypyu.png", Name: "Warning"}
		embed.Description = w.Message
		embed.Color = 0xFFCC00

		return embed
	}

	// Formatting an error message for system errors
	if se := new(commands.SystemError); errors.As(err, se) {
		embed.Author = &discordgo.MessageEmbedAuthor{IconURL: "https://i.imgur.com/8lYgrbx.png", Name: "System Error"}
		embed.Description = se.Message
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Stack",
			Value: string(se.Stack),
		})

		return embed
	}

	// Formatting an error message for permission errors
	if pe := new(commands.PermissionError); errors.As(err, pe) {
		embed.Author = &discordgo.MessageEmbedAuthor{IconURL: "https://i.imgur.com/WNXPc10.png", Name: "Permission Error"}
		embed.Description = pe.Message
		return embed
	}

	// Attempting to catch other known errors
	groups := usageErrorPattern.FindStringSubmatch(err.Error())
	if groups != nil {

		// Attempting to make the message freiendlier
		m := groups[2]
		if strings.Contains(m, "valid 64 bit") {
			m = "Must be a number"
		}

		embed.Author = &discordgo.MessageEmbedAuthor{IconURL: "https://i.imgur.com/6qVd0MG.png", Name: "Usage Error"}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  groups[1],
			Value: m,
		})

		return embed
	}

	return nil
}
