package modal

import (
	"komainu/interactions/command"
	"komainu/interactions/response"
	"komainu/storage"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"

	"github.com/google/uuid"
)

type Handler struct {
	Code HandlerFunction
}

type HandlerFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.InteractionCreateEvent,
	interaction *discord.ModalInteraction,
) command.Response

type Secret struct {
	Handler string
	User    discord.UserID
	Guild   discord.GuildID
	Created time.Time
}

var modalMaxAge time.Duration = time.Minute * 15

// modals holds the modal handlers to accept.
var modals = map[string]Handler{}

var modalSecrets = map[string]Secret{}

func init() {
	go startRemovingStaleSecrets()
}

// AddHandler adds the modal interactin handler to the given state
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {
		if interaction, ok := e.Data.(*discord.ModalInteraction); ok {
			id := string(interaction.CustomID)
			if secret, exist := modalSecrets[id]; exist {
				if secret.User != e.SenderID() {
					log.Printf("[%s] Modal form submission from WRONG USER: %s, but expected %s", e.GuildID, e.SenderID(), secret.User)
				}
				if val, ok := modals[secret.Handler]; ok {
					response := val.Code(state, kvs, e, interaction)
					if err := state.RespondInteraction(e.ID, e.Token, response.Response); err != nil {
						log.Printf("[%s] Failed to send modal interaction response: %s", e.GuildID, err)
					}
				} else {
					log.Printf("[%s] has UNKNOWN modal interaction %#v", e.GuildID, secret)
				}
				delete(modalSecrets, id)
			} else {
				log.Printf("[%s] expired/invalid modal token %s used by %s\n", e.GuildID, id, e.SenderID())
				if err := state.RespondInteraction(e.ID, e.Token, response.Ephemeral("Sorry, access was denied. Took too long to respond?")); err != nil {
					log.Printf("[%s] ...and there was an error telling them their token expired: %s", e.GuildID, err)
				}
				return
			}
		}
	})
}

func startRemovingStaleSecrets() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		now := time.Now()
		for key, secret := range modalSecrets {
			if now.Sub(secret.Created) > modalMaxAge {
				delete(modalSecrets, key)
			}
		}
	}
}

func Register(name string, handler Handler) {
	modals[name] = handler
}

// DecodeModalResponse tries to cram a Modal dialog response into a string/string map
func DecodeModalResponse(components discord.ContainerComponents) map[string]string {
	response := make(map[string]string)
	for _, component := range components {
		switch cmp := component.(type) {
		case *discord.ActionRowComponent:
			decodeActionRowComponent(*cmp, response)
		// Are there other possible types?
		default:
			log.Printf("Unknown type encountered in DecodeModalResponse: %T", component)
		}
	}
	return response
}

// decodeActionRowComponent tries to figure out the key and text for the given ActionRowComponent
func decodeActionRowComponent(arc discord.ActionRowComponent, response map[string]string) {
	for _, arcElement := range arc {
		switch comp := arcElement.(type) {
		case *discord.TextInputComponent:
			if comp.Value != nil && comp.Value.Init {
				response[string(comp.CustomID)] = string(comp.Value.Val)
			}
		// I don't think there are any other possible types. Yet.
		default:
			log.Printf("Unknown type encountered in decodeActionRowComponent: %T", comp)
		}
	}
}

func Respond(user discord.UserID, guild discord.GuildID, name string, title string, tics ...discord.TextInputComponent) api.InteractionResponse {
	id := uuid.New().String()
	modalSecrets[id] = Secret{
		Handler: name,
		User:    user,
		Guild:   guild,
		Created: time.Now(),
	}
	return api.InteractionResponse{
		Type: api.ModalResponse,
		Data: &api.InteractionResponseData{
			Title:      option.NewNullableString(title),
			CustomID:   option.NewNullableString(id),
			Components: generateModalComponents(tics),
		},
	}
}

func generateModalComponents(tics []discord.TextInputComponent) *discord.ContainerComponents {
	container := make(discord.ContainerComponents, len(tics))
	for i := 0; i < len(tics); i++ {
		container[i] = &discord.ActionRowComponent{&tics[i]}
	}
	return &container
}
