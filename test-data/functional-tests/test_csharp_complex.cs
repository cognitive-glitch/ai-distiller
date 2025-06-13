// Complex C# test file for AI Distiller functional testing
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using static System.Console;

namespace ComplexCSharpNamespace.Services
{
    /// <summary>
    /// Complex C# class demonstrating various language features
    /// </summary>
    [Serializable]
    [Obsolete("This is a test class")]
    public abstract class BaseService<T> : IDisposable, IComparable<BaseService<T>>
        where T : class, new()
    {
        // Static fields
        public static readonly string DefaultMessage = "Default message";
        private static readonly Dictionary<Type, object> _instances = new();
        
        // Instance fields
        [NonSerialized]
        private readonly IServiceProvider _serviceProvider;
        
        protected internal string Name { get; set; }
        public DateTime CreatedAt { get; private set; }
        
        // Auto-implemented properties
        public int Id { get; init; }
        public string? Description { get; set; }
        
        // Expression-bodied property
        public bool IsValid => !string.IsNullOrEmpty(Name) && CreatedAt != default;
        
        // Events
        public event EventHandler<ServiceEventArgs>? ServiceChanged;
        public event Action<string>? MessageReceived;
        
        // Constructor
        protected BaseService(IServiceProvider serviceProvider, string name)
        {
            _serviceProvider = serviceProvider ?? throw new ArgumentNullException(nameof(serviceProvider));
            Name = name;
            CreatedAt = DateTime.UtcNow;
        }
        
        // Abstract method
        protected abstract Task<T> CreateInstanceAsync();
        
        // Virtual method
        public virtual async Task<bool> ProcessAsync(T item)
        {
            if (item == null) return false;
            
            await ValidateAsync(item);
            OnServiceChanged(new ServiceEventArgs { Item = item });
            return true;
        }
        
        // Sealed override method
        public sealed override string ToString()
        {
            return $"{GetType().Name}: {Name}";
        }
        
        // Generic method with constraints
        public TResult Transform<TResult>(T input, Func<T, TResult> transformer)
            where TResult : class
        {
            return transformer(input);
        }
        
        // Extension method placeholder (would be in static class)
        // public static bool IsNullOrEmpty<T>(this T? value) where T : class
        
        // Protected method
        protected virtual async Task ValidateAsync(T item)
        {
            await Task.Delay(10); // Simulate async validation
            ArgumentNullException.ThrowIfNull(item);
        }
        
        // Private method
        private void OnServiceChanged(ServiceEventArgs args)
        {
            ServiceChanged?.Invoke(this, args);
        }
        
        // Explicit interface implementation
        public abstract int CompareTo(BaseService<T>? other);
        
        // IDisposable implementation
        public void Dispose()
        {
            Dispose(true);
            GC.SuppressFinalize(this);
        }
        
        protected virtual void Dispose(bool disposing)
        {
            if (disposing)
            {
                // Dispose managed resources
            }
        }
        
        // Static method
        public static BaseService<T> GetInstance(Type serviceType)
        {
            return _instances.TryGetValue(serviceType, out var instance) 
                ? (BaseService<T>)instance 
                : throw new InvalidOperationException($"No instance found for {serviceType}");
        }
    }
    
    // Concrete implementation
    public sealed class UserService : BaseService<User>, IUserRepository
    {
        private readonly List<User> _users = new();
        
        public UserService(IServiceProvider serviceProvider) 
            : base(serviceProvider, "UserService")
        {
        }
        
        protected override async Task<User> CreateInstanceAsync()
        {
            await Task.Delay(1);
            return new User { Id = Random.Shared.Next(1000, 9999) };
        }
        
        public override int CompareTo(BaseService<User>? other)
        {
            return string.Compare(Name, other?.Name, StringComparison.OrdinalIgnoreCase);
        }
        
        public async Task<User?> FindByIdAsync(int id)
        {
            await Task.Delay(1);
            return _users.FirstOrDefault(u => u.Id == id);
        }
        
        public async Task<bool> SaveAsync(User user)
        {
            if (!await ProcessAsync(user)) return false;
            
            _users.Add(user);
            return true;
        }
        
        // Local function example
        public IEnumerable<User> GetActiveUsers()
        {
            bool IsActive(User user) => user.IsActive && user.LastLoginDate > DateTime.Now.AddDays(-30);
            
            return _users.Where(IsActive);
        }
    }
    
    // Interface
    public interface IUserRepository
    {
        Task<User?> FindByIdAsync(int id);
        Task<bool> SaveAsync(User user);
        
        // Default interface method (C# 8.0+)
        bool CanProcess(User user) => user != null && user.Id > 0;
    }
    
