// Package services provides business logic for the Landing Page Service.
package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ABTestService handles A/B testing functionality
type ABTestService struct {
	db     *gorm.DB
	logger *zap.Logger
	config *ABTestConfig
}

// ABTestConfig contains A/B testing configuration
type ABTestConfig struct {
	MinSampleSize        int     // Minimum visitors per variant before declaring winner
	ConfidenceThreshold  float64 // Statistical confidence level (e.g., 0.95 for 95%)
	DefaultTrafficSplit  int     // Default percentage to variant B (0-100)
	CookieExpiry        int     // Days to remember user's assignment
}

// NewABTestService creates a new A/B test service
func NewABTestService(db *gorm.DB, logger *zap.Logger, config *ABTestConfig) *ABTestService {
	if config == nil {
		config = &ABTestConfig{
			MinSampleSize:       100,
			ConfidenceThreshold: 0.95,
			DefaultTrafficSplit: 50,
			CookieExpiry:       30,
		}
	}

	return &ABTestService{
		db:     db,
		logger: logger,
		config: config,
	}
}

// ABTestVariant represents a test variant with its configuration
type ABTestVariant struct {
	Name   string                 `json:"name"`   // 'A' or 'B'
	Config map[string]interface{} `json:"config"` // Variant configuration
}

// ABTestResult represents the result of an A/B test
type ABTestResult struct {
	TestID              uint    `json:"test_id"`
	TestName            string  `json:"test_name"`
	Status              string  `json:"status"`
	VariantA            Variant `json:"variant_a"`
	VariantB            Variant `json:"variant_b"`
	Winner              string  `json:"winner,omitempty"`
	ConfidenceLevel     float64 `json:"confidence_level"`
	StatisticallySignificant bool `json:"statistically_significant"`
	DaysRunning         int     `json:"days_running"`
	RecommendedAction   string  `json:"recommended_action"`
}

// Variant represents a variant's performance
type Variant struct {
	Name            string  `json:"name"`
	Views           int     `json:"views"`
	Conversions     int     `json:"conversions"`
	ConversionRate  float64 `json:"conversion_rate"`
	Improvement     float64 `json:"improvement,omitempty"` // Percentage improvement over control
}

// CreateABTest creates a new A/B test
func (s *ABTestService) CreateABTest(ctx context.Context, test *models.LandingPageABTest) error {
	s.logger.Info("Creating A/B test", zap.String("name", test.Name))

	// Validate test configuration
	if err := s.validateTestConfig(test); err != nil {
		s.logger.Error("Invalid test configuration", zap.Error(err))
		return err
	}

	// Set default values
	if test.Status == "" {
		test.Status = "draft"
	}
	if test.TrafficSplit == 0 {
		test.TrafficSplit = s.config.DefaultTrafficSplit
	}

	// Save to database
	if err := s.db.WithContext(ctx).Create(test).Error; err != nil {
		s.logger.Error("Failed to create A/B test", zap.Error(err))
		return fmt.Errorf("failed to create A/B test: %w", err)
	}

	s.logger.Info("A/B test created successfully",
		zap.Uint("test_id", test.ID),
		zap.String("status", test.Status))

	return nil
}

// StartABTest starts an A/B test
func (s *ABTestService) StartABTest(ctx context.Context, testID uint) error {
	s.logger.Info("Starting A/B test", zap.Uint("test_id", testID))

	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		s.logger.Error("Test not found", zap.Error(err))
		return fmt.Errorf("test not found: %w", err)
	}

	// Check if test can be started
	if test.Status != "draft" && test.Status != "paused" {
		return fmt.Errorf("test cannot be started from status: %s", test.Status)
	}

	// Update status and start date
	now := time.Now()
	updates := map[string]interface{}{
		"status":     "running",
		"start_date": now,
	}

	if err := s.db.WithContext(ctx).Model(&test).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to start A/B test", zap.Error(err))
		return fmt.Errorf("failed to start A/B test: %w", err)
	}

	s.logger.Info("A/B test started successfully", zap.Uint("test_id", testID))
	return nil
}

