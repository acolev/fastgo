package dev

import "fastgo/internal/infra/database/seeds"

func Names() []string {
	return []string{NumbersName}
}

func Seeders(seed uint64, count int) ([]seeds.Seeder, error) {
	numbers, err := NewNumbers(seed, count)
	if err != nil {
		return nil, err
	}

	return []seeds.Seeder{numbers}, nil
}
