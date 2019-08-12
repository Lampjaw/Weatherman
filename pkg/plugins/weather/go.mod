module github.com/lampjaw/weatherman/pkg/plugins/weather

go 1.12

require (
	github.com/bwmarrin/discordgo v0.19.0
	github.com/lampjaw/discordgobot v0.2.3
	github.com/lampjaw/weatherman/pkg/darksky v0.0.0
	github.com/lampjaw/weatherman/pkg/herelocation v0.0.0
	github.com/mattn/go-sqlite3 v1.11.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/sys v0.0.0-20190812073006-9eafafc0a87e // indirect
)

replace (
	github.com/lampjaw/weatherman/pkg/darksky => ../../darksky
	github.com/lampjaw/weatherman/pkg/herelocation => ../../herelocation
)
