module github.com/lampjaw/weatherman/pkg/plugins/weather

go 1.12

require (
	github.com/bwmarrin/discordgo v0.19.0
	github.com/lampjaw/discordgobot v0.2.2
	github.com/lampjaw/weatherman/pkg/darksky v0.0.0
	github.com/lampjaw/weatherman/pkg/herelocation v0.0.0
	github.com/mattn/go-sqlite3 v1.11.0
)

replace (
	github.com/lampjaw/weatherman/pkg/darksky => ../../darksky
	github.com/lampjaw/weatherman/pkg/herelocation => ../../herelocation
)
