using Discord.Addons.Hosting;
using Discord.Commands;
using Discord.WebSocket;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using System.Reflection;

namespace Weatherman.Bot
{
    internal class CommandHandler : DiscordClientService
    {
        private readonly IServiceProvider _serviceProvider;
        private readonly CommandService _commandService;

        public CommandHandler(DiscordSocketClient client, ILogger<DiscordClientService> logger, IServiceProvider provider, CommandService commandService, IOptions<BotConfiguration> options) : base(client, logger)
        {
            _serviceProvider = provider;
            _commandService = commandService;
        }

        protected override async Task ExecuteAsync(CancellationToken cancellationToken)
        {
            Client.MessageReceived += HandleCommandAsync;

            await _commandService.AddModulesAsync(Assembly.GetEntryAssembly(), _serviceProvider);
        }

        private async Task HandleCommandAsync(SocketMessage socketMessage)
        {
            try
            {
                int argPos = 0;

                var message = socketMessage as SocketUserMessage;
                if (message == null || message.Author.IsBot || !message.HasMentionPrefix(Client.CurrentUser, ref argPos))
                {
                    return;
                }

                var context = new SocketCommandContext(Client, message);

                var result = await _commandService.ExecuteAsync(context, argPos, _serviceProvider);

                if (!result.IsSuccess && result.Error != CommandError.UnknownCommand)
                {
                    Logger.LogWarning($"Failed to handle command: {result.Error}: {result.ErrorReason}");

                    await context.Channel.SendMessageAsync("Something went wrong processing this request.");
                }
            }
            catch (Exception ex)
            {
                Logger.LogError(ex, "Exception occurred whilst attempting to handle command.");
            }
        }
    }
}
