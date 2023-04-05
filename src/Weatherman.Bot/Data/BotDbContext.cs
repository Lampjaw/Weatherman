using Microsoft.EntityFrameworkCore;
using Weatherman.Bot.Data.Models;

namespace Weatherman.Bot.Data
{
    public class BotDbContext : DbContext
    {
        public string DbPath { get; }

        public BotDbContext()
        {
            var currentPath = Path.GetDirectoryName(Environment.CurrentDirectory);
            var dataPath = Path.Join(currentPath, "/data");
            Directory.CreateDirectory(dataPath);

            DbPath = Path.Combine(dataPath, "weatherplugin.db");

            Database.EnsureCreated();
        }

        protected override void OnConfiguring(DbContextOptionsBuilder options)
            => options.UseSqlite($"Data Source={DbPath}");

        public DbSet<UserProfile> UserProfiles { get; set; }
    }
}
