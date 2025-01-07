package main

import (
	"context"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type database struct {
	db *gorm.DB
}

func newDatabase(path string) (*database, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Album{}, &Log{})
	if err != nil {
		return nil, err
	}

	return &database{
		db: db,
	}, nil
}

func (b *database) Close() error {
	return nil
}

func (b *database) CreateAlbum(ctx context.Context, album *Album) error {
	return b.db.Create(album).Error
}

func (b *database) UpdateAlbum(ctx context.Context, album *Album) error {
	return b.db.Save(album).Error
}

func (b *database) GetAlbums(ctx context.Context) ([]*Album, error) {
	var albums []*Album
	return albums, b.db.Find(&albums).Error
}

func (b *database) GetAlbum(ctx context.Context, id uint64) (*Album, error) {
	var album *Album
	return album, b.db.First(&album, id).Error
}

func (b *database) GetAlbumByTag(ctx context.Context, tag string) (*Album, error) {
	var album *Album
	return album, b.db.Where("tag = ?", tag).Find(&album).Error
}

func (b *database) DeleteAlbum(ctx context.Context, id uint64) error {
	return b.db.Delete(&Album{}, id).Error
}

func (d *database) CreateLog(ctx context.Context, album *Album) error {
	return d.db.Create(&Log{
		Time:    time.Now(),
		AlbumID: album.ID,
	}).Error
}

func (d *database) DeleteLog(ctx context.Context, id uint64) error {
	return d.db.Delete(&Album{}, id).Error
}

func (d *database) GetLogs(ctx context.Context) ([]*Log, error) {
	var logs []*Log
	return logs, d.db.Preload("Album").Find(&logs).Error
}

func (d *database) GetLog(ctx context.Context, id uint64) (*Log, error) {
	var log *Log
	return log, d.db.Preload("Album").First(&log, id).Error
}
