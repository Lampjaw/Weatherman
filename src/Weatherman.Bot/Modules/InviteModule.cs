using Discord.Interactions;
using Microsoft.Extensions.Options;

namespace Weatherman.Bot.Modules
{
    [EnabledInDm(true)]
    public class InviteModule : InteractionModuleBase<SocketInteractionContext>
    {
        private readonly string _inviteLink;

        public InviteModule(IOptions<BotConfiguration> options)
        {
            _inviteLink = string.Format("https://discordapp.com/oauth2/authorize?client_id={0}&scope=bot", options.Value.DiscordClientId);
        }

        [SlashCommand("invite", "Get an invite link to add this bot to your server!")]
        public async Task GetInviteAsync()
        {
            var inviteText = string.Format("Please visit <{0}> to add {1} to your server.", _inviteLink, Context.Client.CurrentUser.Username);

            await RespondAsync(inviteText, ephemeral: true);
        }
    }
}
