using Microsoft.Extensions.DependencyInjection;
using System.Reflection;

namespace Weatherman.Bot.Cache
{
    public static class CacheExtensions
    {
        public static IServiceCollection AddCache(this IServiceCollection services, Action<CacheOptions> actionOptions)
        {
            services.AddOptions();
            services.Configure<CacheOptions>(actionOptions);

            services.PostConfigure<CacheOptions>(options =>
            {
                if (string.IsNullOrWhiteSpace(options.KeyPrefix))
                {
                    options.KeyPrefix = typeof(Assembly).Name;
                }
            });

            services.AddSingleton<CacheConnector>();
            services.AddSingleton<ICache, Cache>();

            return services;
        }
    }
}
