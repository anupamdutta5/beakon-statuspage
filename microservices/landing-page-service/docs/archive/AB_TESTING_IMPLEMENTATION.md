# A/B Testing Implementation Guide

## Overview

Complete A/B testing functionality has been implemented for the Landing Page Service, allowing data-driven optimization of content, layouts, and user experience elements.

## Architecture

### Components

1. **Backend Service** (`internal/services/ab_test_service.go`)
   - Test management (CRUD operations)
   - Variant assignment using consistent hashing
   - Conversion tracking with deduplication
   - Statistical significance calculation
   - Results analysis and reporting

2. **API Handlers** (`internal/handlers/ab_test_handler.go`)
   - Admin endpoints for test management
   - Public endpoints for variant assignment
   - Conversion tracking endpoints
   - Results and reporting endpoints

3. **Frontend Client** (`web/static/js/ab-test.js`)
   - Automatic variant application
   - Conversion event tracking
   - DOM-based test configuration
   - Debug mode for development

4. **Database Models** (`internal/models/seo.go`)
   - `LandingPageABTest` - Test configuration
   - `ABTestAssignment` - User-variant assignments

## Features

### Core Functionality

✅ **Test Management**
- Create, update, delete A/B tests
- Start, stop, pause tests
- Configure traffic split (0-100%)
- Set test duration and goals

✅ **Variant Assignment**
- Consistent hashing for deterministic assignment
- Cookie-based session tracking
- Automatic view tracking
- Support for control variants

✅ **Conversion Tracking**
- Deduplication to prevent double-counting
- Multiple conversion events per test
- Automatic and manual tracking options
- Session-based conversion attribution

✅ **Statistical Analysis**
- Conversion rate calculation
- Statistical significance testing (z-test)
- Confidence level calculation
- Minimum sample size enforcement
- Winner determination

✅ **Reporting**
- Real-time results dashboard
- Hourly conversion data
- Actionable insights generation
- Export capabilities

## API Endpoints

### Admin Endpoints (Authenticated)

```
GET    /api/v1/admin/ab-tests              - List all tests
POST   /api/v1/admin/ab-tests              - Create new test
GET    /api/v1/admin/ab-tests/:id          - Get test details
PUT    /api/v1/admin/ab-tests/:id          - Update test
DELETE /api/v1/admin/ab-tests/:id          - Delete test
POST   /api/v1/admin/ab-tests/:id/start    - Start test
POST   /api/v1/admin/ab-tests/:id/stop     - Stop test
GET    /api/v1/admin/ab-tests/:id/results  - Get results
GET    /api/v1/admin/ab-tests/:id/report   - Generate report
```

### Public Endpoints

```
GET    /api/v1/public/ab-test/:id/variant     - Get variant assignment
POST   /api/v1/public/ab-test/:id/conversion  - Track conversion
```

## Usage Examples

### Creating an A/B Test

```bash
curl -X POST http://localhost:8100/api/v1/admin/ab-tests \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hero Headline Test",
    "description": "Testing two different hero headlines",
    "element_type": "hero_title",
    "element_id": 1,
    "traffic_split": 50,
    "variant_a_config": "{\"text\": \"Keep your users informed\"}",
    "variant_b_config": "{\"text\": \"The easiest way to communicate\"}"
  }'
```

### Frontend Integration

#### HTML Setup
```html
<!-- Automatic A/B test on headline -->
<h1 data-ab-test="1"
    data-ab-element="text"
    data-ab-conversion="view">
    Original Headline
</h1>

<!-- A/B test on CTA button -->
<button data-ab-test="2"
        data-ab-element="button"
        data-ab-track="cta">
    Get Started
</button>
```

#### JavaScript Integration
```javascript
// Initialize A/B testing
window.ABTestConfig = {
    apiBase: '/api/v1',
    debug: true
};

// Manual test application
abTest.applyManualTest('3', document.getElementById('pricing'), {
    elementType: 'custom',
    conversionEvent: 'purchase'
});

// Manual conversion tracking
abTest.trackConversion('3');

// Listen for variant application
element.addEventListener('ab-variant-apply', (e) => {
    console.log('Variant:', e.detail.variant);
    console.log('Config:', e.detail.config);
});
```

## Test Types Supported

### 1. Text Variations
Test different headlines, descriptions, or copy:
```json
{
  "element_type": "hero_title",
  "variant_a_config": {"text": "Original text"},
  "variant_b_config": {"text": "Alternative text"}
}
```

### 2. Button Variations
Test button text, color, or style:
```json
{
  "element_type": "cta_button",
  "variant_a_config": {
    "text": "Get Started",
    "class": "btn-primary"
  },
  "variant_b_config": {
    "text": "Start Free Trial",
    "class": "btn-primary btn-pulse"
  }
}
```

### 3. Layout Variations
Test different layouts or component orders:
```json
{
  "element_type": "pricing_layout",
  "variant_a_config": {"layout": "grid"},
  "variant_b_config": {"layout": "cards"}
}
```

