using Discord.Commands;
using Discord.Interactions;
using Discord.WebSocket;
using Microsoft.Extensions.Options;
using System.Diagnostics;
using System.Text;

namespace Weatherman.Bot.Modules
{
    [EnabledInDm(true)]
    public class StatsModule : ModuleBase<SocketCommandContext>
    {
        private readonly DiscordSocketClient _client;
        private readonly BotConfiguration _options;

        public StatsModule(DiscordSocketClient client, IOptions<BotConfiguration> options)
        {
            _client = client;
            _options = options.Value;
        }

        [Command("stats")]
        public async Task GetStatsAsync()
        {
            var sb = new StringBuilder();

            sb.AppendLine("```");

            WriteHostMetrics(sb);

            if (Context.User.Id == _options.OwnerId)
            {
                WriteGuildData(sb);
            }

            sb.AppendLine("```");

            await Context.Channel.SendMessageAsync(sb.ToString());
        }

        private void WriteHostMetrics(StringBuilder sb)
        {
            var uptime = DateTime.UtcNow - Process.GetCurrentProcess().StartTime.ToUniversalTime();

            var memoryUsed = GetHumanReadableMemory(GC.GetTotalMemory(false));
            var memoryAllocated = GetHumanReadableMemory(GC.GetTotalAllocatedBytes(false));

            sb.AppendLine($"Runtime: {Environment.Version}");
            sb.AppendLine($"Uptime: {uptime.TotalHours:N0}:{uptime:mm\\:ss}");
            sb.AppendLine($"Memory used: {memoryUsed} / {memoryAllocated}");
            sb.AppendLine();
            sb.AppendLine($"Connected servers: {_client.Guilds.Count()}");
            sb.AppendLine($"Connected users: {_client.Guilds.Sum(a => a.MemberCount)}");
        }

        private void WriteGuildData(StringBuilder sb)
        {
            sb.AppendLine();
            sb.AppendLine("Connected Guilds:");

            var topGuilds = _client.Guilds.OrderByDescending(a => a.MemberCount).ToList().Take(3);
            foreach (var guild in topGuilds)
            {
                sb.AppendLine($"{guild.Name}: {guild.MemberCount}");
            }

            sb.AppendLine("----------");

            var newestGuilds = _client.Guilds.OrderByDescending(a => a.CurrentUser.JoinedAt).ToList().Take(10);
            foreach (var guild in newestGuilds)
            {
                sb.AppendLine($"{guild.Name}: {guild.MemberCount}");
            }
        }

        private string GetHumanReadableMemory(long value)
        {
            string[] sizes = { "b", "kb", "mb", "gb", "tb" };
            int order = 0;
            while (value >= 1024 && order < sizes.Length - 1)
            {
                order++;
                value = value / 1024;
            }

            return string.Format("{0:0.##} {1}", value, sizes[order]);
        }
    }
}
