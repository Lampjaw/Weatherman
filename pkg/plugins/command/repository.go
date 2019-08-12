package commandplugin

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	databaseDirectory = "/data/commandplugin/"
	databaseName      = "commandplugin.db"
)

type repository struct {
	Database *sql.DB
}

func newRepository() *repository {
	return &repository{}
}

func (r *repository) initRepository() {
	databaseDirectoryPath, _ := filepath.Abs(databaseDirectory)
	databaseFilePath := filepath.Join(databaseDirectoryPath, databaseName)

	os.MkdirAll(databaseDirectoryPath, 0755)

	db, err := sql.Open("sqlite3", databaseFilePath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(initSQL)
	if err != nil {
		log.Fatal("%q: %s\n", err, initSQL)
	}

	r.Database = db
}

func (r *repository) getGuildProfile(guildID string) (*guildProfile, error) {
	stmt, err := r.Database.Prepare("select id, prefix, lastChangedBy, lastChangedDate from guild_profile where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var record = &guildProfile{}
	err = stmt.QueryRow(guildID).Scan(
		&record.ID,
		&record.Prefix,
		&record.LastChangedBy,
		&record.LastChangedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func (r *repository) updateGuildPrefix(guildID string, userID string, prefix string) error {
	stmt, err := r.Database.Prepare("insert into guild_profile (id, prefix, lastChangedBy, lastChangedDate) values (?,?,?,?) on conflict (id) do update set prefix = excluded.prefix, lastChangedBy = excluded.lastChangedBy, lastChangedDate = excluded.lastChangedDate")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	_, err = stmt.Exec(guildID, prefix, userID, now)
	if err != nil {
		return err
	}

	return nil
}
