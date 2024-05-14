package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	albumsBucket = []byte("albums")
	logsBucket   = []byte("logs")
	tagsBucket   = []byte("tags")

	errNoItem = errors.New("item does not exist")
)

type database struct {
	db *bolt.DB
}

func newDatabase(path string) (*database, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	return &database{
		db: db,
	}, nil
}

func (b *database) Close() error {
	return b.db.Close()
}

func (b *database) CreateAlbum(ctx context.Context, album *Album) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(albumsBucket)
		if err != nil {
			return err
		}

		if album.ID == 0 {
			id, err := b.NextSequence()
			if err != nil {
				return err
			}

			album.ID = id
		}

		var buf bytes.Buffer
		err = gob.NewEncoder(&buf).Encode(album)
		if err != nil {
			return err
		}

		return b.Put(encodeUint64(album.ID), buf.Bytes())
	})
}

func (b *database) UpdateAlbum(ctx context.Context, album *Album) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(albumsBucket)
		if b == nil {
			return errNoItem
		}

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(album)
		if err != nil {
			return err
		}

		return b.Put(encodeUint64(album.ID), buf.Bytes())
	})
}

func (b *database) GetAlbums(ctx context.Context) ([]*Album, error) {
	var albums []*Album

	return albums, b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(albumsBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var album *Album
			err := gob.NewDecoder(bytes.NewReader(v)).Decode(&album)
			if err != nil {
				return err
			}

			albums = append(albums, album)
			return nil
		})
	})
}

func (b *database) GetAlbum(ctx context.Context, id uint64) (*Album, error) {
	var album *Album

	return album, b.db.View(func(tx *bolt.Tx) error {
		return getAlbum(tx, encodeUint64(id), &album)
	})
}

func (b *database) DeleteAlbum(ctx context.Context, id uint64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(albumsBucket)
		if b == nil {
			return nil
		}

		return b.Delete(encodeUint64(id))
	})
}

func (d *database) CreateLog(ctx context.Context, album *Album) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(logsBucket)
		if err != nil {
			return err
		}

		ts := time.Now().UTC().Format(time.RFC3339)
		return b.Put([]byte(ts), encodeUint64(album.ID))
	})
}

func (d *database) DeleteLog(ctx context.Context, ts string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(logsBucket)
		if b == nil {
			return nil
		}

		return b.Delete([]byte(ts))
	})
}

func (d *database) GetLogs(ctx context.Context) ([]*Log, error) {
	var logs []*Log

	return logs, d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(logsBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var album *Album
			err := getAlbum(tx, v, &album)
			if err != nil {
				return err
			}

			logs = append(logs, &Log{
				Timestamp: string(k),
				Album:     album,
			})

			return nil
		})
	})
}

func (d *database) GetLog(ctx context.Context, ts string) (*Log, error) {
	log := &Log{
		Timestamp: ts,
	}

	return log, d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(logsBucket)
		if b == nil {
			return errNoItem
		}

		v := b.Get([]byte(ts))
		if v == nil {
			return errNoItem
		}

		return getAlbum(tx, v, &log.Album)
	})
}

func (d *database) CreateTag(ctx context.Context, tag string, album *Album) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(tagsBucket)
		if err != nil {
			return err
		}

		return b.Put([]byte(tag), encodeUint64(album.ID))
	})
}

func (d *database) DeleteTag(ctx context.Context, tag string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(tagsBucket)
		if b == nil {
			return nil
		}

		return b.Delete([]byte(tag))
	})
}

func (d *database) GetTags(ctx context.Context) ([]*Tag, error) {
	var tags []*Tag

	return tags, d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(tagsBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var album *Album
			err := getAlbum(tx, v, &album)
			if err != nil {
				return err
			}

			tags = append(tags, &Tag{
				ID:    string(k),
				Album: album,
			})

			return nil
		})
	})
}

func (d *database) GetTag(ctx context.Context, tagID string) (*Tag, error) {
	tag := &Tag{
		ID: tagID,
	}

	return tag, d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(tagsBucket)
		if b == nil {
			return errNoItem
		}

		v := b.Get([]byte(tagID))
		if v == nil {
			return errNoItem
		}

		return getAlbum(tx, v, &tag.Album)
	})
}

// encodeUint64 returns an 8-byte big endian representation of v.
func encodeUint64(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func getAlbum(tx *bolt.Tx, id []byte, e any) error {
	b := tx.Bucket(albumsBucket)
	if b == nil {
		return errNoItem
	}

	v := b.Get(id)
	if v == nil {
		return errNoItem
	}

	return gob.NewDecoder(bytes.NewReader(v)).Decode(e)
}
