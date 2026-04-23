package server

import (
	"testing"
)

func TestStoreSaveAndLoad(t *testing.T) {
	s := store{path: t.TempDir()}

	type payload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	err := s.save("test1", payload{Name: "hello", Value: 42})
	if err != nil {
		t.Fatalf("save failed: %s", err)
	}

	var loaded payload
	err = s.load("test1", &loaded)
	if err != nil {
		t.Fatalf("load failed: %s", err)
	}

	if loaded.Name != "hello" || loaded.Value != 42 {
		t.Fatalf("loaded data mismatch: %+v", loaded)
	}
}

func TestStoreLoadNonExistent(t *testing.T) {
	s := store{path: t.TempDir()}

	var v map[string]string
	err := s.load("does_not_exist", &v)
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
}

func TestStoreLoadNilTarget(t *testing.T) {
	s := store{path: t.TempDir()}

	err := s.load("key", nil)
	if err == nil {
		t.Fatal("expected error for nil target")
	}
}

func TestStoreLoadNonPointer(t *testing.T) {
	s := store{path: t.TempDir()}

	err := s.load("key", "not a pointer")
	if err == nil {
		t.Fatal("expected error for non-pointer target")
	}
}

func TestStoreRemove(t *testing.T) {
	s := store{path: t.TempDir()}

	err := s.save("removeme", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("save failed: %s", err)
	}

	removed, err := s.remove("removeme")
	if err != nil {
		t.Fatalf("remove failed: %s", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	// second remove should return false
	removed, err = s.remove("removeme")
	if err != nil {
		t.Fatalf("remove failed: %s", err)
	}
	if removed {
		t.Fatal("expected removed=false for already-removed key")
	}
}

func TestStoreRemoveNonExistent(t *testing.T) {
	s := store{path: t.TempDir()}

	removed, err := s.remove("nope")
	if err != nil {
		t.Fatalf("remove failed: %s", err)
	}
	if removed {
		t.Fatal("expected removed=false for non-existent key")
	}
}
