// Test Pattern 4: Global using (C# 10+) and Conditional using
// Tests global using and preprocessor conditional imports

#if NET6_0_OR_GREATER
global using System.Text.Json;
global using System.Net.Http.Json;
#endif

using System;
using System.Collections.Generic;
using System.Threading.Tasks;

#if DEBUG
using System.Diagnostics;
using System.Runtime.CompilerServices;
#else
using System.Runtime.InteropServices;
#endif

using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

#if USE_NEWTONSOFT
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
#else
using JsonSerializer = System.Text.Json.JsonSerializer;
#endif

#if WINDOWS
using Microsoft.Win32;
#elif LINUX
using Mono.Unix;
#endif

// Attributes for testing
using System.Runtime.Serialization;
using System.Xml.Serialization;

// Not using: DataAnnotations.Schema, XmlSerialization, and conditional imports based on defines

namespace ImportFilteringTests
{
    [DataContract]
    public class User
    {
        [DataMember]
        [Required]
        public string Name { get; set; }
        
        [DataMember]
        [EmailAddress]
        public string Email { get; set; }
        
        public int Age { get; set; }
    }
    
    public class GlobalAndConditionalImports
    {
        private readonly HttpClient httpClient = new HttpClient();
        
        public async Task<User> FetchUserAsync(string userId)
        {
            #if NET6_0_OR_GREATER
            // Using global System.Net.Http.Json
            var user = await httpClient.GetFromJsonAsync<User>($"https://api.example.com/users/{userId}");
            return user;
            #else
            var response = await httpClient.GetStringAsync($"https://api.example.com/users/{userId}");
            // Manual deserialization without HttpClient JSON extensions
            return DeserializeUser(response);
            #endif
        }
        
        public string SerializeUser(User user)
        {
            #if USE_NEWTONSOFT
            // Using Newtonsoft.Json
            return JsonConvert.SerializeObject(user, Formatting.Indented);
            #else
            // Using System.Text.Json (via global using or alias)
            var options = new JsonSerializerOptions 
            { 
                WriteIndented = true 
            };
            return JsonSerializer.Serialize(user, options);
            #endif
        }
        
        public User DeserializeUser(string json)
        {
            #if USE_NEWTONSOFT
            return JsonConvert.DeserializeObject<User>(json);
            #else
            return JsonSerializer.Deserialize<User>(json);
            #endif
        }
        
        public void DebugLog(string message, [CallerMemberName] string caller = null)
        {
            #if DEBUG
            // Using System.Diagnostics in debug mode
            Debug.WriteLine($"[{caller}] {message}");
            Trace.WriteLine($"Trace: {message}");
            
            // Using CallerMemberName from System.Runtime.CompilerServices
            Console.WriteLine($"Called from: {caller}");
            #else
            // In release mode, just use console
            Console.WriteLine($"Log: {message}");
            #endif
        }
        
        public void PlatformSpecificOperation()
        {
            #if WINDOWS
            // Using Microsoft.Win32 for Windows
            using (var key = Registry.CurrentUser.OpenSubKey(@"SOFTWARE\MyApp"))
            {
                if (key != null)
                {
                    var value = key.GetValue("Setting");
                    Console.WriteLine($"Registry value: {value}");
                }
            }
            #elif LINUX
            // Using Mono.Unix for Linux
            var info = new UnixFileInfo("/etc/hosts");
            Console.WriteLine($"File permissions: {info.FileAccessPermissions}");
            #else
            Console.WriteLine("Platform-specific operation not available");
            #endif
        }
        
        public bool ValidateUser(User user)
        {
            // Using DataAnnotations
            var context = new ValidationContext(user);
            var results = new List<ValidationResult>();
            
            bool isValid = Validator.TryValidateObject(user, context, results, true);
            
            if (!isValid)
            {
                foreach (var result in results)
                {
                    Console.WriteLine($"Validation error: {result.ErrorMessage}");
                }
            }
            
            return isValid;
        }
        
        #if DEBUG
        [Conditional("DEBUG")]
        public void DebugOnlyMethod()
        {
            Debug.Assert(httpClient != null, "HttpClient should not be null");
            Debug.WriteLine("This method only runs in debug builds");
        }
        #endif
    }
}