//go:build ignore

// This needs so much more work.
package storage

import (
	"sort"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// TODO: Refactor the whole damn thing to take a channel as the storage key.
// This will let us have more than one set of these per guild.
// It's either that, or make the rules acceptance thing entirely separate.

type RoleDescription struct {
	Description string
	Name        string
	Icon        string
	Color       discord.Color
	ID          discord.RoleID
}

type RoleSelector struct {
	Roles   map[discord.RoleID]RoleDescription
	Message discord.MessageID
}

func (rs *RoleSelector) Store(kvs KeyValueStore, guildID discord.GuildID) error {
	return kvs.Set(guildID, "roles", "selector", rs)
}

func (rs *RoleSelector) List() (roles []RoleDescription) {
	for _, role := range rs.Roles {
		roles = append(roles, role)
	}

	sort.SliceStable(roles, func(i int, j int) bool {
		return roles[i].Name > roles[j].Name
	})

	return

}

func (rs *RoleSelector) RefreshNames(guild *discord.Guild) {
	if guild == nil {
		return
	}

	for _, role := range guild.Roles {
		if desc, ok := rs.Roles[role.ID]; ok {
			desc.Name = role.Name
			desc.Icon = role.Icon
			desc.Color = role.Color
		}
	}
}

func (rs *RoleSelector) RolesMap(guild *discord.Guild) (roles map[discord.RoleID]discord.Role) {
	for _, role := range guild.Roles {
		if rs.Has(role.ID) {
			roles[role.ID] = role
		}
	}
	return
}

func (rs *RoleSelector) Has(roleID discord.RoleID) bool {
	if _, ok := rs.Roles[roleID]; ok {
		return true
	}
	return false
}

func (rs *RoleSelector) String() string {
	var sb strings.Builder
	for _, desc := range rs.List() {
		sb.WriteString("<@")
		sb.WriteString(desc.ID.String())
		sb.WriteString("> - ")
		sb.WriteString(desc.Description)
		sb.WriteRune('\n')
	}
	return sb.String()
}

func GetRoleSelector(kvs KeyValueStore, guildID discord.GuildID) (exist bool, selector RoleSelector, err error) {
	exist, err = kvs.Get(guildID, "roles", "selector", &selector)
	return
}

// HandleInteractionAsRoleSelect determines if the given interaction is a vote button click, and acts accordingly.
func HandleInteractionAsRoleSelect(state *state.State, kvs KeyValueStore, e *gateway.InteractionCreateEvent, interaction discord.ComponentInteraction) (hasRoleSelector bool, response string, err error) {
	exist, selector, err := GetRoleSelector(kvs, e.GuildID)

	if err != nil {
		// TODO: Report error
		return
	}
	if !exist {
		// TODO: Reort that someone tried to role select in a guild with no role selector
		return
	}
}
