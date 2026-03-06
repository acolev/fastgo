package redis

import "testing"

func TestRedisOptionsParsesURL(t *testing.T) {
	options, err := redisOptions("redis://:secret@127.0.0.1:6379/2")
	if err != nil {
		t.Fatalf("redisOptions returned error: %v", err)
	}

	if options.Addr != "127.0.0.1:6379" {
		t.Fatalf("Addr = %q, want %q", options.Addr, "127.0.0.1:6379")
	}

	if options.Password != "secret" {
		t.Fatalf("Password = %q, want %q", options.Password, "secret")
	}

	if options.DB != 2 {
		t.Fatalf("DB = %d, want %d", options.DB, 2)
	}
}

func TestRedisOptionsFallbackToAddr(t *testing.T) {
	options, err := redisOptions("127.0.0.1:6379")
	if err != nil {
		t.Fatalf("redisOptions returned error: %v", err)
	}

	if options.Addr != "127.0.0.1:6379" {
		t.Fatalf("Addr = %q, want %q", options.Addr, "127.0.0.1:6379")
	}

	if options.DB != 0 {
		t.Fatalf("DB = %d, want %d", options.DB, 0)
	}
}

func TestRedisOptionsRejectsEmptyValue(t *testing.T) {
	if _, err := redisOptions("   "); err == nil {
		t.Fatal("expected error for empty REDIS_URL")
	}
}
