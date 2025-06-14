// 05_very_complex.cs
// High-end demo: generic-math vectors, discriminated-union style records,
// exhaustive pattern matching, default interface methods, and dependency
// injection-ready abstractions.

#nullable enable
using System;
using System.Collections.Generic;
using System.Numerics;
using Microsoft.Extensions.DependencyInjection;

namespace Constructs.VeryComplex05;

/// <summary>
/// N-dimensional immutable vector leveraging generic math (C# 11).
/// </summary>
public readonly record struct VectorN<T>(T[] Components)
    where T : INumber<T>
{
    public int Dim => Components.Length;

    public static VectorN<T> operator +(VectorN<T> a, VectorN<T> b)
    {
        if (a.Dim != b.Dim) throw new ArgumentException("Dimension mismatch.");
        var result = new T[a.Dim];
        for (var i = 0; i < a.Dim; i++)
            result[i] = a.Components[i] + b.Components[i];
        return new(result);
    }

    public T Dot(VectorN<T> other)
    {
        if (Dim != other.Dim) throw new ArgumentException("Dimension mismatch.");
        var sum = T.Zero;
        for (var i = 0; i < Dim; i++)
            sum += Components[i] * other.Components[i];
        return sum;
    }

    public override string ToString() => $"[{string.Join(", ", Components)}]";

    /// <summary>
    /// Private validation method
    /// </summary>
    private bool _isValid() => Components.Length > 0;

    /// <summary>
    /// Internal normalization method
    /// </summary>
    internal VectorN<T> Normalize() where T : IFloatingPointIeee754<T>
    {
        var magnitude = _calculateMagnitude();
        if (magnitude == T.Zero) return this;

        var normalized = new T[Dim];
        for (int i = 0; i < Dim; i++)
            normalized[i] = Components[i] / magnitude;
        
        return new(normalized);
    }

    /// <summary>
    /// Private magnitude calculation
    /// </summary>
    private T _calculateMagnitude() where T : IFloatingPointIeee754<T>
    {
        var sum = T.Zero;
        foreach (var component in Components)
            sum += component * component;
        return T.Sqrt(sum);
    }
}

/// <summary>
/// Algebra service demonstrating DI and default interface implementation.
/// </summary>
public interface IAlgebraService
{
    VectorN<T>    Add<T>(VectorN<T> a, VectorN<T> b) where T : INumber<T>
        => a + b; // default interface method (C# 8)

    T             Dot<T>(VectorN<T> a, VectorN<T> b) where T : INumber<T>
        => a.Dot(b);

    /// <summary>
    /// Advanced operation with constraints
    /// </summary>
    VectorN<T> Normalize<T>(VectorN<T> vector) where T : IFloatingPointIeee754<T>
        => vector.Normalize();
}

/// <summary>
/// Concrete implementation overriding only parts of the default interface.
/// </summary>
public sealed class AlgebraService : IAlgebraService
{
    // Override to inject logging, etc.
    public VectorN<T> Add<T>(VectorN<T> a, VectorN<T> b) where T : INumber<T>
    {
        Console.WriteLine($"Adding {a} and {b}");
        return a + b;
    }

    /// <summary>
    /// Private logging method
    /// </summary>
    private void _logOperation(string operation, object result)
    {
        Console.WriteLine($"Operation: {operation}, Result: {result}");
    }

    /// <summary>
    /// Internal performance tracking
    /// </summary>
    internal void TrackPerformance(string operation, TimeSpan duration)
    {
        _logOperation($"{operation} (Duration: {duration.TotalMilliseconds}ms)", "Completed");
    }
}

/// <summary>
/// Discriminated-union-style command model using record inheritance.
/// </summary>
public abstract record CalcCommand
{
    public sealed record AddCommand  (VectorN<double> A, VectorN<double> B) : CalcCommand;
    public sealed record DotCommand  (VectorN<double> A, VectorN<double> B) : CalcCommand;
    public sealed record NormalizeCommand(VectorN<double> Vector) : CalcCommand;
    public sealed record UnknownCommand                                   : CalcCommand;

