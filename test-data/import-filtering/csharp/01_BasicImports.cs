// Test Pattern 1: Basic using Directives
// Tests standard namespace imports and their usage

using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.IO;
using System.Net.Http;
using System.Text.RegularExpressions;
using System.Diagnostics;
using System.Reflection;

// Not using: System.Text, System.Threading.Tasks, System.Reflection

namespace ImportFilteringTests
{
    public class BasicImports
    {
        private List<string> data = new List<string>();
        
        public void ProcessData()
        {
            // Using System for Console
            Console.WriteLine("Processing data...");
            
            // Using System.Collections.Generic
            data.Add("apple");
            data.Add("banana");
            data.Add("cherry");
            
            // Using System.Linq
            var sortedData = data.OrderBy(x => x).ToList();
            var filtered = data.Where(x => x.StartsWith("b")).ToList();
            
            Console.WriteLine("Sorted data:");
            foreach (var item in sortedData)
            {
                Console.WriteLine($"  {item}");
            }
            
            // Using System.IO
            string filePath = "data.txt";
            File.WriteAllLines(filePath, sortedData);
            
            if (File.Exists(filePath))
            {
                var lines = File.ReadAllLines(filePath);
                Console.WriteLine($"Read {lines.Length} lines from file");
            }
        }
        
        public async void FetchData()
        {
            // Using System.Net.Http
            using (var client = new HttpClient())
            {
                try
                {
                    var response = await client.GetStringAsync("https://api.example.com/data");
                    Console.WriteLine($"Received: {response}");
                }
                catch (HttpRequestException e)
                {
                    Console.WriteLine($"Error: {e.Message}");
                }
            }
        }
        
        public bool ValidateEmail(string email)
        {
            // Using System.Text.RegularExpressions
            string pattern = @"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$";
            return Regex.IsMatch(email, pattern);
        }
        
        public void MeasurePerformance()
        {
            // Using System.Diagnostics
            var stopwatch = Stopwatch.StartNew();
            
            // Simulate some work
            for (int i = 0; i < 1000000; i++)
            {
                var temp = i * 2;
            }
            
            stopwatch.Stop();
            Console.WriteLine($"Elapsed time: {stopwatch.ElapsedMilliseconds} ms");
        }
    }
}