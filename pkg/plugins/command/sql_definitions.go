package commandplugin

import "time"

const initSQL = `
CREATE TABLE IF NOT EXISTS guild_profile (
	id TEXT NOT NULL PRIMARY KEY,
	prefix TEXT,
	lastChangedBy TEXT,
	lastChangedDate TIMESTAMP
);
`

type guildProfile struct {
	ID              string
	Prefix          *string
	LastChangedBy   *string
	LastChangedDate *time.Time
}