    /// <summary>
    /// Private validation for commands
    /// </summary>
    private protected virtual bool IsValid => true;

    /// <summary>
    /// Internal command metadata
    /// </summary>
    internal DateTime CreatedAt { get; init; } = DateTime.UtcNow;
}

/// <summary>Command dispatcher using exhaustive pattern matching.</summary>
public static class CalcDispatcher
{
    public static double Execute(CalcCommand cmd, IAlgebraService svc) => cmd switch
    {
        CalcCommand.AddCommand(var a, var b) => svc.Add(a, b).Dot(new VectorN<double>(new[] {1d,1d})), // dummy use
        CalcCommand.DotCommand(var a, var b) => svc.Dot(a, b),
        CalcCommand.NormalizeCommand(var v) => svc.Normalize(v).Components.Sum(),
        _                                     => throw new NotSupportedException($"Unhandled {cmd.GetType().Name}")
    };

    /// <summary>
    /// Private command validation
    /// </summary>
    private static bool ValidateCommand(CalcCommand command)
    {
        return command switch
        {
            CalcCommand.AddCommand(var a, var b) => a.Dim == b.Dim,
            CalcCommand.DotCommand(var a, var b) => a.Dim == b.Dim,
            CalcCommand.NormalizeCommand(var v) => v.Dim > 0,
            _ => false
        };
    }

    /// <summary>
    /// Internal command processing statistics
    /// </summary>
    internal static Dictionary<Type, int> GetCommandStats()
    {
        // In a real implementation, this would track command execution
        return new Dictionary<Type, int>();
    }
}

/// <summary>
/// Advanced generic constraint demonstration
/// </summary>
public class MathProcessor<T> where T : INumber<T>, IMinMaxValue<T>
{
    /// <summary>
    /// Process values with advanced constraints
    /// </summary>
    public T ProcessValue(T input)
    {
        if (input > T.MaxValue / (T.One + T.One))
            return T.MaxValue;
        if (input < T.MinValue / (T.One + T.One))
            return T.MinValue;
        
        return input * (T.One + T.One);
    }

    /// <summary>
    /// Private bounds checking
    /// </summary>
    private static bool _isInBounds(T value)
    {
        return value >= T.MinValue && value <= T.MaxValue;
    }

    /// <summary>
    /// Protected virtual method for derived classes
    /// </summary>
    protected virtual T TransformValue(T input)
    {
        return _isInBounds(input) ? input : T.Zero;
    }

    /// <summary>
    /// Internal factory method
    /// </summary>
    internal static MathProcessor<T> CreateProcessor()
    {
        return new MathProcessor<T>();
    }
}

/// <summary>
/// Bootstrapper illustrating Microsoft.Extensions.DependencyInjection usage.
/// </summary>
public static class Program
{
    public static void Main()
    {
        // Build DI container.
        using var provider = new ServiceCollection()
            .AddSingleton<IAlgebraService, AlgebraService>()
            .AddTransient<MathProcessor<double>>()
            .BuildServiceProvider();

        var algebra = provider.GetRequiredService<IAlgebraService>();

        // Create some vectors.
        var a = new VectorN<double>(new[] {1.0, 2.0});
        var b = new VectorN<double>(new[] {3.5, -1.0});

        // Issue commands.
        IReadOnlyList<CalcCommand> commands = new CalcCommand[]
        {
            new CalcCommand.AddCommand(a, b),
            new CalcCommand.DotCommand(a, b),
            new CalcCommand.NormalizeCommand(a),
            new CalcCommand.UnknownCommand()
        };

        foreach (var cmd in commands)
        {
            try
            {
                var result = CalcDispatcher.Execute(cmd, algebra);
                Console.WriteLine($"Result = {result}");
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex.Message);
            }
        }
    }

    /// <summary>
    /// Private initialization helper
    /// </summary>
    private static IServiceCollection ConfigureServices()
    {
        return new ServiceCollection()
            .AddSingleton<IAlgebraService, AlgebraService>();
    }

    /// <summary>
    /// Internal configuration method
    /// </summary>
    internal static void ConfigureLogging(IServiceCollection services)
    {
        // Logging configuration would go here
    }
}