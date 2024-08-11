package database

import (
	"cache/model"
	"context"
	"database/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type BunDB struct {
	client *bun.DB
}

func Connect() (*BunDB, error) {
	sqlDB, err := sql.Open(sqliteshim.ShimName, "netflix.db")
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqlDB, sqlitedialect.New())
	return &BunDB{
		client: db,
	}, nil
}

func (b *BunDB) Close() error {
	return b.client.Close()
}

func (b *BunDB) GetKeys(ctx context.Context, kids []string) ([]model.Key, error) {
	keys := make([]model.Key, 0)
	err := b.client.NewSelect().Model(&keys).Where("kid IN (?)", bun.In(kids)).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (b *BunDB) GetKey(ctx context.Context, kid string) (*model.Key, error) {
	key := new(model.Key)
	err := b.client.NewSelect().Model(key).Where("kid = ?", kid).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (b *BunDB) SaveKey(ctx context.Context, key *model.Key) error {
	_, err := b.client.NewInsert().Model(key).Exec(ctx)
	return err
}
