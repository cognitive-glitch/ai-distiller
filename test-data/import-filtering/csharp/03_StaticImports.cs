// Test Pattern 3: Static using Directives
// Tests import static for methods and fields

using System;
using System.Collections.Generic;
using static System.Console;
using static System.Math;
using static System.String;
using static System.Linq.Enumerable;
using static System.IO.Path;
using static System.Environment;
using static System.Convert;
using static System.DateTime;
using static System.Guid;

// Not using: String static members, Convert static members, Guid static members

namespace ImportFilteringTests
{
    public class StaticImports
    {
        public void MathCalculations()
        {
            // Using static Math members
            double radius = 5.0;
            double circumference = 2 * PI * radius;
            double area = PI * Pow(radius, 2);

            // Using static Console
            WriteLine($"Circle with radius {radius}:");
            WriteLine($"  Circumference: {circumference:F2}");
            WriteLine($"  Area: {area:F2}");

            // More Math static members
            double angle = PI / 4; // 45 degrees
            double sinValue = Sin(angle);
            double cosValue = Cos(angle);
            double tanValue = Tan(angle);

            WriteLine($"Trigonometry for 45 degrees:");
            WriteLine($"  sin: {sinValue:F4}");
            WriteLine($"  cos: {cosValue:F4}");
            WriteLine($"  tan: {tanValue:F4}");

            // Using Sqrt, Abs, Max, Min
            double number = -16.0;
            WriteLine($"Absolute value of {number}: {Abs(number)}");
            WriteLine($"Square root of {Abs(number)}: {Sqrt(Abs(number))}");
            WriteLine($"Max of 10 and 20: {Max(10, 20)}");
            WriteLine($"Min of 10 and 20: {Min(10, 20)}");
        }

        public void LinqOperations()
        {
            // Using static Enumerable members
            var numbers = Range(1, 10);  // Generates 1 to 10
            var evenNumbers = numbers.Where(n => n % 2 == 0);

            WriteLine("Even numbers from 1 to 10:");
            foreach (var num in evenNumbers)
            {
                Write($"{num} ");
            }
            WriteLine();

            // Using Repeat
            var repeated = Repeat("Hello", 3);
            WriteLine($"Repeated: {string.Join(", ", repeated)}");

            // Using Empty
            var emptyList = Empty<int>();
            WriteLine($"Empty list count: {emptyList.Count()}");
        }

        public void PathOperations()
        {
            // Using static Path members
            string fullPath = @"C:\Users\Documents\file.txt";

            WriteLine($"Full path: {fullPath}");
            WriteLine($"Directory: {GetDirectoryName(fullPath)}");
            WriteLine($"Filename: {GetFileName(fullPath)}");
            WriteLine($"Extension: {GetExtension(fullPath)}");
            WriteLine($"Filename without extension: {GetFileNameWithoutExtension(fullPath)}");

            // Using Combine
            string combined = Combine(@"C:\Users", "Documents", "file.txt");
            WriteLine($"Combined path: {combined}");

            // Using GetTempPath and GetRandomFileName
            string tempPath = GetTempPath();
            string randomFile = GetRandomFileName();
            WriteLine($"Temp path: {tempPath}");
            WriteLine($"Random filename: {randomFile}");
        }

        public void EnvironmentInfo()
        {
            // Using static Environment members
            WriteLine($"Machine name: {MachineName}");
            WriteLine($"User name: {UserName}");
            WriteLine($"OS Version: {OSVersion}");
            WriteLine($"Processor count: {ProcessorCount}");
            WriteLine($"Current directory: {CurrentDirectory}");
            WriteLine($"System directory: {SystemDirectory}");
            WriteLine($"Is 64-bit OS: {Is64BitOperatingSystem}");
            WriteLine($"Is 64-bit process: {Is64BitProcess}");

            // Using NewLine
            Write($"Line 1{NewLine}Line 2{NewLine}");
        }

        public void DateTimeOperations()
        {
            // Using static DateTime members
            var now = Now;
            var today = Today;
            var utcNow = UtcNow;

            WriteLine($"Now: {now}");
            WriteLine($"Today: {today}");
            WriteLine($"UTC Now: {utcNow}");

            // Note: Most DateTime operations are instance methods, not static
        }
    }
}