package install

import "fastgo/internal/infra/database/seeds"

// Seeders returns mandatory application data. Keep every install seeder
// idempotent because this group may run more than once during deployments.
func Seeders() []seeds.Seeder {
	return nil
}

func Names() []string {
	return nil
}
