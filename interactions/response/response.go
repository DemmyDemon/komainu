package response

import (
	"io"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

// Ephemeral generates an emphemeral response message from the strings given.
func Ephemeral(message ...string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(strings.Join(message, " ")),
			Flags:   api.EphemeralResponse,
		},
	}
}

// Message generates an InteractionResponse from the strings given.
func Message(message ...string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(strings.Join(message, " ")),
		},
	}
}

// MessageNoMention generates an InteractionResponse from the strings given, and suppresses any mentions this might cause.
func MessageNoMention(message ...string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(strings.Join(message, " ")),
			AllowedMentions: &api.AllowedMentions{
				Parse: []api.AllowedMentionType{},
			},
		},
	}
}

// MessageAttachFile generates an InteractionResponse from the strings given, and attaches the given file.
func MessageAttachFile(message string, name string, reader io.Reader) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(message),
			Files: []sendpart.File{
				{
					Name:   name,
					Reader: reader,
				},
			},
		},
	}
}
