//go:build windows

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBoardCleanupBoundaries(t *testing.T) {
	at := time.Date(2026, 10, 12, 12, 0, 0, 0, time.Local)
	mk := func(id, body, fw string, sh, p uint32, minute int) boardRecord {
		return boardRecord{ID: id, Body: body, Model: "DC-S5", Firmware: fw, Verified: true, Shutter: sh, Power: p, Time: at.Add(time.Duration(minute) * time.Minute)}
	}
	records := []boardRecord{mk("a0", "A", "1", 10, 5, 0), mk("b0", "B", "1", 10, 5, 0), mk("a1", "A", "1", 10, 5, 1), mk("b1", "B", "1", 11, 5, 1), mk("a2", "A", "1", 10, 5, 2), mk("a3", "A", "1", 10, 6, 3), mk("a4", "A", "2", 10, 6, 4), mk("a5", "A", "2", 10, 6, 5), mk("a6", "A", "2", 1, 1, 6), mk("a7", "A", "2", 1, 1, 7)}
	out, n := boardCompactRecords(records, "A")
	if n != 1 || len(out) != 9 {
		t.Fatalf("removed %d: %+v", n, out)
	}
	for _, r := range out {
		if r.ID == "a1" {
			t.Fatal("duplicate retained")
		}
	}
	if len(records) != 10 || records[2].ID != "a1" {
		t.Fatal("input mutated")
	}
	if _, n = boardCompactRecords(records, "B"); n != 0 {
		t.Fatal("two endpoints removed")
	}
	records[2].Verified = false
	if _, n = boardCompactRecords(records, "A"); n != 0 {
		t.Fatal("verification boundary lost", n)
	}
}
func TestBoardCleanupBackupAndFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage-board.json")
	before := boardStore{Schema: 1, Key: "local-private-key", ExternalAuto: true, Records: []boardRecord{{ID: "first"}, {ID: "duplicate"}, {ID: "last"}}}
	after := before
	after.Records = []boardRecord{before.Records[0], before.Records[2]}
	after.ExternalAuto = false
	if e := boardPersist(path, before); e != nil {
		t.Fatal(e)
	}
	original, _ := os.ReadFile(path)
	// An unusable backup directory must prevent modifying the store.
	if e := os.WriteFile(filepath.Join(dir, "history-backups"), []byte("block"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := boardCommitCleanup(path, before, after); e == nil {
		t.Fatal("backup failure ignored")
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, original) {
		t.Fatal("store changed despite backup failure")
	}
	os.Remove(filepath.Join(dir, "history-backups"))
	backup, e := boardCommitCleanup(path, before, after)
	if e != nil {
		t.Fatal(e)
	}
	saved, _ := os.ReadFile(backup)
	if !bytes.Equal(saved, original) {
		t.Fatal("backup differs")
	}
	var loaded boardStore
	got, _ = os.ReadFile(path)
	if e = json.Unmarshal(got, &loaded); e != nil || len(loaded.Records) != 2 || loaded.ExternalAuto {
		t.Fatal("cleanup not persisted", e)
	}
	if _, e = os.Stat(backup); e != nil {
		t.Fatal("backup missing")
	}
}
