package interactions

import (
	"fmt"
	"komainu/interactions/command"
	"komainu/interactions/response"
	"komainu/storage"
	"log"
	"math/rand"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func init() {
	command.Register("ateball", commandAteHandler)
}

var commandAteHandler = command.Handler{
	Description: "Ask the mystical ateball!",
	Code:        CommandAte,
	Options: []discord.CommandOption{
		&discord.StringOption{
			OptionName:  "question",
			Description: "What would you like the mystical ateball to advice you on?",
			Required:    true,
		},
	},
}

var responses = []string{
	"It is all crust.",                  // "It is certain.",
	"It is deliciously so.",             // "It is decidedly so.",
	"Without a hotdog.",                 // "Without a doubt.",
	"Outlook scrumcious.",               // "Outlook good.",
	"It's on the menu.",                 // "Yes.",
	"Signs point to yeast.",             // "Signs point to yes.",
	"Reply overcooked, try again.",      // "Reply hazy, try again.",
	"Eat again later.",                  // "Ask again later.",
	"Butter not tell you now.",          // "Better not tell you now.",
	"Canneloni predict now.",            // "Cannot predict now.",
	"Thicken the sauce and ask again",   // "Concentrate and ask again.",
	"Don't fast on it.",                 // "Don't count on it.",
	"My sauces say no.",                 // "My sources say no.",
	"Very doughful.",                    // "Very doubtful.",
	"Yes - deliciously!",                // "Yes – definitely.",
	"You may serve it.",                 // "You may rely on it.",
	"As I see it, yummy!",               // "As I see it, yes.",
	"Five stars!",                       // "Most likely.",
	"Dog food, if you don't like dogs.", // "My reply is no.",
	"Slightly burnt.",                   // "Outlook not so good.",
}

var foodEmoji = []rune{
	'🍇', '🍈', '🍉', '🍊', '🍋', '🍌', '🍍', '🥭', '🍎', '🍏', '🍐', '🍑', '🍒', '🍓', '🥝', '🍅', '🥥',
	'🥑', '🍆', '🥔', '🥕', '🌽', '🌶', '🥒', '🥬', '🥦', '🍄', '🥜', '🌰',
	'🍞', '🥐', '🥖', '🥨', '🥯', '🥞', '🧀', '🍖', '🍗', '🥩', '🥓',
	'🍔', '🍟', '🍕', '🌭', '🥪', '🌮', '🌯', '🥙', '🥚', '🍳', '🥘', '🍲', '🥣', '🥗', '🍿', '🧂', '🥫',
	'🍱', '🍘', '🍙', '🍚', '🍛', '🍜', '🍝', '🍠', '🍢', '🍣', '🍤', '🍥', '🥮', '🍡', '🥟', '🥠', '🥡',
	'🦀', '🦞', '🦐', '🦑',
	'🍦', '🍧', '🍨',
	'🍩', '🍪', '🎂', '🍰', '🧁', '🥧', '🍫', '🍬', '🍭', '🍮', '🍯',
	'🍼', '🥛', '☕', '🍵', '🍶', '🍾', '🍷', '🍸', '🍹', '🍺', '🍻', '🥂', '🥃', '🥤',
	'🥢', '🍽', '🍴', '🥄', '🔪', '🏺',
}

func CommandAte(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {

	if cmd.Options == nil || len(cmd.Options) != 1 || cmd.Options[0].String() == "" {
		log.Printf("[%s] /ateball command structure somehow did not include the question portion. Wat.\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("You forgot your question!")}
	}
	question := cmd.Options[0].String()
	rand.Seed(time.Now().Unix())
	respondWith := responses[rand.Intn(len(responses))]
	food := foodEmoji[rand.Intn(len(foodEmoji))]
	return command.Response{Response: response.MessageNoMention(fmt.Sprintf("> %s\n%c %s", question, food, respondWith))}
}
