package main

import (
	"strings"
	"testing"

	"github.com/sshreyx1/cyberditto/services/collector/internal/collect"
	"github.com/sshreyx1/cyberditto/services/collector/internal/pseudonymise"
	"github.com/sshreyx1/cyberditto/services/collector/internal/snapshot"
)

func TestPseudonymiseSnapshot_HostAndCreatedAccountsOnly(t *testing.T) {
	snap := snapshot.Snapshot{
		Hostname: "ALICE-LAPTOP",
		Fields: map[string]snapshot.Field{
			"accounts.local_users": {Value: []collect.LocalUser{
				{Name: "Administrator", RID: 500},
				{Name: "Guest", RID: 501},
				{Name: "alice", RID: 1001, Enabled: true},
			}},
		},
	}
	if err := pseudonymiseSnapshot(&snap, pseudonymise.NewMapper()); err != nil {
		t.Fatalf("pseudonymiseSnapshot: %v", err)
	}
	if !strings.HasPrefix(snap.Hostname, "HOST-") {
		t.Errorf("hostname = %q, want a HOST- placeholder", snap.Hostname)
	}
	users := snap.Fields["accounts.local_users"].Value.([]collect.LocalUser)
	if users[0].Name != "Administrator" || users[1].Name != "Guest" {
		t.Errorf("built-in account names must be kept: %+v", users[:2])
	}
	if !strings.HasPrefix(users[2].Name, "USER-") || !users[2].Enabled {
		t.Errorf("created account = %+v, want a USER- placeholder with flags intact", users[2])
	}
}
