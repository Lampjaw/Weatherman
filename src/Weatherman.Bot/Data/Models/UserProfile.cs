using System.ComponentModel.DataAnnotations.Schema;

namespace Weatherman.Bot.Data.Models
{
    [Table("user_profile")]
    public class UserProfile
    {
        public string Id { get; set; }
        public string HomeLocation { get; set; }
        public DateTime HomeLocationChangedDate { get; set; }
    }
}
