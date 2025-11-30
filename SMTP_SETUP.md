# SMTP Setup Guide for Yukti FinOps

## Quick Setup Options

### Option 1: Gmail SMTP (Recommended for Development)

1. **Enable 2-Factor Authentication** on your Gmail account
2. **Generate App Password**:
   - Go to Google Account Settings → Security
   - Under "2-Step Verification", click "App passwords"
   - Select "Mail" and generate password
3. **Update Environment Variables**:
   ```bash
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-16-digit-app-password
   FROM_EMAIL=your-email@gmail.com
   FROM_NAME=Yukti FinOps
   ```

### Option 2: SendGrid (Recommended for Production)

1. **Create SendGrid Account** at sendgrid.com
2. **Generate API Key**:
   - Go to Settings → API Keys
   - Create API Key with "Mail Send" permissions
3. **Update Environment Variables**:
   ```bash
   SMTP_HOST=smtp.sendgrid.net
   SMTP_PORT=587
   SMTP_USERNAME=apikey
   SMTP_PASSWORD=your-sendgrid-api-key
   FROM_EMAIL=noreply@yourdomain.com
   FROM_NAME=Yukti FinOps
   ```

### Option 3: AWS SES (Enterprise)

1. **Setup AWS SES** in your AWS account
2. **Verify Domain/Email** in SES console
3. **Create SMTP Credentials**:
   - Go to SES → SMTP Settings
   - Create SMTP credentials
4. **Update Environment Variables**:
   ```bash
   SMTP_HOST=email-smtp.us-east-1.amazonaws.com
   SMTP_PORT=587
   SMTP_USERNAME=your-ses-smtp-username
   SMTP_PASSWORD=your-ses-smtp-password
   FROM_EMAIL=noreply@yourdomain.com
   FROM_NAME=Yukti FinOps
   ```

## Development Mode

For development/testing, set `DEV_MODE=true` to log emails to console instead of sending:

```bash
DEV_MODE=true
```

## Testing SMTP Configuration

1. **Start the backend** with your SMTP configuration
2. **Try signup** with a real email address
3. **Check logs** for email delivery status
4. **Verify OTP** functionality works

## Security Best Practices

- ✅ Use App Passwords (not account passwords)
- ✅ Store credentials in environment variables
- ✅ Use TLS/SSL encryption (port 587)
- ✅ Implement rate limiting for OTP requests
- ✅ Set OTP expiration (10 minutes)
- ✅ Limit OTP attempts (5 max)

## Email Templates

The system includes professional HTML email templates for:
- ✅ OTP Verification codes
- ✅ Welcome emails
- ✅ Password reset (future)
- ✅ Billing notifications (future)

## Troubleshooting

**Common Issues:**
- `Authentication failed`: Check username/password
- `Connection timeout`: Check SMTP host/port
- `TLS errors`: Ensure port 587 with STARTTLS
- `Rate limiting`: Some providers limit emails per hour

**Gmail Specific:**
- Must use App Password (not account password)
- Enable "Less secure app access" if using account password
- Check Gmail's sending limits (500 emails/day)

## Production Recommendations

1. **Use dedicated email service** (SendGrid, AWS SES, Mailgun)
2. **Setup SPF/DKIM records** for your domain
3. **Monitor email delivery rates**
4. **Implement email bounce handling**
5. **Use separate email for different types** (OTP, marketing, etc.)

---

**Ready to test!** Set your SMTP credentials and try the signup flow.