using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
global using System.Text.Json;

// File-scoped namespace (C# 10+)
namespace MyApp.Services;

// Traditional namespace
namespace MyApp.Models
{
    // Record (C# 9+)
    public record Person(string FirstName, string LastName, int Age)
    {
        public string FullName => $"{FirstName} {LastName}";
        
        // Init-only property (C# 9+)
        public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    }
    
    // Record struct (C# 10+)
    public record struct Point(double X, double Y);
    
    // Interface with default implementation (C# 8+)
    public interface IService
    {
        string Name { get; }
        
        void Execute();
        
        // Default interface method
        public void Log(string message) => Console.WriteLine($"[{Name}]: {message}");
        
        // Static abstract member (C# 11+)
        static abstract IService CreateInstance();
    }
    
    // Generic interface
    public interface IRepository<T> where T : class
    {
        Task<T?> GetByIdAsync(int id);
        Task<IEnumerable<T>> GetAllAsync();
        Task AddAsync(T entity);
        Task UpdateAsync(T entity);
        Task DeleteAsync(int id);
    }
    
    // Delegate
    public delegate void NotificationHandler(string message);
    public delegate TResult Func<in T, out TResult>(T arg);
    
    // Abstract base class
    public abstract class ServiceBase : IService
    {
        public abstract string Name { get; }
        
        public abstract void Execute();
        
        protected virtual void OnExecuting() { }
        protected virtual void OnExecuted() { }
    }
    
    // Sealed class with events
    public sealed class UserService : ServiceBase, IRepository<User>
    {
        private readonly ILogger _logger;
        private readonly List<User> _users = new();
        
        // Events
        public event NotificationHandler? UserCreated;
        public event EventHandler<UserEventArgs>? UserDeleted;
        
        // Auto-implemented property
        public override string Name { get; } = "UserService";
        
        // Expression-bodied property
        public int UserCount => _users.Count;
        
        // Constructor
        public UserService(ILogger logger)
        {
            _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        }
        
        // Override method
        public override void Execute()
        {
            OnExecuting();
            _logger.LogInformation("Executing user service");
            OnExecuted();
        }
        
        // Async methods
        public async Task<User?> GetByIdAsync(int id)
        {
            await Task.Delay(100); // Simulate async work
            return _users.FirstOrDefault(u => u.Id == id);
        }
        
        public async Task<IEnumerable<User>> GetAllAsync()
        {
            await Task.Delay(100);
            return _users.ToList();
        }
        
        public async Task AddAsync(User entity)
        {
            _users.Add(entity);
            UserCreated?.Invoke($"User {entity.Name} created");
            await Task.CompletedTask;
        }
        
        public Task UpdateAsync(User entity) => Task.CompletedTask;
        
        public Task DeleteAsync(int id)
        {
            var user = _users.FirstOrDefault(u => u.Id == id);
            if (user != null)
            {
                _users.Remove(user);
                UserDeleted?.Invoke(this, new UserEventArgs(user));
            }
            return Task.CompletedTask;
        }
        
        // Static factory method
        public static IService CreateInstance() => new UserService(new ConsoleLogger());
        
        // Operator overloading
        public static UserService operator +(UserService service, User user)
        {
            service._users.Add(user);
            return service;
        }
    }
    
    // Class with indexer
    public class UserCollection
    {
        private readonly Dictionary<int, User> _users = new();
        
        // Indexer
        public User? this[int id]
        {
            get => _users.TryGetValue(id, out var user) ? user : null;
            set
            {
                if (value != null)
                    _users[id] = value;
            }
        }
        
        // Indexer with multiple parameters
        public User? this[string name, int age] => 
            _users.Values.FirstOrDefault(u => u.Name == name && u.Age == age);
    }
    
    // Struct (value type)
    public struct Temperature
    {
        public double Celsius { get; init; }
        
        public double Fahrenheit => Celsius * 9 / 5 + 32;
        
        public Temperature(double celsius)
        {
            Celsius = celsius;
        }
        
        // User-defined conversion
        public static implicit operator double(Temperature temp) => temp.Celsius;
        public static explicit operator Temperature(double celsius) => new(celsius);
    }
    
    // Enum
    public enum UserRole
    {
        Guest = 0,
        User = 1,
        Moderator = 2,
        Admin = 4,
        SuperAdmin = Admin | 8
    }
    
