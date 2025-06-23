// Test Pattern 2: Aliased using Directives
// Tests aliasing namespaces and types

using System;
using System.Collections.Generic;
using MyDict = System.Collections.Generic.Dictionary<string, object>;
using JsonConvert = Newtonsoft.Json.JsonConvert;
using HttpClient = System.Net.Http.HttpClient;
using IO = System.IO;
using Regex = System.Text.RegularExpressions.Regex;
using MyUser = MyCompany.Models.User;
using MyLogger = MyCompany.Services.Logger;
using Threading = System.Threading;
using Reflection = System.Reflection;

// Not using: MyLogger, Threading alias, Reflection alias

namespace ImportFilteringTests
{
    public class AliasedImports
    {
        private MyDict configuration;
        private HttpClient httpClient;
        
        public AliasedImports()
        {
            // Using aliased Dictionary
            configuration = new MyDict
            {
                ["apiUrl"] = "https://api.example.com",
                ["timeout"] = 30,
                ["retryCount"] = 3
            };
            
            // Using aliased HttpClient
            httpClient = new HttpClient();
        }
        
        public string SerializeUser(MyUser user)
        {
            // Using aliased JsonConvert (from Newtonsoft.Json)
            return JsonConvert.SerializeObject(user, Formatting.Indented);
        }
        
        public MyUser DeserializeUser(string json)
        {
            // Using JsonConvert alias again
            return JsonConvert.DeserializeObject<MyUser>(json);
        }
        
        public void FileOperations()
        {
            // Using IO alias for System.IO
            string path = @"C:\temp\data.txt";
            
            if (IO.File.Exists(path))
            {
                string content = IO.File.ReadAllText(path);
                Console.WriteLine($"File content length: {content.Length}");
                
                // Using IO.Path
                string directory = IO.Path.GetDirectoryName(path);
                string filename = IO.Path.GetFileName(path);
                
                Console.WriteLine($"Directory: {directory}");
                Console.WriteLine($"Filename: {filename}");
            }
            
            // Using IO.Directory
            var files = IO.Directory.GetFiles(@"C:\temp", "*.txt");
            Console.WriteLine($"Found {files.Length} text files");
        }
        
        public bool ValidateInput(string input)
        {
            // Using Regex alias
            var emailPattern = @"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$";
            var phonePattern = @"^\+?1?\d{9,15}$";
            
            bool isEmail = Regex.IsMatch(input, emailPattern);
            bool isPhone = Regex.IsMatch(input, phonePattern);
            
            return isEmail || isPhone;
        }
        
        public MyDict GetConfiguration()
        {
            // Returning the aliased dictionary type
            return configuration;
        }
        
        public void UpdateConfiguration(string key, object value)
        {
            if (configuration.ContainsKey(key))
            {
                configuration[key] = value;
                Console.WriteLine($"Updated {key} to {value}");
            }
            else
            {
                configuration.Add(key, value);
                Console.WriteLine($"Added {key} with value {value}");
            }
        }
    }
}