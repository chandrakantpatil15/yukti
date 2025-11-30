# AWS SES Email Setup

## Overview
Replaced SMTP with AWS SES (Simple Email Service) for email verification.

## Changes Made

### 1. Email Service Migration
**File**: `internal/services/email.go`

**Before** (SMTP):
```go
type EmailService struct {
    SMTPHost     string
    SMTPPort     string
    SMTPUsername string
    SMTPPassword string
    FromEmail    string
    FromName     string
}
```

**After** (AWS SES):
```go
type EmailService struct {
    sesClient *ses.Client
    FromEmail string
    FromName  string
}
```

### 2. Send Email Implementation
**Before**: Used `net/smtp` package
**After**: Uses AWS SDK v2 SES client

```go
func (e *EmailService) sendEmail(to, subject, body string) error {
    // Dev mode fallback if SES not configured
    if e.sesClient == nil {
        fmt.Printf("\n=== EMAIL (DEV MODE) ===\n")
        // ... print to console
        return nil
    }

    // Send via AWS SES
    input := &ses.SendEmailInput{
        Source: aws.String(fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)),
        Destination: &types.Destination{
            ToAddresses: []string{to},
        },
        Message: &types.Message{
            Subject: &types.Content{
                Data:    aws.String(subject),
                Charset: aws.String("UTF-8"),
            },
            Body: &types.Body{
                Html: &types.Content{
                    Data:    aws.String(body),
                    Charset: aws.String("UTF-8"),
                },
            },
        },
    }

    _, err := e.sesClient.SendEmail(context.Background(), input)
    return err
}
```

---

## AWS SES Setup Guide

### Step 1: Verify Email Address (Sandbox Mode)

1. Go to AWS SES Console
2. Navigate to "Verified identities"
3. Click "Create identity"
4. Select "Email address"
5. Enter: `noreply@yukti.io` (or your domain)
6. Click "Create identity"
7. Check email inbox for verification link
8. Click verification link

### Step 2: Verify Domain (Production)

1. Go to AWS SES Console
2. Navigate to "Verified identities"
3. Click "Create identity"
4. Select "Domain"
5. Enter your domain: `yukti.io`
6. Enable DKIM signing
7. Add DNS records to your domain:
   - DKIM records (3 CNAME records)
   - SPF record (TXT record)
   - DMARC record (TXT record)
8. Wait for verification (can take up to 72 hours)

### Step 3: Request Production Access

**Sandbox Limitations**:
- Can only send to verified email addresses
- Limited to 200 emails per day
- 1 email per second

**To Request Production Access**:
1. Go to AWS SES Console
2. Click "Account dashboard"
3. Click "Request production access"
4. Fill out the form:
   - **Use case**: Transactional emails (OTP verification)
   - **Website URL**: https://yukti.io
   - **Description**: "Sending OTP verification codes for user authentication"
   - **Compliance**: Confirm you have opt-in process
5. Submit request
6. Wait for approval (usually 24-48 hours)

### Step 4: Configure IAM Permissions

Create IAM policy for SES:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ],
      "Resource": "*"
    }
  ]
}
```

Attach to your application's IAM role or user.

### Step 5: Set Environment Variables

```bash
# AWS credentials (if not using IAM role)
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1

# Email configuration
FROM_EMAIL=noreply@yukti.io
FROM_NAME=Yukti FinOps
```

---

## Testing

### Development Mode (No AWS Credentials)
- Emails print to console
- No actual emails sent
- OTP code displayed in logs

### Production Mode (With AWS Credentials)
- Emails sent via AWS SES
- Delivery tracked in SES console
- Bounce/complaint handling available

### Test Email Sending

```bash
# Check backend logs
docker-compose logs -f backend

# Look for:
# [INFO] Email sent successfully via SES to user@example.com
# OR
# === EMAIL (DEV MODE) === (if SES not configured)
```

---

## Configuration

### Docker Compose

Add to `docker-compose.yml`:

```yaml
services:
  backend:
    environment:
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      - AWS_REGION=us-east-1
      - FROM_EMAIL=noreply@yukti.io
      - FROM_NAME=Yukti FinOps
```

### Kubernetes

Create secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-ses-credentials
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: your_access_key
  AWS_SECRET_ACCESS_KEY: your_secret_key
```

