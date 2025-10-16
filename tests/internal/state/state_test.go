package state_test

import (
	"testing"

	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/state"
)

func TestSetAndCheckSlaveFresh(t *testing.T) {
	n := &global.Node{Name: "n1"}
	state.SetSlaveAsFresh(n)
	if !state.IsSlaveFresh("n1") {
		t.Fatalf("expected slave n1 to be fresh")
	}
}

func TestMarkSlaveSyncing(t *testing.T) {
	state.MarkSlaveSyncing("n2", true)
	if !state.IsSlaveSyncing("n2") {
		t.Fatalf("expected slave n2 to be syncing")
	}
	state.MarkSlaveSyncing("n2", false)
	if state.IsSlaveSyncing("n2") {
		t.Fatalf("expected slave n2 to not be syncing")
	}
}

func TestMarkAllSlavesDirty(t *testing.T) {
	n := &global.Node{Name: "n3"}
	state.SetSlaveAsFresh(n)
	state.MarkAllSlavesDirty()
	if state.IsSlaveFresh("n3") {
		t.Fatalf("expected slave n3 to be dirty")
	}
}
