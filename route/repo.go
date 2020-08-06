package route

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/duke605/NickFury/datastore"
)

// Route ...
type Route struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Section   int    `json:"section"`
	Path      string `json:"path"`
}

// Map ...
type Map struct {
	ID       string   `json:"id"`
	Sections byte     `json:"sections"`
	MaxPaths []string `json:"max_paths"`
}

// Paths returns the valid paths for the provided section
func (m Map) Paths(section int) []string {
	limit := ([]byte(m.MaxPaths[section-1])[0] - 'A') + 1
	return strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ"[:limit], "")
}

// IsValidPath returns true if the path is valid for the section
func (m Map) IsValidPath(section int, p string) bool {
	paths := m.Paths(section)
	p = strings.ToUpper(p)
	return p[0] >= paths[0][0] && p[0] <= paths[len(paths)-1][0]
}

// GetID returns the ID for the route
func (r Route) GetID() []byte {
	if r.ID == "" {
		return []byte(fmt.Sprintf("%s:%s:%d:%s", r.UserID, r.ChannelID, r.Section, r.Path))
	}

	return []byte(r.ID)
}

// Repository handles the communication between the application and
// persistant storage
type Repository struct {
	*datastore.Datastore
}

// NewRepo creates a new Repository
func NewRepo(db *bolt.DB) *Repository {
	return &Repository{
		Datastore: &datastore.Datastore{DB: db},
	}
}

// GetRoutesInChannel gets all the routes currently linked to the channel
func (repo *Repository) GetRoutesInChannel(ctx context.Context, channelID string) ([]Route, error) {

	// Getting a route for the user ID
	routes := []Route{}
	err := repo.InTransaction(ctx, false, func(_ context.Context, tx *bolt.Tx) error {
		b := tx.Bucket([]byte("routes"))
		if b == nil {
			return nil
		}

		// Looking for route
		return b.ForEach(func(k, v []byte) error {
			r := Route{}
			if err := json.Unmarshal(v, &r); err != nil {
				return err
			}

			if r.ChannelID == channelID {
				routes = append(routes, r)
			}

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return routes, nil
}

// DeleteAllRoutesForChannel deletes all assined routes for a channel
func (repo *Repository) DeleteAllRoutesForChannel(ctx context.Context, channelID string) error {
	return repo.InTransaction(ctx, true, func(ctx context.Context, tx *bolt.Tx) error {
		routes, err := repo.GetRoutesInChannel(ctx, channelID)
		if err != nil {
			return err
		}

		// Deleting the routes
		for _, r := range routes {
			err = repo.DeleteRoute(ctx, r)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// InsertRoute persists a route
func (repo *Repository) InsertRoute(ctx context.Context, r Route) error {
	return repo.InTransaction(ctx, true, func(_ context.Context, tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("routes"))
		if err != nil {
			return err
		}

		buf, err := json.Marshal(r)
		if err != nil {
			return err
		}

		return b.Put(r.GetID(), buf)
	})
}

// DeleteRoute deletes the route
func (repo *Repository) DeleteRoute(ctx context.Context, r Route) error {
	return repo.InTransaction(ctx, true, func(_ context.Context, tx *bolt.Tx) error {
		b := tx.Bucket([]byte("routes"))
		if b == nil {
			return nil
		}

		return b.Delete(r.GetID())
	})
}

// GetMapForChannel returns a map from the database that is for the channel. Returns sql.ErrNoRows if no
// map could be found for the channel
func (repo *Repository) GetMapForChannel(ctx context.Context, channelID string) (Map, error) {
	m := Map{}
	err := repo.InTransaction(ctx, false, func(_ context.Context, tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("maps"))
		if buk == nil {
			return sql.ErrNoRows
		}

		data := buk.Get([]byte(channelID))
		if data == nil {
			return sql.ErrNoRows
		}

		return json.Unmarshal(data, &m)
	})

	return m, err
}

// InsertMap persists a map
func (repo *Repository) InsertMap(ctx context.Context, m Map) error {
	return repo.InTransaction(ctx, true, func(_ context.Context, tx *bolt.Tx) error {
		buk, err := tx.CreateBucketIfNotExists([]byte("maps"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		return buk.Put([]byte(m.ID), data)
	})
}
