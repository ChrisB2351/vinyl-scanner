package main

import (
	"context"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (d *database) CountAlbums(ctx context.Context) (int64, error) {
	var count int64
	return count, d.db.WithContext(ctx).Model(&Album{}).Count(&count).Error
}

func (d *database) GetAlbums(ctx context.Context, sort, order string, offset, limit int) ([]*Album, error) {
	var albums []*Album
	return albums, d.db.WithContext(ctx).
		Order(clause.OrderByColumn{Column: clause.Column{Name: sort}, Desc: order == "desc"}).
		Offset(offset).Limit(limit).
		Find(&albums).Error
}

func (d *database) GetAlbum(ctx context.Context, id uint64) (*Album, error) {
	var album *Album
	return album, d.db.WithContext(ctx).First(&album, id).Error
}

func (d *database) GetAlbumByTag(ctx context.Context, tag string) (*Album, error) {
	var album *Album
	return album, d.db.WithContext(ctx).Where("tag = ?", tag).First(&album).Error
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

func (d *database) CountLogs(ctx context.Context) (int64, error) {
	var count int64
	return count, d.db.WithContext(ctx).Model(&Log{}).Count(&count).Error
}

func (d *database) GetLogs(ctx context.Context, order string, offset, limit int) ([]*Log, error) {
	var logs []*Log
	return logs, d.db.WithContext(ctx).Preload("Album").
		Order(clause.OrderByColumn{Column: clause.Column{Name: "time"}, Desc: order == "desc"}).
		Offset(offset).Limit(limit).
		Find(&logs).Error
}

func (d *database) GetLog(ctx context.Context, id uint64) (*Log, error) {
	var log *Log
	return log, d.db.WithContext(ctx).Preload("Album").First(&log, id).Error
}