// StopABTest stops an A/B test
func (s *ABTestService) StopABTest(ctx context.Context, testID uint) error {
	s.logger.Info("Stopping A/B test", zap.Uint("test_id", testID))

	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		s.logger.Error("Test not found", zap.Error(err))
		return fmt.Errorf("test not found: %w", err)
	}

	if test.Status != "running" {
		return fmt.Errorf("test is not running")
	}

	// Calculate final results
	result, err := s.CalculateResults(ctx, testID)
	if err != nil {
		s.logger.Error("Failed to calculate results", zap.Error(err))
	}

	// Update status and end date
	now := time.Now()
	updates := map[string]interface{}{
		"status":   "completed",
		"end_date": now,
		"winner":   result.Winner,
		"confidence_level": result.ConfidenceLevel,
	}

	if err := s.db.WithContext(ctx).Model(&test).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to stop A/B test", zap.Error(err))
		return fmt.Errorf("failed to stop A/B test: %w", err)
	}

	s.logger.Info("A/B test stopped successfully",
		zap.Uint("test_id", testID),
		zap.String("winner", result.Winner))

	return nil
}

// GetVariant returns the variant assignment for a user/session
func (s *ABTestService) GetVariant(ctx context.Context, testID uint, sessionID string) (*ABTestVariant, error) {
	// Check if test is running
	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		return nil, fmt.Errorf("test not found: %w", err)
	}

	if test.Status != "running" {
		// Return control variant if test is not running
		return s.parseVariantConfig("A", test.VariantAConfig)
	}

	// Check for existing assignment
	var assignment models.ABTestAssignment
	err := s.db.WithContext(ctx).Where(
		"test_id = ? AND session_id = ?",
		testID,
		sessionID,
	).First(&assignment).Error

	if err == nil {
		// User already assigned to a variant
		if assignment.Variant == "B" {
			return s.parseVariantConfig("B", test.VariantBConfig)
		}
		return s.parseVariantConfig("A", test.VariantAConfig)
	}

	// Assign new user to variant
	variant := s.assignVariant(sessionID, test.TrafficSplit)

	// Save assignment
	assignment = models.ABTestAssignment{
		TestID:    testID,
		SessionID: sessionID,
		Variant:   variant,
		CreatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&assignment).Error; err != nil {
		s.logger.Error("Failed to save assignment", zap.Error(err))
	}

	// Track view
	s.TrackView(ctx, testID, variant)

	// Return variant configuration
	if variant == "B" {
		return s.parseVariantConfig("B", test.VariantBConfig)
	}
	return s.parseVariantConfig("A", test.VariantAConfig)
}

// TrackView tracks a view for a variant
func (s *ABTestService) TrackView(ctx context.Context, testID uint, variant string) error {
	field := "variant_a_views"
	if variant == "B" {
		field = "variant_b_views"
	}

	return s.db.WithContext(ctx).Model(&models.LandingPageABTest{}).
		Where("id = ?", testID).
		Update(field, gorm.Expr(field+" + ?", 1)).Error
}

