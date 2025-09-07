package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InvoicingService handles invoice generation and management
type InvoicingService struct {
	db *gorm.DB
}

// NewInvoicingService creates a new invoicing service
func NewInvoicingService() *InvoicingService {
	return &InvoicingService{
		db: database.DB,
	}
}

// InvoiceRequest represents a request to create an invoice
type InvoiceRequest struct {
	TenantID       uint                   `json:"tenant_id"`
	SubscriptionID uint                   `json:"subscription_id"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Description    string                 `json:"description"`
	DueDate        time.Time              `json:"due_date"`
	Items          []InvoiceItem          `json:"items"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// InvoiceItem represents an item on an invoice
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	TaxRate     float64 `json:"tax_rate"`
	TaxAmount   float64 `json:"tax_amount"`
}

// InvoiceResponse represents an invoice response
type InvoiceResponse struct {
	ID             uint                   `json:"id"`
	TenantID       uint                   `json:"tenant_id"`
	SubscriptionID uint                   `json:"subscription_id"`
	InvoiceNumber  string                 `json:"invoice_number"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Status         string                 `json:"status"`
	DueDate        time.Time              `json:"due_date"`
	PaidAt         *time.Time             `json:"paid_at"`
	Items          []InvoiceItem          `json:"items"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreateInvoice creates a new invoice
func (s *InvoicingService) CreateInvoice(ctx context.Context, req *InvoiceRequest) (*InvoiceResponse, error) {
	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber()

	// Calculate total amount from items
	totalAmount := 0.0
	for _, item := range req.Items {
		totalAmount += item.TotalPrice
	}

	// Create invoice
	invoice := &models.BillingInvoice{
		TenantID:       req.TenantID,
		SubscriptionID: req.SubscriptionID,
		Amount:         totalAmount,
		Currency:       req.Currency,
		Status:         "pending",
		DueDate:        req.DueDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(invoice).Error; err != nil {
		logger.Error("Failed to create invoice", zap.Error(err))
		return nil, err
	}

	// Create invoice items
	for _, item := range req.Items {
		invoiceItem := &models.InvoiceItem{
			InvoiceID:   invoice.ID,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			TaxRate:     item.TaxRate,
			TaxAmount:   item.TaxAmount,
			CreatedAt:   time.Now(),
		}

		if err := s.db.Create(invoiceItem).Error; err != nil {
			logger.Error("Failed to create invoice item", zap.Error(err))
			return nil, err
		}
	}

	// Create response
	response := &InvoiceResponse{
		ID:             invoice.ID,
		TenantID:       invoice.TenantID,
		SubscriptionID: invoice.SubscriptionID,
		InvoiceNumber:  invoiceNumber,
		Amount:         invoice.Amount,
		Currency:       invoice.Currency,
		Status:         invoice.Status,
		DueDate:        invoice.DueDate,
		PaidAt:         invoice.PaidAt,
		Items:          req.Items,
		Metadata:       req.Metadata,
		CreatedAt:      invoice.CreatedAt,
		UpdatedAt:      invoice.UpdatedAt,
	}

	logger.Info("Invoice created successfully",
		zap.Uint("invoice_id", invoice.ID),
		zap.Uint("tenant_id", req.TenantID),
		zap.Float64("amount", totalAmount))

	return response, nil
}

// GetInvoice retrieves an invoice by ID
func (s *InvoicingService) GetInvoice(ctx context.Context, invoiceID uint) (*InvoiceResponse, error) {
	var invoice models.BillingInvoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Get invoice items
	var items []models.InvoiceItem
	if err := s.db.Where("invoice_id = ?", invoice.ID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoice items: %w", err)
	}

	// Convert to response format
	response := &InvoiceResponse{
		ID:             invoice.ID,
		TenantID:       invoice.TenantID,
		SubscriptionID: invoice.SubscriptionID,
		InvoiceNumber:  s.generateInvoiceNumberFromID(invoice.ID),
		Amount:         invoice.Amount,
		Currency:       invoice.Currency,
		Status:         invoice.Status,
		DueDate:        invoice.DueDate,
		PaidAt:         invoice.PaidAt,
		CreatedAt:      invoice.CreatedAt,
		UpdatedAt:      invoice.UpdatedAt,
	}

	// Convert items
	for _, item := range items {
		response.Items = append(response.Items, InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			TaxRate:     item.TaxRate,
			TaxAmount:   item.TaxAmount,
		})
	}

	return response, nil
}

// GetInvoicesByTenant retrieves all invoices for a tenant
func (s *InvoicingService) GetInvoicesByTenant(ctx context.Context, tenantID uint, limit, offset int) ([]*InvoiceResponse, error) {
	var invoices []models.BillingInvoice
	query := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	var responses []*InvoiceResponse
	for _, invoice := range invoices {
		response := &InvoiceResponse{
			ID:             invoice.ID,
			TenantID:       invoice.TenantID,
			SubscriptionID: invoice.SubscriptionID,
			InvoiceNumber:  s.generateInvoiceNumberFromID(invoice.ID),
			Amount:         invoice.Amount,
			Currency:       invoice.Currency,
			Status:         invoice.Status,
			DueDate:        invoice.DueDate,
			PaidAt:         invoice.PaidAt,
			CreatedAt:      invoice.CreatedAt,
			UpdatedAt:      invoice.UpdatedAt,
		}

		// Get invoice items
		var items []models.InvoiceItem
		if err := s.db.Where("invoice_id = ?", invoice.ID).Find(&items).Error; err == nil {
			for _, item := range items {
				response.Items = append(response.Items, InvoiceItem{
					Description: item.Description,
					Quantity:    item.Quantity,
					UnitPrice:   item.UnitPrice,
					TotalPrice:  item.TotalPrice,
					TaxRate:     item.TaxRate,
					TaxAmount:   item.TaxAmount,
				})
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// UpdateInvoiceStatus updates the status of an invoice
func (s *InvoicingService) UpdateInvoiceStatus(ctx context.Context, invoiceID uint, status string) error {
	var invoice models.BillingInvoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	invoice.Status = status
	invoice.UpdatedAt = time.Now()

	if status == "paid" {
		now := time.Now()
		invoice.PaidAt = &now
	}

	if err := s.db.Save(&invoice).Error; err != nil {
		logger.Error("Failed to update invoice status", zap.Error(err))
		return err
	}

	logger.Info("Invoice status updated",
		zap.Uint("invoice_id", invoiceID),
		zap.String("status", status))

	return nil
}

// GenerateInvoicePDF generates a PDF for an invoice
func (s *InvoicingService) GenerateInvoicePDF(ctx context.Context, invoiceID uint) ([]byte, error) {
	// Get invoice details
	invoice, err := s.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Get tenant details
	var tenant models.Tenant
	if err := s.db.First(&tenant, invoice.TenantID).Error; err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Generate PDF (this would use a PDF generation library)
	pdfData := s.generatePDFContent(invoice, &tenant)

	return pdfData, nil
}

// SendInvoiceEmail sends an invoice via email
func (s *InvoicingService) SendInvoiceEmail(ctx context.Context, invoiceID uint, recipientEmail string) error {
	// Get invoice details
	_, err := s.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Generate PDF
	_, err = s.GenerateInvoicePDF(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Send email with PDF attachment
	// This would integrate with your email service
	logger.Info("Invoice email sent",
		zap.Uint("invoice_id", invoiceID),
		zap.String("recipient", recipientEmail))

	return nil
}

// GetOverdueInvoices retrieves overdue invoices
func (s *InvoicingService) GetOverdueInvoices(ctx context.Context) ([]*InvoiceResponse, error) {
	var invoices []models.BillingInvoice
	if err := s.db.Where("status = ? AND due_date < ?", "pending", time.Now()).Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	var responses []*InvoiceResponse
	for _, invoice := range invoices {
		response := &InvoiceResponse{
			ID:             invoice.ID,
			TenantID:       invoice.TenantID,
			SubscriptionID: invoice.SubscriptionID,
			InvoiceNumber:  s.generateInvoiceNumberFromID(invoice.ID),
			Amount:         invoice.Amount,
			Currency:       invoice.Currency,
			Status:         invoice.Status,
			DueDate:        invoice.DueDate,
			PaidAt:         invoice.PaidAt,
			CreatedAt:      invoice.CreatedAt,
			UpdatedAt:      invoice.UpdatedAt,
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// generateInvoiceNumber generates a unique invoice number
func (s *InvoicingService) generateInvoiceNumber() string {
	// Generate invoice number with timestamp
	return fmt.Sprintf("INV-%d", time.Now().Unix())
}

// generateInvoiceNumberFromID generates an invoice number from ID
func (s *InvoicingService) generateInvoiceNumberFromID(id uint) string {
	return fmt.Sprintf("INV-%06d", id)
}

// generatePDFContent generates PDF content for an invoice
func (s *InvoicingService) generatePDFContent(invoice *InvoiceResponse, tenant *models.Tenant) []byte {
	// This is a placeholder for PDF generation
	// In a real implementation, you would use a PDF library like gofpdf or unidoc
	content := fmt.Sprintf(`
Invoice: %s
Tenant: %s
Amount: %.2f %s
Status: %s
Due Date: %s
Created: %s
`,
		invoice.InvoiceNumber,
		tenant.Name,
		invoice.Amount,
		invoice.Currency,
		invoice.Status,
		invoice.DueDate.Format("2006-01-02"),
		invoice.CreatedAt.Format("2006-01-02 15:04:05"))

	return []byte(content)
}
