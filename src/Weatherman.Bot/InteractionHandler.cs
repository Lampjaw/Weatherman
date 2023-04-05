using Discord;
using Discord.Addons.Hosting;
using Discord.Addons.Hosting.Util;
using Discord.Interactions;
using Discord.WebSocket;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using System.Reflection;

namespace Weatherman.Bot
{
    internal class InteractionHandler : DiscordClientService
    {
        private readonly IServiceProvider _serviceProvider;
        private readonly InteractionService _interactionService;
        private readonly BotConfiguration _options;

        public InteractionHandler(DiscordSocketClient client, ILogger<DiscordClientService> logger, IServiceProvider provider, InteractionService interactionService, IOptions<BotConfiguration> options) : base(client, logger)
        {
            _serviceProvider = provider;
            _interactionService = interactionService;
            _options = options.Value;
        }

        protected override async Task ExecuteAsync(CancellationToken cancellationToken)
        {
            Client.InteractionCreated += HandleInteractionAsync;

            await _interactionService.AddModulesAsync(Assembly.GetEntryAssembly(), _serviceProvider);
            await Client.WaitForReadyAsync(cancellationToken);

            if (_options.GuildId != null)
            {
                await _interactionService.RegisterCommandsToGuildAsync(_options.GuildId.Value, true);
            }
            else
            {
                await _interactionService.RegisterCommandsGloballyAsync(true);
            }
        }

        private async Task HandleInteractionAsync(SocketInteraction interaction)
        {
            try
            {
                var context = new SocketInteractionContext(Client, interaction);

                var result = await _interactionService.ExecuteCommandAsync(context, _serviceProvider);

                if (!result.IsSuccess)
                {
                    await context.Interaction.RespondAsync(result.ErrorReason);
                }
            }
            catch (Exception ex)
            {
                Logger.LogError(ex, "Exception occurred whilst attempting to handle interaction.");

                if (interaction.Type is InteractionType.ApplicationCommand)
                {
                    await interaction.GetOriginalResponseAsync().ContinueWith(async (msg) => await msg.Result.DeleteAsync());
                }
            }
        }
    }
}
