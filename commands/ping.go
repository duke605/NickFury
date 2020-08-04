package commands

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// Ping ...
type Ping struct{}

// Run ...
func (p Ping) Run(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	alloc := p.byteCountDecimal(memStats.Alloc)
	totalAlloc := p.byteCountDecimal(memStats.TotalAlloc)
	uptime := func() string {
		d := time.Now().Sub(viper.Get("start").(time.Time))
		parts := []string{}

		h, d := d/time.Hour, d%time.Hour
		m, d := d/time.Minute, d%time.Minute
		s, d := d/time.Second, d%time.Second
		ms := d / time.Millisecond

		if h > 0 {
			parts = []string{
				fmt.Sprintf("%dh", h),
				fmt.Sprintf("%dm", m),
				fmt.Sprintf("%ds", s),
			}
		} else if m > 0 {
			parts = []string{
				fmt.Sprintf("%dm", m),
				fmt.Sprintf("%ds", s),
			}
		} else if s > 0 {
			parts = []string{
				fmt.Sprintf("%ds", s),
				fmt.Sprintf("%dms", ms),
			}
		} else {
			parts = []string{
				fmt.Sprintf("%dms", ms),
			}
		}

		return strings.Join(parts, " ")
	}()

	// Getting size of datastore
	dsSize := "ERROR!"
	if stat, err := os.Stat(".data"); err == nil {
		dsSize = p.byteCountDecimal(uint64(stat.Size()))
	}

	sess.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Pong!",
			IconURL: "https://i.imgur.com/kfVOddd.png",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/3hb4SrY.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "# Goroutines",
				Value: fmt.Sprint(runtime.NumGoroutine()),
			},
			{
				Name:  "# CPUs",
				Value: fmt.Sprint(runtime.NumCPU()),
			},
			{
				Name:  "Datastore Size",
				Value: fmt.Sprintf("%s", dsSize),
			},
			{
				Name:  "Memory Usage",
				Value: fmt.Sprintf("%s/%s", alloc, totalAlloc),
			},
			{
				Name:  "Latency",
				Value: fmt.Sprintf("%dms", sess.HeartbeatLatency().Milliseconds()),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Uptime: %s", uptime),
		},
	})
	return nil
}

func (Ping) byteCountDecimal(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
