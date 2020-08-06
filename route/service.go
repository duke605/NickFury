package route

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Service ...
type Service struct {
	*Repository
}

// NewService creates a new Service
func NewService(repo *Repository) *Service {
	return &Service{
		Repository: repo,
	}
}

// ComposeEmbed creates an embed showing who is linked to what route in the channel
func (s *Service) ComposeEmbed(m Map, idx map[string][]string) *discordgo.MessageEmbed {

	// Creating fields
	fields := make([]*discordgo.MessageEmbedField, m.Sections)
	for i := 0; i < int(m.Sections); i++ {
		suffix := ""
		if i != int(m.Sections-1) {
			suffix = string(0x200B)
		}

		fields[i] = &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("__Section %d__", i+1),
			Value: s.ComposeSectionText(m, idx, i+1) + "\n" + suffix,
		}
	}

	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Routes",
			IconURL: "https://cdn0.iconfinder.com/data/icons/small-n-flat/24/678111-map-marker-512.png",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    "https://i.imgur.com/KHHO0DY.png",
			Height: 1000,
		},
		Color:  0x99B2DD,
		Fields: fields,
	}
}

// ComposeSectionText creates a string for an embed field showing who is linked to a section
func (s *Service) ComposeSectionText(m Map, idx map[string][]string, section int) string {
	paths := m.Paths(section)

	str := ""
	for i := 0; i < len(paths); i++ {
		list := ""

		// Getting persons assigned to current route
		key := fmt.Sprintf("%d:%s", section, paths[i])
		if ids, ok := idx[key]; ok {
			mentions := make([]string, len(ids))
			for i, id := range ids {
				mentions[i] = fmt.Sprintf("<@!%s>", id)
			}

			list = strings.Join(mentions, "/")
		}

		str += strings.Trim(fmt.Sprintf("**%s:** %s", paths[i], list), " ")
		str += "\n"
	}

	str = strings.Trim(str, " \n")
	return str
}
