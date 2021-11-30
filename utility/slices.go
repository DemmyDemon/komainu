package utility

import "github.com/diamondburned/arikawa/v3/discord"

func RoleInCommon(one []discord.RoleID, other []discord.RoleID) bool {
	for _, oneCandidate := range one {
		for _, otherConeCandidate := range other {
			if oneCandidate == otherConeCandidate {
				return true
			}
		}
	}
	return false
}

func ContainsString(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}

func ContainsRole(haystack []discord.RoleID, needle discord.RoleID) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}