### 4. Image Variations
Test different images or graphics:
```json
{
  "element_type": "hero_image",
  "variant_a_config": {"src": "/images/hero-a.jpg"},
  "variant_b_config": {"src": "/images/hero-b.jpg"}
}
```

## Statistical Significance

The system uses a two-tailed z-test to determine statistical significance:

```go
// Minimum sample size: 100 per variant
// Confidence threshold: 95% (configurable)
//
// Formula:
// z = (p₂ - p₁) / SE
// where SE = √(p_pooled * (1-p_pooled) * (1/n₁ + 1/n₂))
```

### Interpretation

- **Confidence Level > 95%**: Statistically significant difference
- **Confidence Level < 95%**: Continue testing for more data
- **Minimum Sample**: 100 visitors per variant before declaring winner

## Best Practices

### 1. Test Design
- Test one variable at a time
- Ensure variants are meaningfully different
- Define clear success metrics upfront
- Run tests for complete weeks (avoid day-of-week bias)

### 2. Sample Size
- Wait for minimum sample size (100+ per variant)
- Don't stop tests too early (false positives)
- Consider your baseline conversion rate
- Account for seasonal variations

### 3. Implementation
- Use consistent test IDs across environments
- Implement proper error handling
- Track both macro and micro conversions
- Document test hypotheses and results

### 4. Analysis
- Look beyond conversion rates (consider revenue, engagement)
- Segment results by user type if relevant
- Consider practical significance, not just statistical
- Document learnings for future tests

## Configuration

### Service Configuration
```go
config := &ABTestConfig{
    MinSampleSize:       100,    // Minimum visitors per variant
    ConfidenceThreshold: 0.95,   // 95% confidence level
    DefaultTrafficSplit: 50,     // 50/50 split by default
    CookieExpiry:       30,      // 30 days session persistence
}
```

### Database Migration
```sql
-- Run migration to create tables
CREATE TABLE landing_page_ab_tests (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    element_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    traffic_split INT DEFAULT 50,
    variant_a_config TEXT,
    variant_b_config TEXT,
    -- ... (see models for full schema)
);

CREATE TABLE ab_test_assignments (
    id SERIAL PRIMARY KEY,
    test_id INT NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    variant VARCHAR(10) NOT NULL,
    converted BOOLEAN DEFAULT false,
    -- ...
);
```

## Monitoring & Debugging

### Debug Mode
Enable debug mode to see variant assignments:
```javascript
window.ABTestConfig = {
    debug: true
};
```

### Debug Panel
The example template includes a debug panel showing:
- Active tests and their variants
- Assigned variants per element
- Conversion tracking events

### Logging
All A/B test operations are logged:
```
INFO: Creating A/B test name=Hero Headline Test
INFO: Starting A/B test test_id=1
INFO: Variant assigned test_id=1 session_id=sess_123 variant=B
INFO: Conversion tracked test_id=1 variant=B
```

## Performance Considerations

### Caching
- Variant assignments are cached per session
- Test configurations cached in memory
- Results calculated on-demand (consider caching)

### Database Queries
- Indexed on test_id and session_id for fast lookups
- Batch updates for view/conversion counts
- Soft deletes to preserve historical data

### Frontend
- Lightweight client library (~5KB minified)
- Async variant fetching
- Debounced conversion tracking
- Local storage for offline support (future)

## Security

### Authentication
- Admin endpoints require JWT authentication
- Public endpoints use session-based tracking
- CORS configured for API access

### Data Privacy
- No PII stored in test data
- Session IDs are anonymous
- Cookie consent compliance ready
- GDPR-friendly data retention

## Troubleshooting

### Common Issues

1. **Variant not applying**
   - Check test status (must be "running")
   - Verify element selector/ID
   - Check browser console for errors
   - Ensure cookies are enabled

2. **Conversions not tracking**
   - Verify session cookie exists
   - Check conversion event binding
   - Ensure no duplicate prevention
   - Verify API endpoint accessibility

3. **Statistical significance not reached**
   - Increase test duration
   - Check traffic volume
   - Verify conversion events fire correctly
   - Consider increasing variant difference

## Future Enhancements

### Planned Features
- [ ] Multi-variate testing (test multiple variables)
- [ ] Audience segmentation (test by user attributes)
- [ ] Revenue tracking (not just conversions)
- [ ] Bayesian statistics option
- [ ] Visual editor for creating tests
- [ ] Automated test recommendations
- [ ] Integration with analytics platforms
- [ ] Real-time results streaming
- [ ] Test scheduling and automation
- [ ] Machine learning for winner prediction

### Integration Opportunities
- Google Analytics events
- Segment.io tracking
- Amplitude analytics
- Mixpanel events
- Custom webhooks
- Slack notifications
- Email reports

## Conclusion

The A/B testing implementation provides a robust, statistically-sound platform for optimizing the landing page experience. With automatic variant assignment, conversion tracking, and comprehensive reporting, teams can make data-driven decisions to improve conversion rates and user experience.

---

**Documentation Version**: 1.0
**Last Updated**: October 2024
**Maintained By**: Platform Team