<?php
// Test Pattern 3: Grouped use Statements and Complex Patterns
// Tests PHP's use grouping syntax and complex import scenarios

namespace App\Reporting;

use App\Enums\{OrderStatus, PaymentStatus, ShippingStatus};
use App\Events\{ReportGenerated, ReportFailed, ReportScheduled};
use App\Services\{
    PdfGenerator,
    ExcelGenerator,
    CsvGenerator,
    EmailService as Mailer
};
use App\Models\{User, Order, Product};
use Symfony\Component\HttpFoundation\{Request, Response, JsonResponse};

// Not using: PaymentStatus, ShippingStatus, ReportFailed, ReportScheduled, 
// ExcelGenerator, CsvGenerator, User, Product, Request

class SalesReport
{
    private PdfGenerator $pdfGenerator;
    private Mailer $mailer;
    
    public function __construct(PdfGenerator $pdfGenerator, Mailer $mailer)
    {
        $this->pdfGenerator = $pdfGenerator;
        $this->mailer = $mailer;
    }
    
    /**
     * Generate a sales report for a given order status
     */
    public function generate(OrderStatus $status, string $format = 'pdf'): Response
    {
        // Fetch orders with the given status
        $orders = Order::where('status', $status->value)->get();
        
        if ($orders->isEmpty()) {
            // Using JsonResponse from grouped import
            return new JsonResponse([
                'error' => 'No orders found',
                'status' => $status->name
            ], 404);
        }
        
        try {
            // Generate PDF using the injected generator
            $pdfContent = $this->pdfGenerator->generateReport([
                'title' => 'Sales Report',
                'status' => $status->name,
                'orders' => $orders->toArray(),
                'generated_at' => date('Y-m-d H:i:s')
            ]);
            
            // Dispatch event (using from grouped import)
            event(new ReportGenerated([
                'type' => 'sales',
                'status' => $status->value,
                'order_count' => $orders->count()
            ]));
            
            // Return response
            return new Response($pdfContent, 200, [
                'Content-Type' => 'application/pdf',
                'Content-Disposition' => 'attachment; filename="sales-report.pdf"'
            ]);
            
        } catch (\Exception $e) {
            // Could dispatch ReportFailed event here, but choosing not to
            
            return new JsonResponse([
                'error' => 'Report generation failed',
                'message' => $e->getMessage()
            ], 500);
        }
    }
    
    /**
     * Send report via email
     */
    public function emailReport(OrderStatus $status, string $recipient): bool
    {
        $report = $this->generate($status);
        
        if ($report->getStatusCode() === 200) {
            // Using aliased EmailService as Mailer
            return $this->mailer->send([
                'to' => $recipient,
                'subject' => 'Sales Report - ' . $status->name,
                'attachment' => $report->getContent()
            ]);
        }
        
        return false;
    }
}