package file

import (
	"context"
	"sync"
	"testing"
)

func newTestService(t *testing.T) *service {
	t.Helper()
	svc, err := newService(ts.Profile)
	if err != nil {
		t.Fatalf("failed to create service: %s", err)
	}
	return svc
}

func TestServiceProcessSyncSingleFile(t *testing.T) {
	svc := newTestService(t)

	stream := make(chan string, 1)
	stream <- fakeMalwareSample
	close(stream)

	var results []resultRecord
	var mu sync.Mutex

	err := svc.process(context.Background(), stream, 1, "", false, map[string]string{"m1": "v1"}, func(r resultRecord) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("process failed: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := &results[0]
	if r.err != nil {
		t.Fatalf("expected no error, got %s", r.err)
	}
	checkResponseContent(t, r)
	if r.metadata["m1"] != "v1" {
		t.Fatalf("expected metadata m1=v1, got %s", r.metadata["m1"])
	}
}

func TestServiceProcessAsyncSingleFile(t *testing.T) {
	svc := newTestService(t)

	stream := make(chan string, 1)
	stream <- fakeMalwareSample
	close(stream)

	var results []resultRecord
	var mu sync.Mutex

	err := svc.process(context.Background(), stream, 1, "", true, map[string]string{"m1": "v1"}, func(r resultRecord) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("process failed: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.err != nil {
		t.Fatalf("expected no error, got %s", r.err)
	}
	if r.id == "" {
		t.Fatalf("expected result to have an id")
	}
	if r.location == "" {
		t.Fatalf("expected result to have a location")
	}
}

func TestServiceProcessMultipleFiles(t *testing.T) {
	svc := newTestService(t)

	stream := make(chan string, 3)
	stream <- fakeMalwareSample
	stream <- fakeMalwareSample
	stream <- fakeMalwareSample
	close(stream)

	var results []resultRecord
	var mu sync.Mutex

	err := svc.process(context.Background(), stream, 2, "", false, nil, func(r resultRecord) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("process failed: %s", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		if r.err != nil {
			t.Fatalf("result %d: expected no error, got %s", i, r.err)
		}
		checkResponseContent(t, &r)
	}
}

func TestServiceProcessNonExistentFile(t *testing.T) {
	svc := newTestService(t)

	stream := make(chan string, 1)
	stream <- "testdata/does_not_exist.txt"
	close(stream)

	var results []resultRecord
	var mu sync.Mutex

	err := svc.process(context.Background(), stream, 1, "", false, nil, func(r resultRecord) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("process failed: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].err == nil {
		t.Fatalf("expected an error for non-existent file")
	}
}

func TestServiceRetrieve(t *testing.T) {
	svc := newTestService(t)

	// first process a file to get an id
	stream := make(chan string, 1)
	stream <- fakeMalwareSample
	close(stream)

	var results []resultRecord
	var mu sync.Mutex

	err := svc.process(context.Background(), stream, 1, "", false, map[string]string{"m1": "v1"}, func(r resultRecord) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("process failed: %s", err)
	}

	if len(results) != 1 || results[0].id == "" {
		t.Fatalf("expected a result with an id")
	}

	// now retrieve it
	retrieved, err := svc.retrieve(context.Background(), results[0].id)
	if err != nil {
		t.Fatalf("retrieve failed: %s", err)
	}

	checkResponseContent(t, retrieved)
	if retrieved.metadata["m1"] != "v1" {
		t.Fatalf("expected metadata m1=v1, got %s", retrieved.metadata["m1"])
	}
}

func TestServiceRetrieveAsync(t *testing.T) {
	svc := newTestService(t)

	// process a file async
	stream := make(chan string, 1)
	stream <- fakeMalwareSample
	close(stream)

	var results []resultRecord
	var mu sync.Mutex

	err := svc.process(context.Background(), stream, 1, "", true, map[string]string{"m1": "v1"}, func(r resultRecord) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("process failed: %s", err)
	}

	if len(results) != 1 || results[0].id == "" {
		t.Fatalf("expected a result with an id")
	}

	// retrieve the async result
	retrieved, err := svc.retrieve(context.Background(), results[0].id)
	if err != nil {
		t.Fatalf("retrieve failed: %s", err)
	}

	checkResponseContent(t, retrieved)
}
