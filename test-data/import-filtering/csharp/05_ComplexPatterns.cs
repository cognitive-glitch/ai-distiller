// Test Pattern 5: Complex Import Patterns
// Tests file-scoped namespaces, nested types, generics, and advanced scenarios

using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System.Threading.Channels;
using System.Collections.Concurrent;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using System.Reflection;
using System.Linq.Expressions;
using System.Dynamic;

// File-scoped namespace (C# 10+)
namespace ImportFilteringTests.Advanced;

// Not using: Threading.Channels, Reactive.Subjects, Options, Dynamic

public interface IRepository<T> where T : class
{
    Task<T> GetByIdAsync(int id);
    Task<IEnumerable<T>> GetAllAsync();
}

// Using generics and expressions
public class Repository<T> : IRepository<T> where T : class
{
    private readonly ILogger<Repository<T>> logger;
    private readonly ConcurrentDictionary<int, T> cache = new();

    public Repository(ILogger<Repository<T>> logger)
    {
        this.logger = logger;
    }

    public async Task<T> GetByIdAsync(int id)
    {
        // Using ConcurrentDictionary
        if (cache.TryGetValue(id, out var cached))
        {
            logger.LogInformation("Cache hit for id {Id}", id);
            return cached;
        }

        // Simulate async operation
        await Task.Delay(100);

        // Using Reflection to create instance
        var instance = Activator.CreateInstance<T>();
        var idProperty = typeof(T).GetProperty("Id");
        idProperty?.SetValue(instance, id);

        cache.TryAdd(id, instance);
        return instance;
    }

    public async Task<IEnumerable<T>> GetAllAsync()
    {
        // Using LINQ
        return await Task.FromResult(cache.Values.ToList());
    }

    // Using Expression trees
    public IQueryable<T> Query(Expression<Func<T, bool>> predicate)
    {
        return cache.Values.AsQueryable().Where(predicate);
    }
}

// Using Reactive Extensions
public class EventAggregator
{
    private readonly Subject<object> subject = new();

    public IObservable<T> GetEvent<T>()
    {
        // Using Reactive LINQ
        return subject.OfType<T>();
    }

    public void Publish<T>(T eventData)
    {
        subject.OnNext(eventData);
    }
}

// Nested types with imports usage
public class DataProcessor
{
    // Nested class using imports
    public class ProcessingResult
    {
        public DateTime Timestamp { get; set; } = DateTime.UtcNow;
        public Dictionary<string, object> Data { get; set; } = new();
        public List<string> Errors { get; set; } = new();
    }

    // Nested interface
    public interface IProcessingStrategy
    {
        Task<ProcessingResult> ProcessAsync(byte[] data);
    }

    // Using DI container
    private readonly IServiceProvider serviceProvider;
    private readonly ILogger<DataProcessor> logger;

    public DataProcessor(IServiceProvider serviceProvider, ILogger<DataProcessor> logger)
    {
        this.serviceProvider = serviceProvider;
        this.logger = logger;
    }

    public async Task<ProcessingResult> ProcessWithStrategyAsync<TStrategy>(byte[] data)
        where TStrategy : IProcessingStrategy
    {
        // Using DI to resolve strategy
        var strategy = serviceProvider.GetRequiredService<TStrategy>();

        logger.LogInformation("Processing {Bytes} bytes with {Strategy}",
            data.Length, typeof(TStrategy).Name);

        try
        {
            return await strategy.ProcessAsync(data);
        }
        catch (Exception ex)
        {
            logger.LogError(ex, "Processing failed");
            return new ProcessingResult
            {
                Errors = new List<string> { ex.Message }
            };
        }
    }
}

// Extension methods using imports
public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddCustomServices(this IServiceCollection services)
    {
        // Using Microsoft.Extensions.DependencyInjection
        services.AddSingleton<EventAggregator>();
        services.AddScoped(typeof(IRepository<>), typeof(Repository<>));
        services.AddTransient<DataProcessor>();

        // Using Microsoft.Extensions.Logging
        services.AddLogging(builder =>
        {
            builder.AddConsole();
            builder.AddDebug();
        });

        return services;
    }
}

// Async enumerable pattern
public class DataStream
{
    public async IAsyncEnumerable<int> GenerateNumbersAsync()
    {
        for (int i = 0; i < 100; i++)
        {
            await Task.Delay(10);
            yield return i;
        }
    }

    public async Task ConsumeAsync()
    {
        // Using async LINQ
        await foreach (var number in GenerateNumbersAsync().Where(n => n % 2 == 0))
        {
            Console.WriteLine($"Even number: {number}");
        }
    }
}