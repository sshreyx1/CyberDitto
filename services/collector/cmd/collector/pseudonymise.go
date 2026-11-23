package main

import (
	"github.com/sshreyx1/cyberditto/services/collector/internal/collect"
	"github.com/sshreyx1/cyberditto/services/collector/internal/pseudonymise"
	"github.com/sshreyx1/cyberditto/services/collector/internal/snapshot"
)

// firstUserRID: RIDs below 1000 are Windows' built-in accounts
// (Administrator, Guest, DefaultAccount, ...), whose names identify no one;
// accounts somebody created start at 1000.
const firstUserRID = 1000

// pseudonymiseSnapshot replaces the hostname and the names of non-built-in
// local accounts with placeholders from m, in place.
func pseudonymiseSnapshot(snap *snapshot.Snapshot, m *pseudonymise.Mapper) error {
	host, err := m.Replace("HOST", snap.Hostname)
	if err != nil {
		return err
	}
	snap.Hostname = host

	field, ok := snap.Fields["accounts.local_users"]
	if !ok {
		return nil
	}
	users, ok := field.Value.([]collect.LocalUser)
	if !ok {
		return nil
	}
	out := make([]collect.LocalUser, len(users))
	for i, u := range users {
		if u.RID >= firstUserRID {
			name, err := m.Replace("USER", u.Name)
			if err != nil {
				return err
			}
			u.Name = name
		}
		out[i] = u
	}
	field.Value = out
	snap.Fields["accounts.local_users"] = field
	return nil
}
