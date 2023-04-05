using Discord.Interactions;
using Weatherman.Bot.Services;

namespace Weatherman.Bot.Modules
{
    [Group("home", "Home commands")]
    [EnabledInDm(false)]
    public class HomeModule : InteractionModuleBase<SocketInteractionContext>
    {
        private readonly LocationService _locationService;
        private readonly HomeService _homeService;

        public HomeModule(LocationService locationService, HomeService homeService)
        {
            _locationService = locationService;
            _homeService = homeService;
        }

        [SlashCommand("set", "Set a default location to be used for your user.")]
        public async Task SetHomeAsync(string location)
        {
            if (string.IsNullOrWhiteSpace(location))
            {
                await RespondAsync("sethome requires a location to set!", ephemeral: true);
                return;
            }

            var foundLocation = await _locationService.GetGeocodeForLocationStringAsync(location);
            if (foundLocation == null)
            {
                await RespondAsync("failed to resolve a location", ephemeral: true);
                return;
            }

            await _homeService.SetHomeAsync(Context.User.Id, foundLocation);

            await RespondAsync("Home set!", ephemeral: true);
        }

        [SlashCommand("remove", "Remove any home settings for your user.")]
        public async Task RemoveHomeAsync()
        {
            await _homeService.RemoveHomeAsync(Context.User.Id);

            await RespondAsync("Home removed!", ephemeral: true);
        }
    }
}
