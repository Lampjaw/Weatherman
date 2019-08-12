package weatherplugin

import "time"

const initSQL = `
CREATE TABLE IF NOT EXISTS user_profile (
	id TEXT NOT NULL PRIMARY KEY,
	homeLocation TEXT,
	lastLocation TEXT,
	homeLocationChangedDate TIMESTAMP,
	lastLocationChangedDate TIMESTAMP
);
`

type userProfile struct {
	ID                      string
	HomeLocation            *string
	LastLocation            *string
	HomeLocationChangedDate *time.Time
	LastLocationChangedDate *time.Time
}