    // Record (C# 9.0+)
    public record User(int Id, string Name, string Email)
    {
        public bool IsActive { get; init; } = true;
        public DateTime LastLoginDate { get; set; } = DateTime.UtcNow;
        
        // Record with validation
        public User() : this(0, string.Empty, string.Empty)
        {
        }
        
        public void Deconstruct(out int id, out string name)
        {
            id = Id;
            name = Name;
        }
    }
    
    // Record struct (C# 10.0+)
    public readonly record struct Point(double X, double Y)
    {
        public static Point Origin => new(0, 0);
        
        public double DistanceFromOrigin => Math.Sqrt(X * X + Y * Y);
    }
    
    // Struct with interface implementation
    public struct ComplexNumber : IEquatable<ComplexNumber>
    {
        public double Real { get; init; }
        public double Imaginary { get; init; }
        
        public ComplexNumber(double real, double imaginary)
        {
            Real = real;
            Imaginary = imaginary;
        }
        
        public static ComplexNumber operator +(ComplexNumber a, ComplexNumber b)
        {
            return new ComplexNumber(a.Real + b.Real, a.Imaginary + b.Imaginary);
        }
        
        public bool Equals(ComplexNumber other)
        {
            return Real.Equals(other.Real) && Imaginary.Equals(other.Imaginary);
        }
        
        public override bool Equals(object? obj)
        {
            return obj is ComplexNumber other && Equals(other);
        }
        
        public override int GetHashCode()
        {
            return HashCode.Combine(Real, Imaginary);
        }
    }
    
    // Enum with underlying type
    [Flags]
    public enum UserRole : byte
    {
        None = 0,
        User = 1,
        Moderator = 2,
        Admin = 4,
        SuperAdmin = 8,
        All = User | Moderator | Admin | SuperAdmin
    }
    
    // Enum with methods (extension methods would be in static class)
    public enum Status
    {
        [Description("Not started")]
        NotStarted,
        
        [Description("In progress")]
        InProgress,
        
        [Description("Completed successfully")]
        Completed,
        
        [Description("Failed with errors")]
        Failed
    }
    
    // Delegate declarations
    public delegate Task<bool> ValidationDelegate<in T>(T item);
    public delegate TResult TransformDelegate<in T, out TResult>(T input);
    
    // Generic class with multiple constraints
    public class Repository<TEntity, TKey> : IDisposable
        where TEntity : class, IEntity<TKey>, new()
        where TKey : struct, IComparable<TKey>
    {
        private readonly Dictionary<TKey, TEntity> _storage = new();
        private bool _disposed;
        
        public void Add(TEntity entity)
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            _storage[entity.Id] = entity;
        }
        
        public TEntity? Find(TKey id)
        {
            _storage.TryGetValue(id, out var entity);
            return entity;
        }
        
        public void Dispose()
        {
            _disposed = true;
        }
    }
    
    // Nested classes
    public class OuterClass
    {
        private readonly string _outerField;
        
        public OuterClass(string outerField)
        {
            _outerField = outerField;
        }
        
        public class NestedClass
        {
            public string NestedProperty { get; set; } = string.Empty;
            
            public void AccessOuter(OuterClass outer)
            {
                // Can access private members of outer class
                WriteLine(outer._outerField);
            }
        }
        
        private class PrivateNestedClass
        {
            internal void InternalMethod() { }
        }
    }
    
    // Partial class
    public partial class PartialClass
    {
        partial void OnPropertyChanged(string propertyName);
        
        public string Property1 { get; set; } = string.Empty;
    }
    
    // Event arguments
    public class ServiceEventArgs : EventArgs
    {
        public object? Item { get; set; }
        public string Message { get; set; } = string.Empty;
    }
    
    // Attribute
    [AttributeUsage(AttributeTargets.Property | AttributeTargets.Field)]
    public class DescriptionAttribute : Attribute
    {
        public string Description { get; }
        
        public DescriptionAttribute(string description)
        {
            Description = description;
        }
    }
    
    // Interface with generic constraints
    public interface IEntity<out T> where T : struct
    {
        T Id { get; }
    }
}

// File-scoped namespace (C# 10.0+)
namespace ComplexCSharpNamespace.Extensions;

public static class StringExtensions
{
    public static bool IsNullOrWhiteSpace(this string? value)
    {
        return string.IsNullOrWhiteSpace(value);
    }
    
    public static string Truncate(this string value, int maxLength)
    {
        return value.Length <= maxLength ? value : value[..maxLength];
    }
}

// Global using example (would typically be in GlobalUsings.cs)
global using GlobalAlias = System.Collections.Generic.Dictionary<string, object>;