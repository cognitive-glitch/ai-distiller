// Test Pattern 1: Basic Includes
// Tests standard library and local header includes

#include <iostream>
#include <vector>
#include <string>
#include <map>
#include <algorithm>
#include <memory>
#include <fstream>
#include <sstream>
#include <cmath>
#include <cstdlib>
#include "myheader.h"
#include "utils/stringutils.h"

// Not using: map, memory, fstream, sstream, cstdlib, utils/stringutils.h

using namespace std;

class DataProcessor {
private:
    vector<string> data;
    
public:
    DataProcessor() {
        // Using cout from iostream
        cout << "DataProcessor initialized" << endl;
    }
    
    void addData(const string& item) {
        // Using vector and string
        data.push_back(item);
    }
    
    void processData() {
        // Using algorithm for sorting
        sort(data.begin(), data.end());
        
        // Using iostream for output
        cout << "Sorted data:" << endl;
        for (const auto& item : data) {
            cout << "  " << item << endl;
        }
        
        // Using cmath
        double value = 16.0;
        double result = sqrt(value);
        cout << "Square root of " << value << " is " << result << endl;
    }
    
    void useMyHeader() {
        // Using something from myheader.h
        MyClass obj;
        obj.doSomething();
    }
};

int main() {
    DataProcessor processor;
    
    // Using string
    processor.addData("banana");
    processor.addData("apple");
    processor.addData("cherry");
    
    processor.processData();
    processor.useMyHeader();
    
    return 0;
}