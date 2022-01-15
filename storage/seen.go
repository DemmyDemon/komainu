package storage

import (
	"fmt"
	"komainu/utility"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

// See saves the given user as being seen in the given guild.
func See(kvs KeyValueStore, guildID discord.GuildID, userID discord.UserID) error {
	return kvs.Set(guildID, "seen", userID, time.Now().Unix())
}

// LastSeen checks to see when the given user was seen in the given guild.
func LastSeen(kvs KeyValueStore, guildID discord.GuildID, userID discord.UserID) (bool, int64, error) {
	return kvs.GetInt64(guildID, "seen", userID)
}

func MaybeGiveActiveRole(kvs KeyValueStore, state *state.State, guildID discord.GuildID, member *discord.Member) (err error) {
	role := discord.NullRoleID
	exist, err := kvs.GetObject(guildID, "activerole", "role", &role)
	if err != nil {
		return fmt.Errorf("MaybeGiveActiveRole GetObject: %w", err)
	}

	if exist {
		if !utility.ContainsRole(member.RoleIDs, role) {
			return state.AddRole(guildID, member.User.ID, role, api.AddRoleData{
				AuditLogReason: "Role automatically granted for chat activity.",
			})
		}
	}

	return nil
}

func RemoveActiveRole(kvs KeyValueStore, state *state.State, guildID discord.GuildID, member *discord.Member) error {
	role := discord.NullRoleID
	exist, err := kvs.GetObject(guildID, "activerole", "role", &role)
	if err != nil {
		return fmt.Errorf("RemoveActiveRole GetObject: %w", err)
	}

	if exist {
		if utility.ContainsRole(member.RoleIDs, role) {
			return state.RemoveRole(guildID, member.User.ID, role, api.AuditLogReason("Role automatically revoked for chat inactivity."))
		}
	}

	return nil
}

func RevokeActiveRoles(state *state.State, kvs KeyValueStore) error {
	guilds, err := state.Guilds()
	if err != nil {
		return fmt.Errorf("revoking active roles failed to get guilds slice: %w", err)
	}
	now := time.Now().Unix()
	secondsInDay := int64(24 * 60 * 60)
	for _, guild := range guilds {
		role := discord.NullRoleID
		exist, err := kvs.GetObject(guild.ID, "activerole", "role", &role)
		if err != nil {
			log.Printf("[%s] Failed to fetch the active role object: %s\n", guild.ID, err)
		}
		if !exist {
			continue // Because if the role isn't set, this guild has no "active role"
		}

		exist, days, err := kvs.GetInt64(guild.ID, "activerole", "days")
		if err != nil {
			log.Printf("[%s] Failed to fetch the active role time: %s\n", guild.ID, err)
		}
		if !exist || days < 1 {
			continue
		}
		inactiveIfSeenBefore := now - (days * secondsInDay)

		members, err := state.Session.Members(guild.ID, 0)
		if err != nil {
			log.Printf("[%s] Failed to fetch the member list: %s\n", guild.ID, err)
			continue
		}
		for _, member := range members {
			wasSeen, when, err := LastSeen(kvs, guild.ID, member.User.ID)
			if err != nil {
				log.Printf("[%s] Failed to fetch seen data for member: %s", guild.ID, err)
			}
			if !wasSeen || when < inactiveIfSeenBefore {
				if utility.ContainsRole(member.RoleIDs, role) {
					err := state.RemoveRole(guild.ID, member.User.ID, role, api.AuditLogReason("Role automatically revoked for chat inactivity."))
					if err != nil {
						log.Printf("[%s] Failed to remove role from inactive user: %s", guild.ID, err)
					}
				}
			}
		}

	}
	return nil
}

func StartRevokingActiveRole(state *state.State, kvs KeyValueStore) {
	ticker := time.NewTicker(10 * time.Minute)
	for {
		<-ticker.C
		if err := RevokeActiveRoles(state, kvs); err != nil {
			log.Printf("Error encountered trying to revoke active roles: %s\n", err)
		}
	}
}
