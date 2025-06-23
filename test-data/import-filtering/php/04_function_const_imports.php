<?php
// Test Pattern 4: use function and use const
// Tests explicitly importing functions and constants

namespace App\Utilities;

use function App\Helpers\format_currency;
use function App\Helpers\calculate_tax;
use function App\Helpers\validate_email;
use function App\Helpers\sanitize_input;
use function array_map;
use function array_filter;
use function array_reduce;
use function str_replace;
use function strlen;
use function strtoupper;

use const App\Constants\MAX_ITEMS;
use const App\Constants\MIN_ITEMS;
use const App\Constants\TAX_RATE;
use const App\Constants\CURRENCY_SYMBOL;
use const App\Constants\DEFAULT_TIMEOUT;
use const PHP_VERSION;
use const PHP_EOL;
use const DIRECTORY_SEPARATOR;

// Not using: validate_email, sanitize_input, array_reduce, strlen, strtoupper,
// MIN_ITEMS, DEFAULT_TIMEOUT, PHP_VERSION, DIRECTORY_SEPARATOR

class PriceCalculator
{
    /**
     * Calculate total price for items with tax
     */
    public function calculateTotal(array $items): array
    {
        // Check item count limit (using const)
        if (count($items) > MAX_ITEMS) {
            throw new \InvalidArgumentException(
                "Too many items. Maximum allowed: " . MAX_ITEMS
            );
        }
        
        // Using imported array functions
        $prices = array_map(function ($item) {
            return $item['price'] * $item['quantity'];
        }, $items);
        
        $validPrices = array_filter($prices, function ($price) {
            return $price > 0;
        });
        
        $subtotal = array_sum($validPrices);  // Note: array_sum not imported
        
        // Using imported function for tax calculation
        $tax = calculate_tax($subtotal, TAX_RATE);
        $total = $subtotal + $tax;
        
        // Format the output using imported function and const
        return [
            'subtotal' => format_currency($subtotal, CURRENCY_SYMBOL),
            'tax' => format_currency($tax, CURRENCY_SYMBOL),
            'total' => format_currency($total, CURRENCY_SYMBOL),
            'items_count' => count($validPrices),
            'line_break' => PHP_EOL  // Using imported constant
        ];
    }
    
    /**
     * Generate invoice number
     */
    public function generateInvoiceNumber(string $prefix): string
    {
        // Using str_replace (imported function)
        $cleanPrefix = str_replace(' ', '-', $prefix);
        
        // Generate unique suffix
        $suffix = date('YmdHis') . mt_rand(1000, 9999);
        
        return $cleanPrefix . '-' . $suffix;
    }
    
    /**
     * Display calculation results
     */
    public function displayResults(array $results): void
    {
        echo "=== CALCULATION RESULTS ===" . PHP_EOL;
        echo "Subtotal: " . $results['subtotal'] . PHP_EOL;
        echo "Tax: " . $results['tax'] . PHP_EOL;
        echo "Total: " . $results['total'] . PHP_EOL;
        echo "Items processed: " . $results['items_count'] . PHP_EOL;
    }
}