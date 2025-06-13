using System;

namespace TestNamespace
{
    public class TestClass
    {
        private int _field;
        
        public string Name { get; set; }
        
        public TestClass(string name)
        {
            Name = name;
        }
        
        public void PrintName()
        {
            Console.WriteLine(Name);
        }
    }
    
    public interface ITest
    {
        void DoSomething();
    }
    
    public enum TestEnum
    {
        Value1,
        Value2
    }
}