package twitch

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

// Database is the str
type Database struct {
	db *bolt.DB

	// bucket for holding the list of twitch names
	// we need to check for updates on
	twitch *bolt.Bucket

	// bucket for holding the discord webhooks
	// for each twitch channel
	discord *bolt.Bucket
}

// webhook stores the data of a single webhook
type webhook struct {
	id    string
	token string
}

// New returns a new database
func New() *Database {
	boltDB, err := bolt.Open(
		"twitch.db",
		// read+write
		0600,
		&bolt.Options{
			// if file cannot be opened in x seconds, throw an error
			// this will fail if another instance of the service
			// has a read lock on the database file
			Timeout: 1 * time.Second,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	db := &Database{
		db: boltDB,
	}
	err = db.init()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// AddChannel adds a twitch channel to motitor and adds the channelID + webhook to be notified
func (d *Database) AddChannel(twitchName, channelID string, hook webhook) (err error) {
	// add the twitch channel name to the twitch bucket so we know to ask for updates for it
	err = d.twitch.Put([]byte(twitchName), []byte("0"))
	if err != nil {
		return
	}

	// get/make the bucket that holds all the channels receiving updates for a twitch channel
	b, err := d.discord.CreateBucketIfNotExists([]byte(twitchName))
	if err != nil {
		return
	}

	// marshal the webhook so it can be saved in the bucket
	hookMarshaled, err := json.Marshal(hook)
	if err != nil {
		return
	}

	// put the webhook data in the bucket
	err = b.Put([]byte(channelID), hookMarshaled)
	return
}

// Close closes the current databse
func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) init() error {
	return d.db.Update(func(tx *bolt.Tx) error {
		var err error
		d.twitch, err = tx.CreateBucketIfNotExists([]byte("twitch-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		d.discord, err = tx.CreateBucketIfNotExists([]byte("discord-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}
