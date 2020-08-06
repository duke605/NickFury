package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/boltdb/bolt"
	"github.com/duke605/NickFury/commands"
	"github.com/duke605/NickFury/route"
	"github.com/google/shlex"
	_ "github.com/joho/godotenv/autoload"

	"github.com/spf13/viper"

	dg "github.com/bwmarrin/discordgo"
)

var (
	bot *dg.Session
	db  *bolt.DB

	routeService *route.Service
)

func init() {
	var err error
	viper.AutomaticEnv()

	// Initializing bot
	bot, err = dg.New("Bot " + viper.GetString("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	bot.State.TrackEmojis = false
	bot.State.TrackPresences = false
	bot.State.TrackVoice = false
	bot.State.TrackChannels = false
	bot.SyncEvents = false
	bot.Identify.Intents = dg.MakeIntent(dg.IntentsGuildMessages)

	bot.AddHandler(onMessage)
	bot.AddHandlerOnce(onReady)

	// Creating repos
	db, err = bolt.Open(".data", 0600, nil)
	if err != nil {
		panic(err)
	}
	routeRepo := route.NewRepo(db)

	// Creating services
	routeService = route.NewService(routeRepo)
}

func main() {
	if err := bot.Open(); err != nil {
		panic(err)
	}

	// Waiting for kill command
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("--- Terminating ---")

	go func() {
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}()

	bot.Close()
	db.Close()
}

// onReady is called when the bot has successfully connected to discord
func onReady(sess *dg.Session, evt *dg.Ready) {
	viper.Set("start", time.Now())
	fmt.Println("--- Press Ctrl + C to stop ---")
}

// onMessage handles messages sent from discord
func onMessage(sess *dg.Session, msg *dg.MessageCreate) {
	defer func() {
		if perr := recover(); perr != nil {
			fmt.Println("Recovered from panic: ", perr)
		}
	}()

	cmdPrefix := viper.GetString("COMMAND_PREFIX")
	start := time.Now()

	// Ruling out messages that are not commands
	if msg.Author.ID == sess.State.User.ID ||
		msg.Author.Bot ||
		!strings.HasPrefix(msg.Content, cmdPrefix) {
		return
	}

	// Splitting message into command parts and leaving out the prefix
	parts, err := shlex.Split(msg.Content[1:])
	if err != nil {
		return
	}

	// Creating parser
	parser, err := kong.New(&commands.Root{},
		kong.Exit(func(int) {}),
		kong.Help(createHelpPrinter(sess, msg)),

		// Binding all the things
		kong.Bind(sess),
		kong.Bind(msg),
		kong.Bind(routeService),
		kong.Bind(start),
	)
	if err != nil {
		panic(err)
	}

	// Trying to find a command to execute with the message content
	cmd, err := parser.Parse(parts)
	if err != nil {
		if embed := errorToEmbed(err); embed != nil {
			sess.ChannelMessageSendEmbed(msg.ChannelID, embed)
		} else {
			fmt.Println(err)
		}

		return
	}

	// Executing the command
	err = cmd.Run()
	if err != nil {
		if embed := errorToEmbed(err); embed != nil {
			sess.ChannelMessageSendEmbed(msg.ChannelID, embed)
		} else {
			fmt.Println(err)
		}

		return
	}
}
