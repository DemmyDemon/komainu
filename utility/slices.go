package utility

import "github.com/diamondburned/arikawa/v3/discord"

// RoleInCommon checks if one slice of RoleIDs has an element in common with the other.
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

// ContainsString checks if the haystack of strings has the needle string.
func ContainsString(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}

// ContainsRole checks if the haystack of RoleIDs has the needle RoleID
func ContainsRole(haystack []discord.RoleID, needle discord.RoleID) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}

// ContainsUser checks if the haystack of UserIDs contains the needle UserID
func ContainsUser(haystack []discord.UserID, needle discord.UserID) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}
