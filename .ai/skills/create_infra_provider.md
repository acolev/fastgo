
# Skill: create_infra_provider

Creates a new infrastructure provider.

Location:

internal/infra/<provider>/

Structure:

init.go
client.go

Example:

internal/infra/storage/

init.go
client.go

Initialization must be called from:

bootstrap/providers.go

Example:

func InitProviders(cfg *config.Config) {
    database.Init(cfg)
    redis.Init(cfg)
    storage.Init(cfg)
}

Provider should expose singleton access:

func Client() *ClientType
