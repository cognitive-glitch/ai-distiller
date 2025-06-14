// 01_basic.cs
// Demonstrates constants, local functions, tuples and basic control-flow.

#nullable enable
namespace Constructs.Basic01;

/// <summary>
/// Collection of elementary math helpers.
/// </summary>
internal static class MathHelpers
{
    /// <summary>Mathematical π (double-precision).</summary>
    public const double Pi = 3.1415926535897931;

    /// <summary>
    /// Calculates the circumference and area of a circle.
    /// </summary>
    /// <param name="radius">Radius >= 0.</param>
    /// <returns>
    /// Tuple (circumference, area).
    /// </returns>
    /// <exception cref="ArgumentOutOfRangeException">
    /// Thrown when <paramref name="radius"/> is negative.
    /// </exception>
    public static (double C, double A) Circle(double radius)
    {
        if (radius < 0) throw new ArgumentOutOfRangeException(nameof(radius));
        var circumference = 2 * Pi * radius;
        var area          = Pi * radius * radius;
        return (circumference, area);
    }

    /// <summary>
    /// Maps an array of radii to their areas using a local function.
    /// </summary>
    public static double[] Areas(double[] radii)
    {
        if (radii is null) throw new ArgumentNullException(nameof(radii));

        // Local expression-bodied function
        double area(double r) => Pi * r * r;

        var results = new double[radii.Length];
        for (var i = 0; i < radii.Length; i++)
            results[i] = area(radii[i]);

        return results;
    }

    /// <summary>
    /// Private helper for advanced calculations
    /// </summary>
    private static double _calculateAdvanced(double input)
    {
        return input * Pi / 2;
    }

    /// <summary>
    /// Internal helper for geometry calculations
    /// </summary>
    internal static double GetVolume(double radius)
    {
        return (4.0 / 3.0) * Pi * radius * radius * radius;
    }
}

/// <summary>
/// String utility demonstrating interpolated strings and null-coalescing operators.
/// </summary>
internal static class Greeting
{
    private const string DefaultName = "world";

    public static string Hello(string? name = null)
        => $"Hello, {name ?? DefaultName}! 👋";

    /// <summary>
    /// Private method for name validation
    /// </summary>
    private static bool IsValidName(string? name)
    {
        return !string.IsNullOrWhiteSpace(name);
    }

    /// <summary>
    /// Internal method for formatting names
    /// </summary>
    internal static string FormatName(string name)
    {
        return IsValidName(name) ? name.Trim().ToTitleCase() : DefaultName;
    }
}

/// <summary>
/// Extension methods for string manipulation
/// </summary>
public static class StringExtensions
{
    /// <summary>
    /// Converts string to title case
    /// </summary>
    public static string ToTitleCase(this string input)
    {
        if (string.IsNullOrEmpty(input))
            return input;

        return char.ToUpper(input[0]) + input.Substring(1).ToLower();
    }
}