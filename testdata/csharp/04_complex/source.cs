// 04_complex.cs
// Reflective validator that discovers attributes at runtime and builds
// expression-tree delegates for performance. Also shows async streams.

#nullable enable
using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.Linq;
using System.Linq.Expressions;
using System.Reflection;
using System.Threading.Tasks;

namespace Constructs.Complex04;

/// <summary>
/// Attribute declaring a required string length range.
/// </summary>
[AttributeUsage(AttributeTargets.Property, AllowMultiple = false)]
public sealed class StringRangeAttribute : ValidationAttribute
{
    public int Min { get; }
    public int Max { get; }

    public StringRangeAttribute(int min, int max)
        => (Min, Max) = (min, max);

    public override bool IsValid(object? value)
        => value is string s && s.Length is >= 0 and var len
                             && len >= Min && len <= Max;
}

/// <summary>
/// Builds compiled validation delegates for any annotated model using reflection.
/// </summary>
public static class ModelValidatorBuilder
{
    private delegate bool Validator<in TModel>(TModel model);

    /// <summary>
    /// Creates a highly-optimized validation function at runtime.
    /// </summary>
    public static Func<TModel, bool> Build<TModel>()
    {
        // Parameter representing the object to validate.
        var param = Expression.Parameter(typeof(TModel), "model");
        var conditions = new List<Expression>();

        foreach (var prop in typeof(TModel).GetProperties(BindingFlags.Public | BindingFlags.Instance))
        {
            var attr = prop.GetCustomAttribute<StringRangeAttribute>();
            if (attr is null) continue;

            // Generate: ((string)model.Prop).Length >= Min && ... <= Max
            var propAccess   = Expression.Property(param, prop);
            var notNull      = Expression.NotEqual(propAccess, Expression.Constant(null, typeof(string)));
            var lengthProp   = Expression.Property(propAccess, nameof(string.Length));

            var greaterEqMin = Expression.GreaterThanOrEqual(lengthProp, Expression.Constant(attr.Min));
            var lessEqMax    = Expression.LessThanOrEqual (lengthProp, Expression.Constant(attr.Max));

            var condition    = Expression.AndAlso(notNull,
                               Expression.AndAlso(greaterEqMin, lessEqMax));

            conditions.Add(condition);
        }

        var body = conditions.Any()
            ? conditions.Aggregate(Expression.AndAlso)
            : Expression.Constant(true);

        var lambda = Expression.Lambda<Func<TModel, bool>>(body, param);
        return lambda.Compile();
    }

    /// <summary>
    /// Private method to cache compiled validators
    /// </summary>
    private static readonly Dictionary<Type, Delegate> _validatorCache = new();

    /// <summary>
    /// Internal method for managing validator cache
    /// </summary>
    internal static void ClearCache() => _validatorCache.Clear();

    /// <summary>
    /// Protected method for building complex expressions
    /// </summary>
    private static Expression BuildComplexValidation<T>(ParameterExpression param, PropertyInfo property)
    {
        // Complex validation logic could go here
        return Expression.Constant(true);
    }
}

/// <summary>Sample DTO to validate.</summary>
public record SignupRequest(
    [property: StringRange(3,12)] string Username,
    [property: StringRange(8,64)] string Password)
{
    /// <summary>
    /// Additional validation method
    /// </summary>
    public bool IsPasswordStrong()
    {
        return _hasUpperCase() && _hasLowerCase() && _hasNumber();
    }

    /// <summary>
    /// Private validation helpers
    /// </summary>
    private bool _hasUpperCase() => Password.Any(char.IsUpper);
    private bool _hasLowerCase() => Password.Any(char.IsLower);
    private bool _hasNumber() => Password.Any(char.IsDigit);
}

/// <summary>Example usage with async stream.</summary>
public static class Demo
{
    public static async Task ConsumeAsync(IAsyncEnumerable<SignupRequest> inputs)
    {
        var isValid = ModelValidatorBuilder.Build<SignupRequest>();

        await foreach (var req in inputs)
        {
            Console.WriteLine(isValid(req)
                ? $"✓ Accepted {req.Username}"
                : $"✗ Rejected {req.Username}");
        }
    }

    /// <summary>
    /// Private method to generate test data
    /// </summary>
    private static async IAsyncEnumerable<SignupRequest> GenerateTestRequests()
    {
        var requests = new[]
        {
            new SignupRequest("john", "password123"),
            new SignupRequest("alice", "SecurePass1"),
            new SignupRequest("x", "short")
        };

        foreach (var req in requests)
        {
            await Task.Delay(100); // Simulate async processing
            yield return req;
        }
    }
}

/// <summary>
/// Advanced reflection-based service locator
/// </summary>
public class ServiceLocator
{
    private readonly Dictionary<Type, object> _services = new();

    /// <summary>
    /// Register a service instance
    /// </summary>
    public void Register<T>(T instance) where T : class
    {
        _services[typeof(T)] = instance;
    }

    /// <summary>
    /// Resolve a service using reflection
    /// </summary>
    public T Resolve<T>() where T : class
    {
        if (_services.TryGetValue(typeof(T), out var service))
            return (T)service;

        // Try to create instance using reflection
        return CreateInstance<T>();
    }

    /// <summary>
    /// Private factory method using reflection
    /// </summary>
    private T CreateInstance<T>() where T : class
    {
        var type = typeof(T);
        var constructors = type.GetConstructors();

        // Find parameterless constructor
        var defaultConstructor = constructors.FirstOrDefault(c => c.GetParameters().Length == 0);
        if (defaultConstructor != null)
        {
            return (T)Activator.CreateInstance(type)!;
        }

        throw new InvalidOperationException($"Cannot create instance of {type.Name}");
    }

    /// <summary>
    /// Internal service discovery
    /// </summary>
    internal Type[] GetRegisteredTypes() => _services.Keys.ToArray();

    /// <summary>
    /// Protected method for service validation
    /// </summary>
    protected virtual bool ValidateService(Type serviceType)
    {
        return serviceType.IsClass && !serviceType.IsAbstract;
    }
}