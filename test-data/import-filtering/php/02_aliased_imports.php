<?php
// Test Pattern 2: Aliased use Statements
// Tests using 'as' to alias imported names

namespace App\Core;

use App\Models\Product as ProdModel;
use App\Models\Category as Cat;
use App\Exceptions\ValidationException as ValEx;
use App\Exceptions\NotFoundException as NotFound;
use App\Utils\StringHelper as Str;
use App\Utils\ArrayHelper as Arr;
use DateTime as DT;
use DateTimeZone as DTZ;
use Exception as Ex;

// Not using: Cat, NotFound, Str, Arr, DTZ, Ex

class OrderProcessor
{
    /**
     * Process an order with validation
     *
     * @param array $data Order data
     * @return ProdModel The created product
     * @throws ValEx When validation fails
     */
    public function processOrder(array $data): ProdModel
    {
        // Validate required fields
        if (!isset($data['product_id']) || !isset($data['quantity'])) {
            // Using aliased ValidationException
            throw new ValEx("Missing required fields");
        }

        if ($data['quantity'] <= 0) {
            throw new ValEx("Invalid quantity: must be greater than 0");
        }

        // Using aliased Product model
        $product = ProdModel::find($data['product_id']);
        if (!$product) {
            // Could use NotFound here, but using ValEx instead
            throw new ValEx("Product not found");
        }

        // Using aliased DateTime
        $orderDate = new DT('now');

        // Create order
        $order = new ProdModel([
            'product_id' => $product->id,
            'quantity' => $data['quantity'],
            'ordered_at' => $orderDate->format('Y-m-d H:i:s'),
            'total' => $product->price * $data['quantity']
        ]);

        $order->save();

        return $order;
    }

    /**
     * Get order statistics for a date
     * Uses DateTime alias in PHPDoc
     *
     * @param DT $date The date to check
     * @return array Statistics
     */
    public function getStatistics(DT $date): array
    {
        // Using the DateTime parameter
        $startOfDay = clone $date;
        $startOfDay->setTime(0, 0, 0);

        $endOfDay = clone $date;
        $endOfDay->setTime(23, 59, 59);

        return [
            'date' => $date->format('Y-m-d'),
            'start' => $startOfDay->format('c'),
            'end' => $endOfDay->format('c')
        ];
    }
}