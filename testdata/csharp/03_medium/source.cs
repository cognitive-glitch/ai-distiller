// 03_medium.cs
// Generic repository pattern with asynchronous LINQ queries and inheritance.

#nullable enable
using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;

namespace Constructs.Medium03;

/// <summary>
/// Contract describing CRUD operations for entities.
/// </summary>
public interface IRepository<TEntity, TKey>
    where TEntity : IEntity<TKey>
    where TKey     : notnull
{
    Task AddAsync    (TEntity entity, CancellationToken ct = default);
    Task<TEntity?> GetAsync(TKey id, CancellationToken ct = default);
    Task<bool>   RemoveAsync(TKey id, CancellationToken ct = default);
    IAsyncEnumerable<TEntity> QueryAsync(Func<TEntity, bool> predicate,
                                         CancellationToken ct = default);
}

/// <summary>Marker for entities.</summary>
public interface IEntity<out TKey>
{
    TKey Id { get; }
}

/// <summary>Base class implementing common <see cref="IEntity{TKey}"/> plumbing.</summary>
public abstract record EntityBase<T>(T Id) : IEntity<T>;

/// <summary>
/// Thread-safe in-memory repository using <see cref="ConcurrentDictionary{TKey,TValue}"/>.
/// </summary>
public class InMemoryRepository<TEntity, TKey> :
    IRepository<TEntity, TKey>
    where TEntity : EntityBase<TKey>
    where TKey     : notnull
{
    private readonly ConcurrentDictionary<TKey, TEntity> _store = new();

    public Task AddAsync(TEntity entity, CancellationToken ct = default)
    {
        _store[entity.Id] = entity;
        return Task.CompletedTask;
    }

    public Task<TEntity?> GetAsync(TKey id, CancellationToken ct = default)
        => Task.FromResult(_store.TryGetValue(id, out var e) ? e : null);

    public Task<bool> RemoveAsync(TKey id, CancellationToken ct = default)
        => Task.FromResult(_store.TryRemove(id, out _));

    public async IAsyncEnumerable<TEntity> QueryAsync(
        Func<TEntity, bool> predicate,
        [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken ct = default)
    {
        // LINQ + async enumeration.
        foreach (var entity in _store.Values.Where(predicate))
        {
            ct.ThrowIfCancellationRequested();
            await Task.Yield(); // Simulate async I/O latency.
            yield return entity;
        }
    }

    /// <summary>
    /// Private method for internal maintenance
    /// </summary>
    private void CleanupExpiredEntities()
    {
        // Cleanup logic would go here
    }

    /// <summary>
    /// Protected method for derived classes
    /// </summary>
    protected virtual bool ValidateEntity(TEntity entity)
    {
        return entity?.Id != null;
    }

    /// <summary>
    /// Internal method for monitoring
    /// </summary>
    internal int GetEntityCount() => _store.Count;
}

/// <summary>
/// Cached repository decorator implementing caching layer
/// </summary>
public class CachedRepository<TEntity, TKey> : IRepository<TEntity, TKey>
    where TEntity : EntityBase<TKey>
    where TKey : notnull
{
    private readonly IRepository<TEntity, TKey> _innerRepository;
    private readonly ConcurrentDictionary<TKey, TEntity> _cache = new();
    private readonly TimeSpan _cacheExpiry;

    public CachedRepository(IRepository<TEntity, TKey> innerRepository, TimeSpan cacheExpiry)
    {
        _innerRepository = innerRepository;
        _cacheExpiry = cacheExpiry;
    }

    public async Task AddAsync(TEntity entity, CancellationToken ct = default)
    {
        await _innerRepository.AddAsync(entity, ct);
        _cache[entity.Id] = entity;
    }

    public async Task<TEntity?> GetAsync(TKey id, CancellationToken ct = default)
    {
        if (_cache.TryGetValue(id, out var cached))
            return cached;

        var entity = await _innerRepository.GetAsync(id, ct);
        if (entity != null)
            _cache[id] = entity;

        return entity;
    }

    public async Task<bool> RemoveAsync(TKey id, CancellationToken ct = default)
    {
        _cache.TryRemove(id, out _);
        return await _innerRepository.RemoveAsync(id, ct);
    }

    public IAsyncEnumerable<TEntity> QueryAsync(Func<TEntity, bool> predicate, CancellationToken ct = default)
        => _innerRepository.QueryAsync(predicate, ct);

    /// <summary>
    /// Private cache management
    /// </summary>
    private void InvalidateCache()
    {
        _cache.Clear();
    }

    /// <summary>
    /// Internal cache statistics
    /// </summary>
    internal (int CacheSize, int RepositorySize) GetStats()
    {
        var repoSize = _innerRepository is InMemoryRepository<TEntity, TKey> inMem
            ? inMem.GetEntityCount()
            : 0;
        return (_cache.Count, repoSize);
    }
}

/// <summary>
/// Example entity for testing
/// </summary>
public record User(Guid Id, string Name, string Email) : EntityBase<Guid>(Id)
{
    /// <summary>
    /// Checks if user is valid
    /// </summary>
    public bool IsValid => !string.IsNullOrEmpty(Name) && Email.Contains("@");

    /// <summary>
    /// Private helper for email validation
    /// </summary>
    private bool ValidateEmail() => Email.Contains("@") && Email.Contains(".");
}

/// <summary>
/// Service layer demonstrating dependency injection pattern
/// </summary>
public class UserService
{
    private readonly IRepository<User, Guid> _userRepository;

    public UserService(IRepository<User, Guid> userRepository)
    {
        _userRepository = userRepository;
    }

    public async Task<User?> CreateUserAsync(string name, string email)
    {
        var user = new User(Guid.NewGuid(), name, email);

        if (!user.IsValid)
            return null;

        await _userRepository.AddAsync(user);
        return user;
    }

    public Task<User?> GetUserAsync(Guid id) => _userRepository.GetAsync(id);

    /// <summary>
    /// Private validation logic
    /// </summary>
    private static bool IsValidUserData(string name, string email)
    {
        return !string.IsNullOrWhiteSpace(name) &&
               !string.IsNullOrWhiteSpace(email) &&
               email.Contains("@");
    }
}