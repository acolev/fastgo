package redis

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"sort"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type fakeCmdable struct {
	data    map[string]string
	ttl     map[string]time.Duration
	scanErr error
}

func newFakeCmdable() *fakeCmdable {
	return &fakeCmdable{
		data: make(map[string]string),
		ttl:  make(map[string]time.Duration),
	}
}

func (f *fakeCmdable) Get(_ context.Context, key string) *goredis.StringCmd {
	value, ok := f.data[key]
	if !ok {
		return goredis.NewStringResult("", goredis.Nil)
	}

	return goredis.NewStringResult(value, nil)
}

func (f *fakeCmdable) Set(_ context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	switch typed := value.(type) {
	case string:
		f.data[key] = typed
	case []byte:
		f.data[key] = string(typed)
	default:
		payload, err := json.Marshal(typed)
		if err != nil {
			return goredis.NewStatusResult("", err)
		}

		f.data[key] = string(payload)
	}

	f.ttl[key] = expiration

	return goredis.NewStatusResult("OK", nil)
}

func (f *fakeCmdable) Del(_ context.Context, keys ...string) *goredis.IntCmd {
	var deleted int64
	for _, key := range keys {
		if _, ok := f.data[key]; ok {
			delete(f.data, key)
			delete(f.ttl, key)
			deleted++
		}
	}

	return goredis.NewIntResult(deleted, nil)
}

func (f *fakeCmdable) Scan(_ context.Context, _ uint64, match string, _ int64) *goredis.ScanCmd {
	if f.scanErr != nil {
		return goredis.NewScanCmdResult(nil, 0, f.scanErr)
	}

	keys := make([]string, 0)
	for key := range f.data {
		matched, err := path.Match(match, key)
		if err != nil {
			return goredis.NewScanCmdResult(nil, 0, err)
		}

		if matched {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	return goredis.NewScanCmdResult(keys, 0, nil)
}

func TestJSONCacheRememberCachesLoadedValue(t *testing.T) {
	ctx := context.Background()
	cache := NewJSONCache[testPayload]("tests:numbers", time.Minute)
	fake := newFakeCmdable()
	cache.client = fake

	loadCount := 0
	loader := func(context.Context) (testPayload, error) {
		loadCount++
		return testPayload{Value: "cached"}, nil
	}

	first, err := cache.Remember(ctx, "list", loader)
	if err != nil {
		t.Fatalf("first Remember returned error: %v", err)
	}

	second, err := cache.Remember(ctx, "list", loader)
	if err != nil {
		t.Fatalf("second Remember returned error: %v", err)
	}

	if loadCount != 1 {
		t.Fatalf("loadCount = %d, want 1", loadCount)
	}

	if first != second {
		t.Fatalf("first = %+v, second = %+v", first, second)
	}

	if got := fake.ttl["tests:numbers:list"]; got != time.Minute {
		t.Fatalf("ttl = %s, want %s", got, time.Minute)
	}
}

func TestJSONCacheInvalidateAllByPrefix(t *testing.T) {
	ctx := context.Background()
	cache := NewJSONCache[testPayload]("tests:numbers", time.Minute)
	fake := newFakeCmdable()
	fake.data["tests:numbers:list"] = `{"value":"one"}`
	fake.data["tests:numbers:random"] = `{"value":"two"}`
	fake.data["other:key"] = `{"value":"keep"}`
	cache.client = fake

	if err := cache.InvalidateAll(ctx); err != nil {
		t.Fatalf("InvalidateAll returned error: %v", err)
	}

	if _, ok := fake.data["tests:numbers:list"]; ok {
		t.Fatal("tests:numbers:list should be deleted")
	}

	if _, ok := fake.data["tests:numbers:random"]; ok {
		t.Fatal("tests:numbers:random should be deleted")
	}

	if _, ok := fake.data["other:key"]; !ok {
		t.Fatal("other:key should remain")
	}
}

func TestJSONCacheGetInvalidJSONFails(t *testing.T) {
	ctx := context.Background()
	cache := NewJSONCache[testPayload]("tests:numbers", time.Minute)
	fake := newFakeCmdable()
	fake.data["tests:numbers:list"] = `{"value":`
	cache.client = fake

	_, _, err := cache.Get(ctx, "list")
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestJSONCacheKeyNormalizesParts(t *testing.T) {
	cache := NewJSONCache[testPayload](" tests:numbers ", time.Minute)

	got := cache.Key("", " list ", ":all:")
	if got != "tests:numbers:list:all" {
		t.Fatalf("key = %q", got)
	}
}

func TestJSONCacheGetMissReturnsFalse(t *testing.T) {
	ctx := context.Background()
	cache := NewJSONCache[testPayload]("tests:numbers", time.Minute)
	cache.client = newFakeCmdable()

	_, found, err := cache.Get(ctx, "missing")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if found {
		t.Fatal("found = true, want false")
	}
}

func TestJSONCacheInvalidateAllScanError(t *testing.T) {
	ctx := context.Background()
	cache := NewJSONCache[testPayload]("tests:numbers", time.Minute)
	fake := newFakeCmdable()
	fake.scanErr = errors.New("scan failed")
	cache.client = fake

	err := cache.InvalidateAll(ctx)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

type testPayload struct {
	Value string `json:"value"`
}
