package store

import (
	"os"
	"testing"
)

func TestSeedDefaultStages(t *testing.T) {
	path := "/tmp/seed_test_fresh.db"
	os.Remove(path)
	defer os.Remove(path)

	s, err := Open(path)
	if err != nil {
		t.Fatal("Open:", err)
	}
	stages, err := s.ListDefaultStages()
	if err != nil {
		t.Fatal("List:", err)
	}
	if len(stages) == 0 {
		t.Fatal("expected default stages, got 0")
	}
	t.Logf("got %d default stages", len(stages))
}

func TestSeedDefaultStagesWithFKEnabled(t *testing.T) {
	path := "/tmp/seed_test_fk.db"
	os.Remove(path)
	defer os.Remove(path)

	s, err := Open(path)
	if err != nil {
		t.Fatal("Open:", err)
	}

	// Enable FK enforcement, wipe default stages, reopen — seed must succeed despite FK constraint
	s.db.Exec("PRAGMA foreign_keys = ON")
	s.db.Where("job_id = 0").Delete(&Stage{})

	s2, err := Open(path)
	if err != nil {
		t.Fatal("Reopen with FK on:", err)
	}
	stages, err := s2.ListDefaultStages()
	if err != nil {
		t.Fatal("List after reopen:", err)
	}
	if len(stages) == 0 {
		t.Fatal("expected default stages after reopen with FK enabled, got 0")
	}
	t.Logf("got %d default stages with FK enforcement on", len(stages))
}