    // Class with attributes
    [Serializable]
    [Obsolete("Use Person record instead")]
    public class User
    {
        // Fields with different access modifiers
        private readonly int _id;
        protected string _internalName;
        internal static int _userCount;
        
        // Properties
        public int Id { get => _id; }
        public string Name { get; set; }
        public int Age { get; private set; }
        public UserRole Role { get; init; }
        
        // Nullable reference type (C# 8+)
        public string? Email { get; set; }
        
        // Constructor with optional parameter
        public User(int id, string name, int age = 0)
        {
            _id = id;
            Name = name;
            Age = age;
            _userCount++;
        }
        
        // Destructor/Finalizer
        ~User()
        {
            _userCount--;
        }
        
        // Method with ref/out parameters
        public bool TryUpdateAge(ref int newAge, out string error)
        {
            if (newAge < 0)
            {
                error = "Age cannot be negative";
                return false;
            }
            
            Age = newAge;
            error = string.Empty;
            return true;
        }
        
        // Method with params
        public void AssignRoles(params UserRole[] roles)
        {
            Role = roles.Aggregate((a, b) => a | b);
        }
    }
    
    // Extension methods
    public static class UserExtensions
    {
        public static bool IsAdult(this User user) => user.Age >= 18;
        
        public static string GetDisplayName(this User user, bool includeAge = false)
        {
            return includeAge ? $"{user.Name} ({user.Age})" : user.Name;
        }
    }
    
    // Generic class with constraints
    public class Cache<TKey, TValue> 
        where TKey : notnull
        where TValue : class?, new()
    {
        private readonly Dictionary<TKey, TValue> _cache = new();
        
        public void Add(TKey key, TValue value) => _cache[key] = value;
        
        public TValue GetOrCreate(TKey key)
        {
            if (!_cache.TryGetValue(key, out var value))
            {
                value = new TValue();
                _cache[key] = value;
            }
            return value;
        }
    }
    
    // Partial class (part 1)
    public partial class DataService
    {
        private readonly string _connectionString;
        
        public DataService(string connectionString)
        {
            _connectionString = connectionString;
        }
    }
    
    // Partial class (part 2)
    public partial class DataService
    {
        public async Task<T?> QueryAsync<T>(string query) where T : class
        {
            // Implementation
            await Task.Delay(100);
            return default;
        }
    }
    
    // Exception class
    public class ServiceException : Exception
    {
        public int ErrorCode { get; }
        
        public ServiceException(string message, int errorCode) : base(message)
        {
            ErrorCode = errorCode;
        }
        
        public ServiceException(string message, int errorCode, Exception innerException) 
            : base(message, innerException)
        {
            ErrorCode = errorCode;
        }
    }
    
    // Event args
    public class UserEventArgs : EventArgs
    {
        public User User { get; }
        public DateTime Timestamp { get; }
        
        public UserEventArgs(User user)
        {
            User = user;
            Timestamp = DateTime.UtcNow;
        }
    }
    
    // Logger interface
    public interface ILogger
    {
        void LogInformation(string message);
        void LogError(string message, Exception? exception = null);
    }
    
    // Console logger implementation
    public class ConsoleLogger : ILogger
    {
        public void LogInformation(string message) => 
            Console.WriteLine($"[INFO] {DateTime.Now:yyyy-MM-dd HH:mm:ss} {message}");
        
        public void LogError(string message, Exception? exception = null)
        {
            Console.WriteLine($"[ERROR] {DateTime.Now:yyyy-MM-dd HH:mm:ss} {message}");
            if (exception != null)
                Console.WriteLine(exception.ToString());
        }
    }
    
    // Static class
    public static class StringHelpers
    {
        public static string Reverse(string input)
        {
            return new string(input.Reverse().ToArray());
        }
        
        public static bool IsNullOrEmpty(string? value) => string.IsNullOrEmpty(value);
    }
    
    // Nested types
    public class Container
    {
        // Nested class
        public class NestedClass
        {
            public void Method() { }
        }
        
        // Nested interface
        public interface INestedInterface
        {
            void DoSomething();
        }
        
        // Nested struct
        public struct NestedStruct
        {
            public int Value { get; set; }
        }
        
        // Nested enum
        public enum NestedEnum
        {
            Option1,
            Option2
        }
        
        // Nested delegate
        public delegate void NestedDelegate(int value);
    }
}