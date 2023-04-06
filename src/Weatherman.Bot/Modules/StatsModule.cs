using Discord.Interactions;
using Discord.WebSocket;
using Microsoft.Extensions.Options;
using System.Diagnostics;
using System.Text;

namespace Weatherman.Bot.Modules
{
    [EnabledInDm(true)]
    public class StatsModule : InteractionModuleBase<SocketInteractionContext>
    {
        private readonly DiscordSocketClient _client;
        private readonly BotConfiguration _options;

        public StatsModule(DiscordSocketClient client, IOptions<BotConfiguration> options)
        {
            _client = client;
            _options = options.Value;
        }

        [SlashCommand("stats", "Get bot stats")]
        public async Task GetInviteAsync()
        {
            var uptime = DateTime.UtcNow - Process.GetCurrentProcess().StartTime.ToUniversalTime();

            var sb = new StringBuilder();

            sb.AppendLine("```");
            sb.AppendLine($"Runtime: {Environment.Version}");
            sb.AppendLine($"Uptime: {uptime:hh\\:mm\\:ss}");
            sb.AppendLine($"Memory used: {GetHumanReadableMemory(GC.GetTotalMemory(false))}");
            sb.AppendLine();
            sb.AppendLine($"Connected servers: {_client.Guilds.Count()}");
            sb.AppendLine($"Connected users: {_client.Guilds.Sum(a => a.MemberCount)}");

            if (Context.User.Id == _options.OwnerId)
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

            sb.AppendLine("```");

            await RespondAsync(sb.ToString(), ephemeral: true);
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

            return String.Format("{0:0.##} {1}", value, sizes[order]);
        }
    }
}
