package weatherplugin

import "time"

const initSQL = `
CREATE TABLE IF NOT EXISTS user_profile (
	id TEXT NOT NULL PRIMARY KEY,
	homeLocation TEXT,
	homeLocationChangedDate TIMESTAMP
);
`

type userProfile struct {
	ID                      string
	HomeLocation            *string
	HomeLocationChangedDate *time.Time
}
