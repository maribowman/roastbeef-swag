package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDiscordConfig(t *testing.T) {
	tests := map[string]struct {
		config  DiscordConfig
		wantErr bool
	}{
		"valid": {
			config:  DiscordConfig{Token: "some-token", BotID: "1234"},
			wantErr: false,
		},
		"empty token": {
			config:  DiscordConfig{Token: "", BotID: "1234"},
			wantErr: true,
		},
		"empty ID": {
			config:  DiscordConfig{Token: "some-token", BotID: ""},
			wantErr: true,
		},
		"placeholder token": {
			config:  DiscordConfig{Token: "BOT_TOKEN", BotID: "1234"},
			wantErr: true,
		},
		"placeholder ID": {
			config:  DiscordConfig{Token: "some-token", BotID: "BOT_ID"},
			wantErr: true,
		},
		"both bad": {
			config:  DiscordConfig{Token: "BOT_TOKEN", BotID: ""},
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// given
			// when
			err := validateDiscordConfig(test.config)

			// then
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
