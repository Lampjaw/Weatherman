using Discord;
using Discord.Addons.Hosting;
using Discord.WebSocket;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Weatherman.Bot;
using Weatherman.Bot.Services;
using Geo.Here.DependencyInjection;
using DarkSky.Services;
using Weatherman.Bot.Data;
using Weatherman.Bot.Cache;
using Serilog;
using Serilog.Events;

const LogSeverity DiscordLogLevel = LogSeverity.Info;

var logger = new LoggerConfiguration()
    .MinimumLevel.Information()
    .MinimumLevel.Override("Microsoft", LogEventLevel.Error)
    .MinimumLevel.Override("System.Net.Http.HttpClient.IHereGeocoding", LogEventLevel.Error);

if (string.Equals(Environment.GetEnvironmentVariable("LoggingOutput"), "flat", StringComparison.OrdinalIgnoreCase))
{
    logger.WriteTo.Console(outputTemplate: "[{Level:u4} {Timestamp:HH:mm:ss.fff}] {SourceContext}{NewLine}{Message} {Exception}{NewLine}");
}
else
{
    logger.WriteTo.Sink<GraylogConsoleSink>();
}

Log.Logger = logger.CreateLogger();

IHost host = Host.CreateDefaultBuilder(args)
    .UseSerilog()
    .ConfigureAppConfiguration((hostingContext, builder) =>
    {
        builder.Sources.Clear();

        builder.AddEnvironmentVariables();
    })
    .ConfigureDiscordHost((context, config) =>
    {
        config.SocketConfig = new DiscordSocketConfig
        {
            LogLevel = DiscordLogLevel,
            MessageCacheSize = 0
        };

        config.Token = context.Configuration.Get<BotConfiguration>().DiscordToken;

        config.SocketConfig.GatewayIntents = GatewayIntents.Guilds | GatewayIntents.GuildMessages | GatewayIntents.DirectMessages;

        config.LogFormat = (message, exception) => $"{message.Source}: {message.Message}";
    })
    .UseInteractionService((context, config) =>
    {
        config.LogLevel = DiscordLogLevel;
        config.UseCompiledLambda = true;
    })
    .UseCommandService((context, config) =>
    {
        config.LogLevel = DiscordLogLevel;
    })
    .ConfigureServices((hostContext, services) =>
    {
        services.Configure<BotConfiguration>(hostContext.Configuration);

        var botConfiguration = hostContext.Configuration.Get<BotConfiguration>();

        services.AddCache(options => options.RedisConfiguration = botConfiguration.RedisAddress);

        services.AddHereServices(builder => builder.UseKey(botConfiguration.HereApiKey));
        services.AddTransient(sp => new DarkSkyService(
            botConfiguration.PirateWeatherKey,
            baseUri: new Uri(Constants.PirateWeatherApi),
            jsonSerializerService: new DarkSkyJsonSerializerService()));

        services.AddSingleton<LocationService>();
        services.AddSingleton<WeatherService>();
        services.AddSingleton<HomeService>();

        services.AddHostedService<InteractionHandler>();
        services.AddHostedService<CommandHandler>();

        services.AddSingleton<DbContextHelper>();
        services.AddDbContext<BotDbContext>();
    })
    .Build();

await host.RunAsync();