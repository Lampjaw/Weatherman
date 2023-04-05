using Newtonsoft.Json;
using Serilog.Core;
using Serilog.Debugging;
using Serilog.Events;
using Serilog.Sinks.Graylog;
using Serilog.Sinks.Graylog.Core;

namespace Weatherman.Bot
{
    public class GraylogConsoleSink : ILogEventSink
    {
        private readonly Lazy<IGelfConverter> _converter;
        private readonly JsonSerializer _serializer;

        public GraylogConsoleSink()
            : this(new GraylogSinkOptions())
        {
        }

        public GraylogConsoleSink(GraylogSinkOptions options)
        {
            _serializer = JsonSerializer.Create(options.SerializerSettings);
            ISinkComponentsBuilder sinkComponentsBuilder = new SinkComponentsBuilder(options);
            _converter = new Lazy<IGelfConverter>(() => sinkComponentsBuilder.MakeGelfConverter());
        }

        public void Emit(LogEvent logEvent)
        {
            try
            {
                var json = _converter.Value.GetGelfJson(logEvent);

                using var textWriter = new StringWriter();
                {
                    _serializer.Serialize(textWriter, json);
                    Console.WriteLine(textWriter.ToString());
                }
            }
            catch (Exception exc)
            {
                SelfLog.WriteLine("Oops something going wrong {0}", exc);
            }
        }
    }
}