// TrackConversion tracks a conversion for a variant
func (s *ABTestService) TrackConversion(ctx context.Context, testID uint, sessionID string) error {
	// Get user's variant assignment
	var assignment models.ABTestAssignment
	err := s.db.WithContext(ctx).Where(
		"test_id = ? AND session_id = ?",
		testID,
		sessionID,
	).First(&assignment).Error

	if err != nil {
		s.logger.Warn("No assignment found for conversion",
			zap.Uint("test_id", testID),
			zap.String("session_id", sessionID))
		return fmt.Errorf("no assignment found: %w", err)
	}

	// Check if already converted
	if assignment.Converted {
		return nil // Already converted, don't count twice
	}

	// Update conversion count
	field := "variant_a_conversions"
	if assignment.Variant == "B" {
		field = "variant_b_conversions"
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()

	// Update test conversion count
	if err := tx.Model(&models.LandingPageABTest{}).
		Where("id = ?", testID).
		Update(field, gorm.Expr(field+" + ?", 1)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update conversion count: %w", err)
	}

	// Mark assignment as converted
	if err := tx.Model(&assignment).Update("converted", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark as converted: %w", err)
	}

	tx.Commit()

	s.logger.Info("Conversion tracked",
		zap.Uint("test_id", testID),
		zap.String("variant", assignment.Variant))

	return nil
}

// CalculateResults calculates the results of an A/B test
func (s *ABTestService) CalculateResults(ctx context.Context, testID uint) (*ABTestResult, error) {
	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		return nil, fmt.Errorf("test not found: %w", err)
	}

	// Calculate conversion rates
	variantA := Variant{
		Name:        "A",
		Views:       test.VariantAViews,
		Conversions: test.VariantAConversions,
	}
	if variantA.Views > 0 {
		variantA.ConversionRate = float64(variantA.Conversions) / float64(variantA.Views) * 100
	}

	variantB := Variant{
		Name:        "B",
		Views:       test.VariantBViews,
		Conversions: test.VariantBConversions,
	}
	if variantB.Views > 0 {
		variantB.ConversionRate = float64(variantB.Conversions) / float64(variantB.Views) * 100
	}

	// Calculate improvement
	if variantA.ConversionRate > 0 {
		variantB.Improvement = ((variantB.ConversionRate - variantA.ConversionRate) / variantA.ConversionRate) * 100
	}

	// Calculate statistical significance
	confidence, significant := s.calculateStatisticalSignificance(
		variantA.Views, variantA.Conversions,
		variantB.Views, variantB.Conversions,
	)

	// Determine winner
	winner := ""
	recommendedAction := "Continue testing for more data"

	if significant {
		if variantB.ConversionRate > variantA.ConversionRate {
			winner = "B"
			recommendedAction = "Implement variant B"
		} else if variantA.ConversionRate > variantB.ConversionRate {
			winner = "A"
			recommendedAction = "Keep variant A (control)"
		}
	} else if variantA.Views+variantB.Views > s.config.MinSampleSize*2 {
		recommendedAction = "No significant difference detected. Consider stopping the test."
	}

	// Calculate days running
	daysRunning := 0
	if test.StartDate != nil {
		daysRunning = int(time.Since(*test.StartDate).Hours() / 24)
	}

	return &ABTestResult{
		TestID:                   test.ID,
		TestName:                 test.Name,
		Status:                   test.Status,
		VariantA:                 variantA,
		VariantB:                 variantB,
		Winner:                   winner,
		ConfidenceLevel:          confidence * 100,
		StatisticallySignificant: significant,
		DaysRunning:              daysRunning,
		RecommendedAction:        recommendedAction,
	}, nil
}

// GetAllTests returns all A/B tests with optional filtering
func (s *ABTestService) GetAllTests(ctx context.Context, status string) ([]*models.LandingPageABTest, error) {
	var tests []*models.LandingPageABTest

	query := s.db.WithContext(ctx)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&tests).Error; err != nil {
		s.logger.Error("Failed to get A/B tests", zap.Error(err))
		return nil, fmt.Errorf("failed to get A/B tests: %w", err)
	}

	return tests, nil
}

// GetTestByID returns a single A/B test by ID
func (s *ABTestService) GetTestByID(ctx context.Context, testID uint) (*models.LandingPageABTest, error) {
	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("test not found")
		}
		return nil, fmt.Errorf("failed to get test: %w", err)
	}
	return &test, nil
}

