package commands

import (
	"errors"
	"regexp"
)

var mentionPattern = regexp.MustCompile(`(?:^<@\!?(\d+)>$|^(\d+)$)`)

// Mention is a argument type that can either be a mention or an id
type Mention string

// UnmarshalText ...
func (m *Mention) UnmarshalText(b []byte) error {
	text := string(b)
	groups := mentionPattern.FindStringSubmatch(text)
	if groups == nil {
		return errors.New("Must be a mention (@someone) or a raw ID")
	}

	// Text was a mention
	if groups[1] != "" {
		*m = Mention(groups[1])
		return nil
	}

	// Text was raw ID
	*m = Mention(groups[2])
	return nil
}
