<?php

declare(strict_types=1);

/**
 * Calculates the final price after applying a discount.
 *
 * @param float $price The original price.
 * @param int $discountPercent The discount percentage.
 * @return float The price after discount.
 */
function calculate_final_price(float $price, int $discountPercent): float
{
    if ($price <= 0) {
        return 0.0;
    }

    $discountAmount = $price * ($discountPercent / 100);

    return $price - $discountAmount;
}

// A simple, empty class definition to test basic OOP parsing.
class Product
{
}

$bookPrice = 20.0;
$finalPrice = calculate_final_price($bookPrice, 15);

echo "Final price: " . $finalPrice;