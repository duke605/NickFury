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

// ComposeEmbed creates an embed showing who is linked to what route in the chaennl
func (s *Service) ComposeEmbed(idx map[string][]string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Routes",
			IconURL: "https://cdn0.iconfinder.com/data/icons/small-n-flat/24/678111-map-marker-512.png",
		},
		Color: 0x99B2DD,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    "https://i.imgur.com/KHHO0DY.png",
			Height: 1000,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Section 1",
				Value: s.ComposeSectionText(idx, 1) + "\n" + string(0x200B),
			},
			{
				Name:  "Section 2",
				Value: s.ComposeSectionText(idx, 2) + "\n" + string(0x200B),
			},
			{
				Name:  "Section 3",
				Value: s.ComposeSectionText(idx, 3),
			},
		},
	}
}

// ComposeSectionText creates a string for an embed field showing who is linked to a section
func (s *Service) ComposeSectionText(idx map[string][]string, section int) string {
	limit := 4 + section
	alpha := "ABCDEFG"

	str := ""
	for i := 0; i < limit; i++ {
		list := ""
		char := string(alpha[i])

		// Getting persons assigned to current route
		key := fmt.Sprintf("%d:%s", section, char)
		if ids, ok := idx[key]; ok {
			mentions := make([]string, len(ids))
			for i, id := range ids {
				mentions[i] = fmt.Sprintf("<@!%s>", id)
			}

			list = strings.Join(mentions, "/")
		}

		str += strings.Trim(fmt.Sprintf("**%s:** %s", char, list), "")
		str += "\n"
	}

	str = strings.Trim(str, " \n")
	return str
}
