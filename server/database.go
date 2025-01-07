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

func (d *database) Close() error {
	return nil
}

func (d *database) CreateAlbum(ctx context.Context, album *Album) error {
	return d.db.WithContext(ctx).Create(album).Error
}

func (d *database) UpdateAlbum(ctx context.Context, album *Album) error {
	return d.db.WithContext(ctx).Save(album).Error
}

func (d *database) GetAlbums(ctx context.Context) ([]*Album, error) {
	var albums []*Album
	return albums, d.db.WithContext(ctx).Find(&albums).Error
}

func (d *database) GetAlbum(ctx context.Context, id uint64) (*Album, error) {
	var album *Album
	return album, d.db.WithContext(ctx).First(&album, id).Error
}

func (d *database) GetAlbumByTag(ctx context.Context, tag string) (*Album, error) {
	var album *Album
	return album, d.db.WithContext(ctx).Where("tag = ?", tag).Find(&album).Error
}

func (d *database) DeleteAlbum(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Delete(&Album{}, id).Error
}

func (d *database) CreateLog(ctx context.Context, album *Album) error {
	return d.db.WithContext(ctx).Create(&Log{
		Time:    time.Now(),
		AlbumID: album.ID,
	}).Error
}

func (d *database) DeleteLog(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Delete(&Log{}, id).Error
}

func (d *database) GetLogs(ctx context.Context) ([]*Log, error) {
	var logs []*Log
	return logs, d.db.WithContext(ctx).Preload("Album").Find(&logs).Error
}

func (d *database) GetLog(ctx context.Context, id uint64) (*Log, error) {
	var log *Log
	return log, d.db.WithContext(ctx).Preload("Album").First(&log, id).Error
}
