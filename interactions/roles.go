package interactions

import (
	"komainu/interactions/command"
	"komainu/interactions/component"
	"komainu/interactions/delete"
	"komainu/interactions/modal"
	"komainu/interactions/response"
	"komainu/storage"
	"komainu/utility"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	delete.Register(delete.Handler{Code: DeleteRoleSelector})
	command.Register("roleselect", command.Handler{
		Description: "Create a role self-assignment message",
		Options:     createRoleOptions(),
		Code:        CommandRoleSelector,
	})
	modal.Register("roleselect", modal.Handler{Code: RoleSelectorModalHandler})
	component.Register("roleselect", component.Handler{Code: ComponentRoleSelector})

	delete.Register(delete.Handler{Code: DeleteRoleButton})
	command.Register("rolebutton", command.Handler{
		Description: "Create a message with a button that assigns a role",
		Options: []discord.CommandOption{
			&discord.RoleOption{
				OptionName:  "role",
				Description: "The role you want a button for",
			},
		},
		Code: CommandRoleButton,
	})
	modal.Register("rolebutton", modal.Handler{Code: RoleButtonModalHandler})
	component.Register("rolebutton", component.Handler{Code: ComponentRoleButton})
}

var roleFinder = regexp.MustCompile("<@&[0-9]+>")

func createRoleOptions() []discord.CommandOption {
	opt := make([]discord.CommandOption, 25)
	for i := 0; i < 25; i++ {
		opt[i] = &discord.RoleOption{
			OptionName:  "role" + strconv.Itoa(i+1),
			Description: "A preset role",
		}
	}
	return opt
}

func makeRoleMap(guild *discord.Guild) map[int64]discord.Role {
	roleMap := map[int64]discord.Role{}
	for _, role := range guild.Roles {
		roleMap[int64(role.ID)] = role
	}
	return roleMap
}

func roleCleanup(roleStrings []string) (roles []int64) {
	for _, roleIDString := range roleStrings {
		roleIDString = strings.TrimPrefix(roleIDString, "<@&")
		roleIDString = strings.TrimSuffix(roleIDString, ">")
		role, err := strconv.ParseInt(roleIDString, 10, 64)
		if err != nil {
			log.Printf("Could not convert role ID string %s into an int64: %s", roleIDString, err)
			continue
		}
		roles = append(roles, role)
	}
	return roles
}

func makeRoleButtons(roleIDs []int64, guildRoles map[int64]discord.Role) *discord.ContainerComponents {
	container := discord.ContainerComponents{}
	row := discord.ActionRowComponent{}
	for _, roleID := range roleIDs {
		if len(row) == 5 {
			rowCopy := make(discord.ActionRowComponent, 5)
			copy(rowCopy, row)
			container = append(container, &rowCopy)
			row = discord.ActionRowComponent{}
		}
		button := &discord.ButtonComponent{
			Style:    discord.PrimaryButtonStyle(),
			CustomID: discord.ComponentID("roleselect/" + strconv.FormatInt(roleID, 10)),
			Label:    guildRoles[roleID].Name,
		}
		row = append(row, button)
	}
	container = append(container, &row)
	return &container
}

// DeleteRoleSelector will delete the role selection settings if the role selction message is removed
func DeleteRoleSelector(state *state.State, kvs storage.KeyValueStore, e *gateway.MessageDeleteEvent) {
	if e.GuildID == discord.NullGuildID {
		return
	}
	err := kvs.Delete(e.GuildID, "roleselect", e.ID)
	if err != nil {
		log.Printf("[%s] Encountered an error removing role select from KVS after message deletion: %s\n", e.GuildID, err)
	}
}

// DeleteRoleButton will delete the role button settings if the role button message is removed
func DeleteRoleButton(state *state.State, kvs storage.KeyValueStore, e *gateway.MessageDeleteEvent) {
	if e.GuildID == discord.NullGuildID {
		return
	}
	err := kvs.Delete(e.GuildID, "rolebutton", e.ID)
	if err != nil {
		log.Printf("[%s] Encountered an error removing role button from KVS after message deletion: %s\n", e.GuildID, err)
	}
}

