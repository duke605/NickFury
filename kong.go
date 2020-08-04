package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// KongError ...
type kongError string

func (e kongError) Error() string {
	return string(e)
}

const (
	// ErrHandled is returned when everything worked but the execution
	// should go no futher
	ErrHandled kongError = "Handled"
)

// createHelpPrinter creates a function that will capture the help message and send
// it to discord instead of os.Stdout
func createHelpPrinter(sess *discordgo.Session, m *discordgo.MessageCreate) kong.HelpPrinter {
	return func(opts kong.HelpOptions, kctx *kong.Context) error {
		cmdPrefix := viper.GetString("COMMAND_PREFIX")
		usage := fmt.Sprintf("Usage: %s%s", cmdPrefix, kctx.Selected().Summary())

		// Replacing the kong instance's stdout with a buffer so we can capture the data
		buf := bytes.Buffer{}
		kctx.Stdout = &buf

		// Defering to default writer
		if err := kong.DefaultHelpPrinter(opts, kctx); err != nil {
			return nil
		}
		kctx.Stdout = os.Stdout

		// Replacing the first line of the help message with the corrected usage
		parts := strings.Split(buf.String(), "\n")
		parts[0] = usage
		msg := "```" + strings.Join(parts, "\n") + "```"

		// Sending help message to discord
		sess.ChannelMessageSend(m.ChannelID, msg)
		return ErrHandled
	}
}

// getActualError unwraps the provided error if it is a parse error and returns
// the actual underlying error. If the error is not a parse error err is returned
// as is
func getActualError(err error) error {
	e, ok := err.(*kong.ParseError)
	if ok {
		return e.Cause()
	}

	return err
}
