"""
Main module for testing semantic analysis.
"""

from utils import Calculator, format_result, PI

def run_calculation():
    """Run a simple calculation demonstration."""
    # Create calculator instance
    calc = Calculator()
    
    # Perform calculations
    result1 = calc.add(5, 3)
    result2 = calc.multiply(result1, 2)
    
    # Format and display results
    formatted1 = format_result(result1)
    formatted2 = format_result(result2)
    
    print(formatted1)
    print(formatted2)
    
    # Show history
    history = calc.get_history()
    print("Calculation history:")
    for entry in history:
        print(f"  {entry}")
    
    # Use constant
    circle_area = PI * 5 * 5
    print(f"Circle area (r=5): {circle_area}")
    
    return result2

def main():
    """Main entry point."""
    final_result = run_calculation()
    print(f"Final result: {final_result}")

if __name__ == "__main__":
    main()