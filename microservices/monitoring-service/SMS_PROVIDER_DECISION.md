# SMS/Email Provider Decision - PENDING

## Status: ⏳ AWAITING DECISION

The monitoring service currently has SMS notification capability implemented with **Twilio integration**, but the final provider selection is pending.

## Options Under Consideration

### Option 1: AWS SNS + AWS SES
**Pros:**
- Integrated with AWS ecosystem
- Cost-effective for high volume
- No separate vendor management
- SES for email, SNS for SMS in one platform

**Cons:**
- Requires AWS account setup
- More complex configuration
- Region-specific

### Option 2: Twilio (SMS) + Twilio SendGrid (Email)
**Pros:**
- Already implemented in code
- Simple API
- Good documentation
- Proven reliability

**Cons:**
- Potentially higher cost at scale
- Separate vendor dependency

## Current Implementation

The monitoring service has a **provider-agnostic architecture** in `internal/services/sms_service.go`:

```go
type SMSService struct {
    db          *gorm.DB
    logger      *zap.Logger
    twilioSID   string  // Can be replaced with AWS credentials
    twilioToken string
    twilioFrom  string
    enabled     bool
}
```

## Migration Path

When the decision is made, update:

1. **Environment Variables:**
   ```bash
   # Current (Twilio)
   TWILIO_ACCOUNT_SID=...
   TWILIO_AUTH_TOKEN=...
   TWILIO_FROM_NUMBER=...

   # Future (AWS SNS)
   AWS_REGION=us-east-1
   AWS_ACCESS_KEY_ID=...
   AWS_SECRET_ACCESS_KEY=...
   SNS_SENDER_ID=...
   ```

2. **Service Implementation:**
   - Update `sendViaTwilio()` → `sendViaSNS()` in `sms_service.go`
   - Replace Twilio HTTP client with AWS SDK
   - Update message format if needed

3. **Email Service:**
   - Create `email_service.go` alongside `sms_service.go`
   - Use AWS SES or Twilio SendGrid
   - Integrate with escalation policies

## Current Status

- ✅ SMS service architecture complete
- ✅ Database schema ready
- ✅ Integration points identified
- ⏳ **Awaiting provider decision**
- ⏸️  SMS service disabled until configured

## Notes

- The SMS service gracefully handles being disabled (no credentials = no-op)
- All code is ready to use once provider is chosen
- No database changes needed regardless of provider choice
- Only `internal/services/sms_service.go` needs updating

---

**Decision Date:** TBD
**Decided By:** User
**Implementation Time:** ~2-4 hours once decided