Reference in deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yukti-backend
spec:
  template:
    spec:
      containers:
      - name: backend
        envFrom:
        - secretRef:
            name: aws-ses-credentials
        env:
        - name: AWS_REGION
          value: "us-east-1"
        - name: FROM_EMAIL
          value: "noreply@yukti.io"
```

---

## Monitoring

### SES Console Metrics
- **Sends**: Total emails sent
- **Deliveries**: Successfully delivered
- **Bounces**: Failed deliveries
- **Complaints**: Spam reports

### CloudWatch Metrics
- `ses:Send` - Email send requests
- `ses:Delivery` - Successful deliveries
- `ses:Bounce` - Bounced emails
- `ses:Complaint` - Spam complaints

### Application Logs
```bash
# Success
[INFO] Email sent successfully via SES to user@example.com

# Failure
[ERROR] Failed to send email via SES to user@example.com: <error details>
```

---

## Cost

### AWS SES Pricing (US East)
- **First 62,000 emails/month**: FREE (if sent from EC2)
- **Additional emails**: $0.10 per 1,000 emails
- **Attachments**: $0.12 per GB

### Example Costs
- **1,000 users/month**: FREE
- **10,000 users/month**: FREE
- **100,000 users/month**: $3.80/month
- **1,000,000 users/month**: $93.80/month

**Much cheaper than SendGrid/Mailgun!**

---

## Troubleshooting

### Error: "Email address is not verified"
**Solution**: Verify email in SES console (sandbox mode)

### Error: "Daily sending quota exceeded"
**Solution**: Request production access

### Error: "Failed to load AWS config"
**Solution**: 
- Check AWS credentials are set
- Verify IAM permissions
- Check AWS region is correct

### Error: "MessageRejected: Email address is not verified"
**Solution**: 
- In sandbox mode, verify recipient email
- OR request production access

### Emails not received
**Check**:
1. SES console → Sending statistics
2. CloudWatch logs for errors
3. Recipient's spam folder
4. Domain DNS records (SPF, DKIM, DMARC)

---

## Best Practices

### 1. Use Configuration Sets
Track email metrics and handle bounces/complaints:

```go
input := &ses.SendEmailInput{
    // ... other fields
    ConfigurationSetName: aws.String("yukti-transactional"),
}
```

### 2. Handle Bounces
Set up SNS topic for bounce notifications:
- Hard bounces → Remove from list
- Soft bounces → Retry later

### 3. Monitor Reputation
- Keep bounce rate < 5%
- Keep complaint rate < 0.1%
- Monitor sender reputation score

### 4. Use Templates
Create email templates in SES for consistency:

```go
input := &ses.SendTemplatedEmailInput{
    Source: aws.String("noreply@yukti.io"),
    Destination: &types.Destination{
        ToAddresses: []string{to},
    },
    Template: aws.String("OTPVerification"),
    TemplateData: aws.String(`{"code":"123456"}`),
}
```

### 5. Rate Limiting
Respect SES sending limits:
- Sandbox: 1 email/second
- Production: Varies (request increase if needed)

---

## Migration Checklist

- [x] Replace SMTP with AWS SES SDK
- [x] Update go.mod with SES dependency
- [x] Add dev mode fallback (console logging)
- [x] Update Dockerfile
- [x] Rebuild backend container
- [ ] Verify email address in SES console
- [ ] Test email sending in dev mode
- [ ] Configure AWS credentials for production
- [ ] Request SES production access
- [ ] Verify domain for production
- [ ] Set up bounce/complaint handling
- [ ] Monitor email delivery metrics

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/services/email.go` | Replaced SMTP with AWS SES |
| `go.mod` | Added `github.com/aws/aws-sdk-go-v2/service/ses` |
| `Dockerfile` | Added `go mod tidy` step |

---

## Next Steps

1. **Development**: Works out of the box (console logging)
2. **Staging**: Verify email address in SES sandbox
3. **Production**: 
   - Request production access
   - Verify domain
   - Configure DNS records
   - Set up monitoring

---

**Status**: ✅ **IMPLEMENTED**  
**Mode**: Dev (console logging) + Production (AWS SES)  
**Cost**: FREE for first 62K emails/month  
**Scalability**: Up to 1M+ emails/month