// CommandRoleSelector handles when the /roleselector command is issued
func CommandRoleSelector(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	thisGuild, err := state.Guild(event.GuildID)
	if err != nil {
		log.Printf("[%s] Could not determine current guild: %s\n", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("OK, something really weird happened, but it has been logged, so maybe it'll get fixed.")}
	}
	guildRoles := makeRoleMap(thisGuild)
	var rolesList strings.Builder
	for _, opt := range cmd.Options {
		if opt.Type == discord.RoleOptionType {
			roleID, err := strconv.ParseInt(opt.String(), 10, 64)
			if err != nil {
				log.Printf("[%s] Oh, for crying out loud! %s", event.GuildID, err)
				continue
			}
			if role, ok := guildRoles[roleID]; ok {
				rolesList.WriteString("<@&")
				rolesList.WriteString(strconv.FormatInt(roleID, 10))
				rolesList.WriteString("> (Describe \"")
				rolesList.WriteString(role.Name)
				rolesList.WriteString("\" here)\n")
			} else {
				rolesList.WriteString("(Could not find role ")
				rolesList.WriteString(opt.Value.String())
				rolesList.WriteString(")\n")
			}
		}
	}
	return command.Response{Response: modal.Respond(
		event.SenderID(), event.GuildID, "roleselect", "Describe and tag the roles",
		discord.TextInputComponent{
			CustomID:     discord.ComponentID("roles"),
			Style:        discord.TextInputParagraphStyle,
			Label:        "Message text including max 25 roles",
			LengthLimits: [2]int{1, 2000},
			Value:        option.NewNullableString(rolesList.String()),
		},
	), Callback: nil}
}

