package weatherplugin

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	databaseDirectory = "/data/weatherplugin/"
	databaseName      = "weatherplugin.db"
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

func (r *repository) getUserProfile(userID string) (*userProfile, error) {
	stmt, err := r.Database.Prepare("select id, homeLocation, lastLocation, homeLocationChangedDate, lastLocationChangedDate from user_profile where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var record = &userProfile{}
	err = stmt.QueryRow(userID).Scan(
		&record.ID,
		&record.HomeLocation,
		&record.LastLocation,
		&record.HomeLocationChangedDate,
		&record.LastLocationChangedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func (r *repository) updateUserHomeLocation(userID string, homeLocation string) error {
	stmt, err := r.Database.Prepare("insert into user_profile (id, homeLocation, homeLocationChangedDate) values (?,?,?) on conflict (id) do update set homeLocation = excluded.homeLocation, homeLocationChangedDate = excluded.homeLocationChangedDate")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	_, err = stmt.Exec(userID, homeLocation, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) updateUserLastLocation(userID string, lastLocation string) error {
	stmt, err := r.Database.Prepare("insert into user_profile (id, lastLocation, lastLocationChangedDate) values (?,?,?) on conflict (id) do update set lastLocation = excluded.lastLocation, lastLocationChangedDate = excluded.lastLocationChangedDate")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	_, err = stmt.Exec(userID, lastLocation, now)
	if err != nil {
		return err
	}

	return nil
}
