using System.Text.Json;
using Weatherman.Bot.Cache;
using Weatherman.Bot.Data;
using Weatherman.Bot.Data.Models;
using Weatherman.Bot.Models;

namespace Weatherman.Bot.Services
{
    public class HomeService
    {
        private readonly DbContextHelper _dbContextHelper;
        private readonly ICache _cache;

        private const string _cacheKeyPrefix = "userhome";

        private readonly TimeSpan _userHomeCacheExpiration = TimeSpan.FromHours(1);

        public HomeService(DbContextHelper dbContextHelper, ICache cache)
        {
            _dbContextHelper = dbContextHelper;
            _cache = cache;
        }

        public async Task SetHomeAsync(ulong userId, LocationDetails location)
        {
            var homeLocation = JsonSerializer.Serialize(location);

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var userProfile = await dbContext.UserProfiles.FindAsync(userId.ToString());
                if (userProfile != null)
                {
                    userProfile.HomeLocation = homeLocation;
                    userProfile.HomeLocationChangedDate = DateTime.UtcNow;

                    dbContext.Update(userProfile);
                    await dbContext.SaveChangesAsync();
                }
                else
                {
                    userProfile = new UserProfile
                    {
                        Id = userId.ToString(),
                        HomeLocation = homeLocation,
                        HomeLocationChangedDate = DateTime.UtcNow
                    };

                    dbContext.Add(userProfile);
                    await dbContext.SaveChangesAsync();
                }

                var cacheKey = $"{_cacheKeyPrefix}-{userId}";
                await _cache.RemoveAsync(cacheKey);
            }
        }

        public async Task<LocationDetails> GetHomeAsync(ulong userId)
        {
            var cacheKey = $"{_cacheKeyPrefix}-{userId}";

            var location = await _cache.GetAsync<LocationDetails>(cacheKey);
            if (location != null)
            {
                return location;
            }

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var userProfile = await dbContext.UserProfiles.FindAsync(userId.ToString());
                if (userProfile == null)
                {
                    return null;
                }

                location = JsonSerializer.Deserialize<LocationDetails>(userProfile.HomeLocation);

                if (location != null)
                {
                    await _cache.SetAsync(cacheKey, location, _userHomeCacheExpiration);
                }

                return location;
            }
        }

        public async Task RemoveHomeAsync(ulong userId)
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var userProfile = await dbContext.UserProfiles.FindAsync(userId.ToString());
                if (userProfile == null)
                {
                    return;
                }

                var cacheKey = $"{_cacheKeyPrefix}-{userId}";
                await _cache.RemoveAsync(cacheKey);

                dbContext.Remove(userProfile);
                await dbContext.SaveChangesAsync();
            }
        }
    }
}