// UpdateABTest updates an A/B test configuration
func (s *ABTestService) UpdateABTest(ctx context.Context, testID uint, updates map[string]interface{}) error {
	// Don't allow updating running tests
	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		return fmt.Errorf("test not found: %w", err)
	}

	if test.Status == "running" {
		return fmt.Errorf("cannot update running test")
	}

	if err := s.db.WithContext(ctx).Model(&test).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update A/B test", zap.Error(err))
		return fmt.Errorf("failed to update A/B test: %w", err)
	}

	return nil
}

// DeleteABTest deletes an A/B test
func (s *ABTestService) DeleteABTest(ctx context.Context, testID uint) error {
	// Don't allow deleting running tests
	var test models.LandingPageABTest
	if err := s.db.WithContext(ctx).First(&test, testID).Error; err != nil {
		return fmt.Errorf("test not found: %w", err)
	}

	if test.Status == "running" {
		return fmt.Errorf("cannot delete running test")
	}

	// Soft delete
	if err := s.db.WithContext(ctx).Delete(&test).Error; err != nil {
		s.logger.Error("Failed to delete A/B test", zap.Error(err))
		return fmt.Errorf("failed to delete A/B test: %w", err)
	}

	// Clean up assignments
	if err := s.db.WithContext(ctx).Where("test_id = ?", testID).Delete(&models.ABTestAssignment{}).Error; err != nil {
		s.logger.Warn("Failed to clean up assignments", zap.Error(err))
	}

	return nil
}

// Helper methods

// assignVariant determines which variant to assign based on traffic split
func (s *ABTestService) assignVariant(sessionID string, trafficSplit int) string {
	// Use consistent hashing for deterministic assignment
	hash := md5.Sum([]byte(sessionID))
	hashInt := int(hash[0])<<8 | int(hash[1])
	percentage := (hashInt * 100) / 65536

	if percentage < trafficSplit {
		return "B"
	}
	return "A"
}

// parseVariantConfig parses variant configuration from JSON string
func (s *ABTestService) parseVariantConfig(name, configJSON string) (*ABTestVariant, error) {
	variant := &ABTestVariant{
		Name:   name,
		Config: make(map[string]interface{}),
	}

	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &variant.Config); err != nil {
			s.logger.Warn("Failed to parse variant config",
				zap.String("variant", name),
				zap.Error(err))
		}
	}

	return variant, nil
}

// validateTestConfig validates A/B test configuration
func (s *ABTestService) validateTestConfig(test *models.LandingPageABTest) error {
	if test.Name == "" {
		return fmt.Errorf("test name is required")
	}

	if test.ElementType == "" {
		return fmt.Errorf("element type is required")
	}

	if test.TrafficSplit < 0 || test.TrafficSplit > 100 {
		return fmt.Errorf("traffic split must be between 0 and 100")
	}

	// Validate variant configurations are valid JSON
	if test.VariantAConfig != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(test.VariantAConfig), &config); err != nil {
			return fmt.Errorf("invalid variant A config: %w", err)
		}
	}

	if test.VariantBConfig != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(test.VariantBConfig), &config); err != nil {
			return fmt.Errorf("invalid variant B config: %w", err)
		}
	}

	return nil
}

// calculateStatisticalSignificance calculates whether the difference is statistically significant
func (s *ABTestService) calculateStatisticalSignificance(
	viewsA, conversionsA, viewsB, conversionsB int,
) (confidence float64, significant bool) {
	// Check minimum sample size
	if viewsA < s.config.MinSampleSize || viewsB < s.config.MinSampleSize {
		return 0, false
	}

	// Calculate conversion rates
	pA := float64(conversionsA) / float64(viewsA)
	pB := float64(conversionsB) / float64(viewsB)

	// Calculate pooled probability
	pPooled := float64(conversionsA+conversionsB) / float64(viewsA+viewsB)

	// Calculate standard error
	se := math.Sqrt(pPooled * (1 - pPooled) * (1.0/float64(viewsA) + 1.0/float64(viewsB)))

	// Avoid division by zero
	if se == 0 {
		return 0, false
	}

	// Calculate z-score
	z := (pB - pA) / se

	// Calculate confidence level (two-tailed test)
	confidence = 2 * (1 - normalCDF(math.Abs(z)))

	// Check if significant at configured threshold
	significant = confidence >= s.config.ConfidenceThreshold

	return confidence, significant
}