// CommandRoleButton handles when the /rolebutton command is issued
func CommandRoleButton(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {

	roleID := discord.NullRoleID

	if cmd.Options != nil && len(cmd.Options) == 1 {
		opt := cmd.Options[0]
		if opt.Type == discord.RoleOptionType {
			roleInt, err := strconv.ParseInt(opt.String(), 10, 64)
			if err != nil {
				log.Printf("[%s] Oh, for crying out loud! %s", event.GuildID, err)
				return command.Response{Response: response.Ephemeral("Something decidely weird happened while looking up that role. It was logged. Kinda.")}
			}

			thisGuild, err := state.Guild(event.GuildID)
			if err != nil {
				log.Printf("[%s] Could not determine current guild: %s\n", event.GuildID, err)
				return command.Response{Response: response.Ephemeral("OK, something really weird happened, but it has been logged, so maybe it'll get fixed.")}
			}
			guildRoles := makeRoleMap(thisGuild)

			useRole, ok := guildRoles[roleInt]
			if !ok {
				log.Printf("[%s] Non-existant role attached to /rolebutton ?!", event.GuildID)
				return command.Response{Response: response.Ephemeral("Sorry, this role is not available.")}
			}
			roleID = useRole.ID
		}
	}
	roleIDString := ""
	if roleID != discord.NullRoleID {
		roleIDString = roleID.String()
	}

	return command.Response{Response: modal.Respond(
		event.SenderID(), event.GuildID, "rolebutton", "Make a button for role assignment",
		discord.TextInputComponent{
			CustomID:     discord.ComponentID("description"),
			Style:        discord.TextInputParagraphStyle,
			Label:        "Preamble",
			Required:     false,
			LengthLimits: [2]int{0, 2000},
		},
		discord.TextInputComponent{
			CustomID:    discord.ComponentID("role"),
			Style:       discord.TextInputShortStyle,
			Label:       "Role ID",
			Required:    true,
			Value:       option.NewNullableString(roleIDString),
			Placeholder: option.NewNullableString("Enter a valid Role ID"),
		},
		discord.TextInputComponent{
			CustomID:    discord.ComponentID("buttontext"),
			Style:       discord.TextInputShortStyle,
			Label:       "Button Text",
			Required:    true,
			Value:       option.NewNullableString(""),
			Placeholder: option.NewNullableString("What should appear on the button?"),
		},
	)}
}

// RoleSelectorModalHandler handles when a modal for a Role Selector configuration is submitted
func RoleSelectorModalHandler(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, interaction *discord.ModalInteraction) command.Response {
	data := modal.DecodeModalResponse(interaction.Components)
	rawText := ""
	if val, ok := data["roles"]; ok {
		rawText = val
	}
	if rawText == "" {
		log.Printf("[%s] Empty string submitted for roles list\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("An empty text won't work for  this.")}
	}
	roleStrings := roleFinder.FindAllString(rawText, 25)
	if len(roleStrings) < 1 {
		log.Printf("[%s] Submitted text contained no regognizable role IDs\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("I'm sorry, but that text did not contain any usable roles!")}
	}
	roleIDs := roleCleanup(roleStrings)
	thisGuild, err := state.Guild(event.GuildID)
	if err != nil {
		log.Printf("[%s] WTF, I could not obtain a guild object to look up it's roles.", event.GuildID)
	}
	guildRoles := makeRoleMap(thisGuild)

	selector := storage.RoleSelector{
		Roles: map[discord.RoleID]bool{},
	}
	selector.GuildID = event.GuildID

	for _, roleID := range roleIDs {
		if role, ok := guildRoles[roleID]; ok {
			selector.Roles[role.ID] = true
		}
	}
	return command.Response{
		Response: api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content:    option.NewNullableString(rawText),
				Components: makeRoleButtons(roleIDs, guildRoles),
				AllowedMentions: &api.AllowedMentions{
					Parse: []api.AllowedMentionType{},
				},
			},
		},
		Callback: func(message *discord.Message) {
			err := selector.Store(kvs, message.ID)
			if err != nil {
				log.Printf("[%s] Created a role selector message, but failed to store it: %s", event.GuildID, err)
			}
		},
	}
}

// RoleButtonModalHandler handles the returned data from the role button creation modal
func RoleButtonModalHandler(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, interaction *discord.ModalInteraction) command.Response {
	data := modal.DecodeModalResponse(interaction.Components)
	roleButton := storage.RoleButton{
		GuildID:   event.GuildID,
		RoleID:    discord.NullRoleID,
		MessageID: discord.NullMessageID,
	}
	roleIDString, ok := data["role"]
	if !ok {
		log.Printf("[%s] Role button modal submitted with no role field?!", event.GuildID)
		return command.Response{Response: response.Ephemeral("Uhm... Something really weird happened, but it has been logged, so maybe it'll get fixed!")}
	}
	roleIDInt, err := strconv.ParseInt(roleIDString, 10, 64)
	if err != nil {
		log.Printf("[%s] Could not convert role ID string %s into an int64: %s", event.GuildID, roleIDString, err)
	}

	thisGuild, err := state.Guild(event.GuildID)
	if err != nil {
		log.Printf("[%s] Could not determine current guild: %s\n", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("Yikes! Something really weird happened, but it has been logged, so maybe it'll get fixed.")}
	}
	buttonText, hasButtonText := data["buttontext"]

	guildRoles := makeRoleMap(thisGuild)
	role, ok := guildRoles[roleIDInt]
	if !ok {
		log.Printf("[%s] Attempt to make role button for nonexistant role: <@&%d>\n", event.GuildID, roleIDInt)
		return command.Response{Response: response.Ephemeral("Ooops! I could not determine what role you were talking about, so I could not create the button.")}
	}
	if !hasButtonText {
		buttonText = role.Name
	}
	roleButton.RoleID = role.ID

	return command.Response{
		Response: api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString(data["description"]),
				Components: &discord.ContainerComponents{
					&discord.ActionRowComponent{
						&discord.ButtonComponent{
							Style:    discord.PrimaryButtonStyle(),
							CustomID: discord.ComponentID("rolebutton"), // WELCOME TO THE PEAK OF MT. INDENT!
							Label:    buttonText,
						},
					},
				},
			},
		},
		Callback: func(message *discord.Message) {
			err := roleButton.Store(kvs, message.ID)
			if err != nil {
				log.Printf("[%s] Created a role button message, but failed to store it: %s", event.GuildID, err)
			}
		},
	}

}

