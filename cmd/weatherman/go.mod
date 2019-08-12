module github.com/lampjaw/weatherman/cmd/weatherman

go 1.12

require (
	github.com/lampjaw/discordgobot v0.2.3
	github.com/lampjaw/weatherman/pkg/plugins/command v0.0.0
	github.com/lampjaw/weatherman/pkg/plugins/invite v0.0.0
	github.com/lampjaw/weatherman/pkg/plugins/stats v0.0.0
	github.com/lampjaw/weatherman/pkg/plugins/weather v0.0.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80 // indirect
	golang.org/x/sys v0.0.0-20190812073006-9eafafc0a87e // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190809145639-6d4652c779c4 // indirect
)

replace (
	github.com/lampjaw/weatherman/pkg/darksky => ../../pkg/darksky
	github.com/lampjaw/weatherman/pkg/herelocation => ../../pkg/herelocation
	github.com/lampjaw/weatherman/pkg/plugins/command => ../../pkg/plugins/command
	github.com/lampjaw/weatherman/pkg/plugins/invite => ../../pkg/plugins/invite
	github.com/lampjaw/weatherman/pkg/plugins/stats => ../../pkg/plugins/stats
	github.com/lampjaw/weatherman/pkg/plugins/weather => ../../pkg/plugins/weather
)