// normalCDF approximates the cumulative distribution function of the standard normal distribution
func normalCDF(x float64) float64 {
	// Using approximation from Zelen & Severo (1964)
	b0 := 0.2316419
	b1 := 0.319381530
	b2 := -0.356563782
	b3 := 1.781477937
	b4 := -1.821255978
	b5 := 1.330274429

	t := 1.0 / (1.0 + b0*x)
	y := 1.0 - (math.Exp(-x*x/2.0) / math.Sqrt(2*math.Pi)) *
		(b1*t + b2*t*t + b3*t*t*t + b4*t*t*t*t + b5*t*t*t*t*t)

	return y
}

// GenerateTestReport generates a detailed report for an A/B test
func (s *ABTestService) GenerateTestReport(ctx context.Context, testID uint) (map[string]interface{}, error) {
	test, err := s.GetTestByID(ctx, testID)
	if err != nil {
		return nil, err
	}

	result, err := s.CalculateResults(ctx, testID)
	if err != nil {
		return nil, err
	}

	// Get hourly conversion data for charts
	hourlyData := s.getHourlyConversionData(ctx, testID)

	report := map[string]interface{}{
		"test":        test,
		"result":      result,
		"hourly_data": hourlyData,
		"insights":    s.generateInsights(result),
		"generated_at": time.Now(),
	}

	return report, nil
}

// getHourlyConversionData gets conversion data by hour for charting
func (s *ABTestService) getHourlyConversionData(ctx context.Context, testID uint) []map[string]interface{} {
	// This would query more detailed tracking data if available
	// For now, returning mock data structure
	return []map[string]interface{}{
		{
			"hour": "2024-01-01T00:00:00Z",
			"variant_a": map[string]int{
				"views":       10,
				"conversions": 2,
			},
			"variant_b": map[string]int{
				"views":       12,
				"conversions": 3,
			},
		},
	}
}

// generateInsights generates actionable insights from test results
func (s *ABTestService) generateInsights(result *ABTestResult) []string {
	insights := []string{}

	// Statistical significance insight
	if result.StatisticallySignificant {
		insights = append(insights, fmt.Sprintf(
			"The test has reached statistical significance with %.1f%% confidence.",
			result.ConfidenceLevel,
		))
	} else {
		needed := s.config.MinSampleSize - result.VariantA.Views
		if needed > 0 {
			insights = append(insights, fmt.Sprintf(
				"Need approximately %d more visitors to reach minimum sample size.",
				needed,
			))
		}
	}

	// Performance insight
	if result.VariantB.Improvement > 0 {
		insights = append(insights, fmt.Sprintf(
			"Variant B is performing %.1f%% better than Variant A.",
			result.VariantB.Improvement,
		))
	} else if result.VariantB.Improvement < 0 {
		insights = append(insights, fmt.Sprintf(
			"Variant A is performing %.1f%% better than Variant B.",
			-result.VariantB.Improvement,
		))
	} else {
		insights = append(insights, "Both variants are performing equally.")
	}

	// Duration insight
	if result.DaysRunning > 14 && !result.StatisticallySignificant {
		insights = append(insights,
			"Test has been running for over 2 weeks without reaching significance. Consider stopping or modifying the test.")
	}

	// Sample size insight
	totalViews := result.VariantA.Views + result.VariantB.Views
	if totalViews < 1000 {
		insights = append(insights,
			"Low traffic volume detected. Consider increasing traffic or test duration.")
	}

	return insights
}