// ComponentRoleButton handles interactions from a role button
func ComponentRoleButton(state *state.State, kvs storage.KeyValueStore, e *gateway.InteractionCreateEvent, interaction discord.ComponentInteraction) api.InteractionResponse {
	exist, roleID, err := storage.GetRoleForButton(kvs, e.GuildID, e.Message.ID)
	if err != nil {
		log.Printf("[%s] Error while trying to fetch the RoleButton while processing a role request:  %s", e.GuildID, err)
		return response.Ephemeral("There was an issue processing your role button press. It has been logged.")
	}
	if !exist {
		log.Printf("[%s] RoleButton does not exist while attempting to process role button press", e.GuildID)
		return response.Ephemeral("I'm very sorry, but I couldn't authenticate the role button press.")
	}

	guild, err := state.Guild(e.GuildID)
	if err != nil {
		log.Printf("[%s] Could not determine relevant guild during button-triggered role request.", e.GuildID)
		return response.Ephemeral("That was kind of a strange button. Nothing happened!")
	}

	roleMap := makeRoleMap(guild)
	_, ok := roleMap[int64(roleID)]
	if !ok {
		log.Printf("[%s] RoleButton request for non-existant role %d", e.GuildID, roleID)
		return response.Ephemeral("That was kind of an odd role request. Maybe that role disappeared?!")
	}

	member, err := state.Member(e.GuildID, e.SenderID())
	if err != nil {
		log.Printf("[%s] Could not get Sender's member object for button-triggered role request", e.GuildID)
		return response.Ephemeral("Wait, wait, wait! Are you *sure* you're a member?!")
	}
	if utility.ContainsRole(member.RoleIDs, roleID) {
		return response.Ephemeral("You already have the role. Enjoy!")
	} else {
		err := state.AddRole(e.GuildID, member.User.ID, roleID, api.AddRoleData{
			AuditLogReason: "Member got the role using a RoleButton",
		})
		if err != nil {
			log.Printf("[%s] Error while trying to grant a role from a RoleButton: %s", e.GuildID, err)
			return response.Ephemeral("Ooops! I'm sorry, but I couldn't grant the role. The error has been logged.")
		}
		return response.Ephemeral("You now have the", roleID.Mention(), "role")
	}

}

// ComponentRoleSelector handkes intractions from the role selector button components
func ComponentRoleSelector(state *state.State, kvs storage.KeyValueStore, e *gateway.InteractionCreateEvent, interaction discord.ComponentInteraction) api.InteractionResponse {
	exist, selector, err := storage.GetRoleSelector(kvs, e.GuildID, e.Message.ID)
	if err != nil {
		log.Printf("[%s] Error while trying to fetch the RoleSelector while processing a role request:  %s", e.GuildID, err)
		return response.Ephemeral("There was an issue processing your role request. It has been logged.")
	}
	if !exist {
		log.Printf("[%s] RoleSelector does not exist while attempting to process role request", e.GuildID)
		return response.Ephemeral("I'm very sorry, but I couldn't authenticate the role request.")
	}

	selectionString := strings.TrimPrefix(string(interaction.ID()), "roleselect/")

	roleID, err := strconv.ParseInt(selectionString, 10, 64)
	if err != nil {
		log.Printf("[%s] Malformed role request: %s", e.GuildID, selectionString)
		return response.Ephemeral("That was kind of a malformed role request. What happened?")
	}

	guild, err := state.Guild(e.GuildID)
	if err != nil {
		log.Printf("[%s] Could not determine relevant guild during role request.", e.GuildID)
		return response.Ephemeral("That was kind of a strange role request. What happened?")
	}

	roleMap := makeRoleMap(guild)
	role, ok := roleMap[roleID]
	if !ok {
		log.Printf("[%s] Role request for non-existant role %d", e.GuildID, roleID)
		return response.Ephemeral("That was kind of an odd role request. What happened?")
	}

	if !selector.Has(role.ID) {
		log.Printf("[%s] Role request for a role not listed for that selector", e.GuildID)
		return response.Ephemeral("That was a curious role request. What happened?")
	}

	member, err := state.Member(e.GuildID, e.SenderID())
	if err != nil {
		log.Printf("[%s] Could not get Sender's member object for role request", e.GuildID)
		return response.Ephemeral("That was kind of an odd role request. What happened?")
	}

	if utility.ContainsRole(member.RoleIDs, role.ID) {
		err := state.RemoveRole(e.GuildID, member.User.ID, role.ID, api.AuditLogReason("Member removed role using a RoleSelector button"))
		if err != nil {
			log.Printf("[%s] Error while trying to revoke a role from a RoleSelector: %s", e.GuildID, err)
			return response.Ephemeral("Ooops! I'm sorry, but I couldn't remove the role. The error has been logged.")
		}
		return response.Ephemeral("As requested, I have removed the", role.ID.Mention(), "role from you.")
	} else {
		err := state.AddRole(e.GuildID, member.User.ID, role.ID, api.AddRoleData{
			AuditLogReason: "Member got the role using a RoleSelector button",
		})
		if err != nil {
			log.Printf("[%s] Error while trying to grant a role from a RoleSelector: %s", e.GuildID, err)
			return response.Ephemeral("Ooops! I'm sorry, but I couldn't grant the role. The error has been logged.")
		}
		return response.Ephemeral("As requested, you now have the", role.ID.Mention(), "role")
	}
}
