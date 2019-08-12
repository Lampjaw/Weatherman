package statsplugin

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/lampjaw/discordgobot"
)

type statsPlugin struct {
	discordgobot.Plugin
	version string
}

func New(appVersion string) *statsPlugin {
	return &statsPlugin{
		version: appVersion,
	}
}

func (p *statsPlugin) Commands() []discordgobot.CommandDefinition {
	return []discordgobot.CommandDefinition{
		discordgobot.CommandDefinition{
			CommandID: "stats",
			Triggers: []string{
				"stats",
			},
			Description: "Get bot stats",
			Callback:    p.runStatsCommand,
		},
	}
}

func (p *statsPlugin) Name() string {
	return "Stats"
}

var statsStartTime = time.Now()

func getDurationString(duration time.Duration) string {
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}

func (p *statsPlugin) runStatsCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message, args map[string]string, trigger string) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	w := &tabwriter.Writer{}
	buf := &bytes.Buffer{}

	w.Init(buf, 0, 4, 0, ' ', 0)
	fmt.Fprintf(w, "```\n")
	fmt.Fprintf(w, "Weatherman: \t%s\n", p.version)
	fmt.Fprintf(w, "discordgobot: \t%s\n", discordgobot.VERSION)
	fmt.Fprintf(w, "Go: \t%s\n", runtime.Version())
	fmt.Fprintf(w, "Uptime: \t%s\n", getDurationString(time.Now().Sub(statsStartTime)))
	fmt.Fprintf(w, "Memory used: \t%s / %s (%s garbage collected)\n", humanize.Bytes(stats.Alloc), humanize.Bytes(stats.Sys), humanize.Bytes(stats.TotalAlloc))
	fmt.Fprintf(w, "Concurrent tasks: \t%d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "Connected servers: \t%d\n", client.ChannelCount())
	fmt.Fprintf(w, "Connected users: \t%d\n", client.UserCount())
	if len(client.Sessions) > 1 {
		shards := 0
		for _, s := range client.Sessions {
			if s.DataReady {
				shards++
			}
		}
		if shards == len(client.Sessions) {
			fmt.Fprintf(w, "Shards: \t%d\n", shards)
		} else {
			fmt.Fprintf(w, "Shards: \t%d (%d connected)\n", len(client.Sessions), shards)
		}
		guild, err := client.Channel(message.Channel())
		if err == nil {
			id, err := strconv.Atoi(guild.ID)
			if err == nil {
				fmt.Fprintf(w, "Current shard: \t%d\n", ((id>>22)%len(client.Sessions) + 1))
			}
		}
	}

	if client.IsBotOwner(message) {
		guilds := client.Guilds()

		sort.SliceStable(guilds, func(i, j int) bool {
			return guilds[i].MemberCount > guilds[j].MemberCount
		})

		fmt.Fprintf(w, "\nConnected Guilds:\n")

		for _, guild := range guilds {
			fmt.Fprintf(w, "%s: \t%d\n", guild.Name, guild.MemberCount)
		}
	}

	fmt.Fprintf(w, "\n```")

	w.Flush()
	out := buf.String()

	end := ""

	if end != "" {
		out += "\n" + end
	}

	p.RLock()
	client.SendMessage(message.Channel(), out)
	p.RUnlock()
}
