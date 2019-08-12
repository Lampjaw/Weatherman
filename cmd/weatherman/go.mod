module github.com/lampjaw/weatherman/cmd/weatherman

go 1.12

require (
	github.com/lampjaw/discordgobot v0.2.2
	github.com/lampjaw/weatherman/pkg/plugins/command v0.0.0
	github.com/lampjaw/weatherman/pkg/plugins/invite v0.0.0
	github.com/lampjaw/weatherman/pkg/plugins/stats v0.0.0
	github.com/lampjaw/weatherman/pkg/plugins/weather v0.0.0
)

replace (
	github.com/lampjaw/weatherman/pkg/darksky => ../../pkg/darksky
	github.com/lampjaw/weatherman/pkg/herelocation => ../../pkg/herelocation
	github.com/lampjaw/weatherman/pkg/plugins/command => ../../pkg/plugins/command
	github.com/lampjaw/weatherman/pkg/plugins/invite => ../../pkg/plugins/invite
	github.com/lampjaw/weatherman/pkg/plugins/stats => ../../pkg/plugins/stats
	github.com/lampjaw/weatherman/pkg/plugins/weather => ../../pkg/plugins/weather
)
