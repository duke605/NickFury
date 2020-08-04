package route

import (
	"context"
	"encoding/json"
	"fmt"

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
