---
title: Add Resend Email Integration
description: Integrate Resend for transactional emails including verification, password reset, and order confirmations
tags: [email, resend, transactional, notifications, messaging]
category: email
dependencies: []
related: [auth_jwt, payments_stripe]
---

# Add Resend Email Integration

⚡ **Target: Complete in 1 minute**

## Speed Optimization

This template is optimized for fast implementation:
- **Create email service file in one step** (all functions together)
- **Copy code exactly as shown** (production-ready templates included)
- **Skip testing unless user requests it**
- **Update auth routes after email service is created**

Integrate Resend for sending transactional emails including email verification, password reset, welcome emails, and order confirmations.

## Overview

This guide implements:
- Resend API setup and configuration
- HTML email templates (verification, password reset, welcome, order confirmation)
- Error handling and retry logic
- Email delivery status tracking
- Testing strategies (optional)

**Prerequisites:** Existing Node.js/TypeScript application (see `create_web_app`)

**Why Resend:**
- Modern API with excellent developer experience
- Built-in DKIM/SPF/DMARC setup
- High deliverability rates
- Generous free tier (3,000 emails/month)
- TypeScript-first SDK

## Step 1: Get Resend API Key

1. Sign up at https://resend.com
2. Verify your domain (or use onboarding domain for testing)
3. Create an API key at https://resend.com/api-keys
4. Copy the API key (starts with `re_`)

## Step 2: Install Resend SDK

```bash
npm install resend
```

Use the `execute` tool with `run_command` operation.

## Step 3: Configure Environment Variables

Add to `.env`:

```bash
# Resend Configuration
RESEND_API_KEY="re_your_api_key_here"

# Email Settings
FROM_EMAIL="noreply@yourdomain.com"
FROM_NAME="Your App Name"

# App URL (for email links)
APP_URL="http://localhost:3000"
```

**For Development:** Use Resend's test domain (e.g., `onboarding@resend.dev`) initially.

**For Production:** Add and verify your custom domain in Resend dashboard.

## Step 4: Create Email Service

Create `src/services/email.ts`:

```typescript
import { Resend } from 'resend';

const resend = new Resend(process.env.RESEND_API_KEY);

const fromEmail = process.env.FROM_EMAIL || 'onboarding@resend.dev';
const fromName = process.env.FROM_NAME || 'Your App';
const appUrl = process.env.APP_URL || 'http://localhost:3000';

if (!process.env.RESEND_API_KEY) {
  throw new Error('RESEND_API_KEY environment variable is required');
}

// Email sending wrapper with error handling
async function sendEmail(params: {
  to: string;
  subject: string;
  html: string;
  text?: string;
}) {
  try {
    const result = await resend.emails.send({
      from: `${fromName} <${fromEmail}>`,
      to: params.to,
      subject: params.subject,
      html: params.html,
      text: params.text,
    });

    if (result.error) {
      throw new Error(`Resend API error: ${result.error.message}`);
    }

    console.log(`✓ Email sent to ${params.to}: ${params.subject} (ID: ${result.data?.id})`);
    return result.data;
  } catch (error) {
    console.error(`✗ Failed to send email to ${params.to}:`, error);

    // Don't throw - let application continue even if email fails
    // Log for monitoring/debugging
    return null;
  }
}

// Email verification
export async function sendVerificationEmail(email: string, token: string) {
  const verificationUrl = `${appUrl}/auth/verify-email?token=${token}`;

  const html = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Verify Your Email</title>
    </head>
    <body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f4f4f4;">
      <table width="100%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 40px 20px;">
        <tr>
          <td align="center">
            <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
              <!-- Header -->
              <tr>
                <td style="padding: 40px 40px 20px; text-align: center;">
                  <h1 style="margin: 0; color: #333; font-size: 28px; font-weight: 600;">
                    Verify Your Email
                  </h1>
                </td>
              </tr>

              <!-- Body -->
              <tr>
                <td style="padding: 20px 40px; color: #666; font-size: 16px; line-height: 1.6;">
                  <p style="margin: 0 0 20px;">
                    Thank you for signing up! Please verify your email address by clicking the button below.
                  </p>
                  <p style="margin: 0 0 30px;">
                    This link will expire in 24 hours.
                  </p>
                </td>
              </tr>

              <!-- Button -->
              <tr>
                <td style="padding: 0 40px 40px; text-align: center;">
                  <a href="${verificationUrl}" style="display: inline-block; padding: 14px 32px; background-color: #0070f3; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                    Verify Email Address
                  </a>
                </td>
              </tr>

              <!-- Footer -->
              <tr>
                <td style="padding: 20px 40px; border-top: 1px solid #eee; color: #999; font-size: 14px; text-align: center;">
                  <p style="margin: 0 0 10px;">
                    If the button doesn't work, copy and paste this link into your browser:
                  </p>
                  <p style="margin: 0; word-break: break-all;">
                    <a href="${verificationUrl}" style="color: #0070f3; text-decoration: none;">
                      ${verificationUrl}
                    </a>
                  </p>
                </td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    </body>
    </html>
  `;

  const text = `
    Verify Your Email

    Thank you for signing up! Please verify your email address by visiting this link:

    ${verificationUrl}

    This link will expire in 24 hours.

    If you didn't create an account, you can safely ignore this email.
  `;

  return sendEmail({
    to: email,
    subject: 'Verify Your Email Address',
    html,
    text,
  });
}

// Password reset
export async function sendPasswordResetEmail(email: string, token: string) {
  const resetUrl = `${appUrl}/auth/reset-password?token=${token}`;

  const html = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Reset Your Password</title>
    </head>
    <body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f4f4f4;">
      <table width="100%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 40px 20px;">
        <tr>
          <td align="center">
            <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
              <!-- Header -->
              <tr>
                <td style="padding: 40px 40px 20px; text-align: center;">
                  <h1 style="margin: 0; color: #333; font-size: 28px; font-weight: 600;">
                    Reset Your Password
                  </h1>
                </td>
              </tr>

              <!-- Body -->
              <tr>
                <td style="padding: 20px 40px; color: #666; font-size: 16px; line-height: 1.6;">
                  <p style="margin: 0 0 20px;">
                    We received a request to reset your password. Click the button below to create a new password.
                  </p>
                  <p style="margin: 0 0 30px;">
                    This link will expire in 1 hour for security reasons.
                  </p>
                </td>
              </tr>

              <!-- Button -->
              <tr>
                <td style="padding: 0 40px 40px; text-align: center;">
                  <a href="${resetUrl}" style="display: inline-block; padding: 14px 32px; background-color: #0070f3; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                    Reset Password
                  </a>
                </td>
              </tr>

              <!-- Security Note -->
              <tr>
                <td style="padding: 20px 40px; background-color: #fff9e6; border-top: 1px solid #eee; color: #856404; font-size: 14px;">
                  <p style="margin: 0 0 10px; font-weight: 600;">
                    ⚠️ Security Notice
                  </p>
                  <p style="margin: 0;">
                    If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
                  </p>
                </td>
              </tr>

              <!-- Footer -->
              <tr>
                <td style="padding: 20px 40px; border-top: 1px solid #eee; color: #999; font-size: 14px; text-align: center;">
                  <p style="margin: 0 0 10px;">
                    If the button doesn't work, copy and paste this link into your browser:
                  </p>
                  <p style="margin: 0; word-break: break-all;">
                    <a href="${resetUrl}" style="color: #0070f3; text-decoration: none;">
                      ${resetUrl}
                    </a>
                  </p>
                </td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    </body>
    </html>
  `;

  const text = `
    Reset Your Password

    We received a request to reset your password. Visit this link to create a new password:

    ${resetUrl}

    This link will expire in 1 hour for security reasons.

    If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
  `;

  return sendEmail({
    to: email,
    subject: 'Reset Your Password',
    html,
    text,
  });
}

// Welcome email (after verification)
export async function sendWelcomeEmail(email: string, name: string) {
  const dashboardUrl = `${appUrl}/dashboard`;

  const html = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Welcome!</title>
    </head>
    <body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f4f4f4;">
      <table width="100%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 40px 20px;">
        <tr>
          <td align="center">
            <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
              <!-- Header -->
              <tr>
                <td style="padding: 40px 40px 20px; text-align: center;">
                  <h1 style="margin: 0; color: #333; font-size: 32px; font-weight: 600;">
                    🎉 Welcome, ${name}!
                  </h1>
                </td>
              </tr>

              <!-- Body -->
              <tr>
                <td style="padding: 20px 40px; color: #666; font-size: 16px; line-height: 1.6;">
                  <p style="margin: 0 0 20px;">
                    Your account has been successfully verified! You're all set to get started.
                  </p>
                  <p style="margin: 0 0 30px;">
                    We're excited to have you on board. Here are a few things you can do:
                  </p>
                  <ul style="margin: 0 0 30px; padding-left: 20px;">
                    <li style="margin-bottom: 10px;">Complete your profile</li>
                    <li style="margin-bottom: 10px;">Explore the dashboard</li>
                    <li style="margin-bottom: 10px;">Customize your settings</li>
                  </ul>
                </td>
              </tr>

              <!-- Button -->
              <tr>
                <td style="padding: 0 40px 40px; text-align: center;">
                  <a href="${dashboardUrl}" style="display: inline-block; padding: 14px 32px; background-color: #0070f3; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                    Get Started
                  </a>
                </td>
              </tr>

              <!-- Footer -->
              <tr>
                <td style="padding: 20px 40px; border-top: 1px solid #eee; color: #999; font-size: 14px; text-align: center;">
                  <p style="margin: 0;">
                    Need help? Reply to this email or visit our support center.
                  </p>
                </td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    </body>
    </html>
  `;

  const text = `
    Welcome, ${name}!

    Your account has been successfully verified! You're all set to get started.

    We're excited to have you on board. Here are a few things you can do:
    - Complete your profile
    - Explore the dashboard
    - Customize your settings

    Get started: ${dashboardUrl}

    Need help? Reply to this email or visit our support center.
  `;

  return sendEmail({
    to: email,
    subject: `Welcome to ${fromName}!`,
    html,
    text,
  });
}

// Order confirmation
export async function sendOrderConfirmationEmail(params: {
  email: string;
  name: string;
  orderId: string;
  orderDate: string;
  items: Array<{ name: string; quantity: number; price: string }>;
  total: string;
}) {
  const orderUrl = `${appUrl}/orders/${params.orderId}`;

  const itemsHtml = params.items
    .map(
      (item) => `
        <tr>
          <td style="padding: 10px; border-bottom: 1px solid #eee; color: #333;">
            ${item.name} × ${item.quantity}
          </td>
          <td style="padding: 10px; border-bottom: 1px solid #eee; color: #333; text-align: right;">
            $${item.price}
          </td>
        </tr>
      `
    )
    .join('');

  const itemsText = params.items
    .map((item) => `${item.name} × ${item.quantity} - $${item.price}`)
    .join('\n');

  const html = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Order Confirmation</title>
    </head>
    <body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f4f4f4;">
      <table width="100%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 40px 20px;">
        <tr>
          <td align="center">
            <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
              <!-- Header -->
              <tr>
                <td style="padding: 40px 40px 20px; text-align: center;">
                  <h1 style="margin: 0; color: #333; font-size: 28px; font-weight: 600;">
                    Order Confirmed!
                  </h1>
                </td>
              </tr>

              <!-- Body -->
              <tr>
                <td style="padding: 20px 40px; color: #666; font-size: 16px; line-height: 1.6;">
                  <p style="margin: 0 0 20px;">
                    Hi ${params.name},
                  </p>
                  <p style="margin: 0 0 30px;">
                    Thank you for your order! We've received your payment and your order is being processed.
                  </p>
                </td>
              </tr>

              <!-- Order Details -->
              <tr>
                <td style="padding: 0 40px 20px;">
                  <table width="100%" cellpadding="0" cellspacing="0" style="background-color: #f9f9f9; border-radius: 6px; overflow: hidden;">
                    <tr>
                      <td style="padding: 20px; border-bottom: 2px solid #eee;">
                        <p style="margin: 0; color: #999; font-size: 14px;">Order Number</p>
                        <p style="margin: 5px 0 0; color: #333; font-size: 18px; font-weight: 600;">
                          #${params.orderId}
                        </p>
                      </td>
                      <td style="padding: 20px; border-bottom: 2px solid #eee; text-align: right;">
                        <p style="margin: 0; color: #999; font-size: 14px;">Order Date</p>
                        <p style="margin: 5px 0 0; color: #333; font-size: 18px; font-weight: 600;">
                          ${params.orderDate}
                        </p>
                      </td>
                    </tr>
                    ${itemsHtml}
                    <tr>
                      <td style="padding: 20px; color: #333; font-weight: 600; font-size: 18px;">
                        Total
                      </td>
                      <td style="padding: 20px; color: #333; font-weight: 600; font-size: 18px; text-align: right;">
                        $${params.total}
                      </td>
                    </tr>
                  </table>
                </td>
              </tr>

              <!-- Button -->
              <tr>
                <td style="padding: 20px 40px 40px; text-align: center;">
                  <a href="${orderUrl}" style="display: inline-block; padding: 14px 32px; background-color: #0070f3; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                    View Order Details
                  </a>
                </td>
              </tr>

              <!-- Footer -->
              <tr>
                <td style="padding: 20px 40px; border-top: 1px solid #eee; color: #999; font-size: 14px; text-align: center;">
                  <p style="margin: 0;">
                    Questions about your order? Contact us at support@yourdomain.com
                  </p>
                </td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    </body>
    </html>
  `;

  const text = `
    Order Confirmed!

    Hi ${params.name},

    Thank you for your order! We've received your payment and your order is being processed.

    Order Details:
    Order Number: #${params.orderId}
    Order Date: ${params.orderDate}

    Items:
    ${itemsText}

    Total: $${params.total}

    View order details: ${orderUrl}

    Questions about your order? Contact us at support@yourdomain.com
  `;

  return sendEmail({
    to: params.email,
    subject: `Order Confirmation - #${params.orderId}`,
    html,
    text,
  });
}

// Export the Resend client for advanced usage
export { resend };
```

## Step 5: Replace Auth Email Placeholders

Update `src/auth/email.ts` (from `auth_jwt` template) to use the Resend service:

```typescript
// Replace the entire file content with:
export {
  sendVerificationEmail,
  sendPasswordResetEmail
} from '../services/email';
```

## Step 6: Add Welcome Email to Auth Flow

Update `src/routes/auth.ts` to send welcome email after verification:

```typescript
import { sendWelcomeEmail } from '../services/email';

// In the verify-email endpoint, after updating user:
await server.db
  .update(users)
  .set({ emailVerified: true })
  .where(eq(users.id, tokenRecord.userId));

// Get user details for welcome email
const [user] = await server.db
  .select()
  .from(users)
  .where(eq(users.id, tokenRecord.userId))
  .limit(1);

// Send welcome email
if (user) {
  await sendWelcomeEmail(user.email, user.name);
}
```

## Step 7: Email Integration Complete

Email service is now integrated with your application. Emails will be sent automatically during:
- User signup (verification email)
- Email verification success (welcome email)
- Password reset request (reset email)
- Order completion (order confirmation email)

### Optional: Testing Email Sending

If you want to test emails manually, create `src/test-email.ts`:

```typescript
import 'dotenv/config';
import {
  sendVerificationEmail,
  sendPasswordResetEmail,
  sendWelcomeEmail,
  sendOrderConfirmationEmail
} from './services/email';

async function testEmails() {
  const testEmail = 'your-email@example.com';

  console.log('Testing email sending...\n');

  // Test verification email
  console.log('1. Sending verification email...');
  await sendVerificationEmail(testEmail, 'test-token-123');

  // Test password reset email
  console.log('2. Sending password reset email...');
  await sendPasswordResetEmail(testEmail, 'reset-token-456');

  // Test welcome email
  console.log('3. Sending welcome email...');
  await sendWelcomeEmail(testEmail, 'Test User');

  // Test order confirmation email
  console.log('4. Sending order confirmation email...');
  await sendOrderConfirmationEmail({
    email: testEmail,
    name: 'Test User',
    orderId: 'ORD-2024-001',
    orderDate: new Date().toLocaleDateString(),
    items: [
      { name: 'Product 1', quantity: 2, price: '29.99' },
      { name: 'Product 2', quantity: 1, price: '49.99' },
    ],
    total: '109.97',
  });

  console.log('\n✓ All test emails sent!');
}

testEmails().catch(console.error);
```

Run tests:

```bash
npx tsx src/test-email.ts
```

### Optional: Check Email Delivery

1. Visit https://resend.com/emails
2. View sent emails and their status
3. Check spam folder if emails don't arrive
4. Review bounce/complaint reports

## Step 8: Email Delivery Monitoring

Create a utility to check email status:

```typescript
// src/services/email-status.ts
import { resend } from './email';

export async function getEmailStatus(emailId: string) {
  try {
    const email = await resend.emails.get(emailId);
    return email.data;
  } catch (error) {
    console.error('Failed to get email status:', error);
    return null;
  }
}

// Usage example
const result = await sendVerificationEmail('user@example.com', 'token');
if (result?.id) {
  // Store email ID in database for tracking
  // Later, check status:
  const status = await getEmailStatus(result.id);
  console.log('Email status:', status);
}
```

## Error Handling Patterns

### Graceful Degradation

Emails are non-critical - app should continue even if email fails:

```typescript
// Bad - throws error, breaks signup
await sendVerificationEmail(email, token);

// Good - logs error, continues
try {
  await sendVerificationEmail(email, token);
} catch (error) {
  console.error('Email failed:', error);
  // Maybe log to monitoring service
  // App continues - user can resend later
}
```

### Retry Strategy

For critical emails, implement retry logic:

```typescript
async function sendEmailWithRetry(
  sendFn: () => Promise<any>,
  maxRetries = 3,
  delay = 1000
) {
  let lastError;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await sendFn();
    } catch (error) {
      lastError = error;
      console.log(`Email attempt ${attempt} failed, retrying...`);

      if (attempt < maxRetries) {
        await new Promise(resolve => setTimeout(resolve, delay * attempt));
      }
    }
  }

  console.error('All email retry attempts failed:', lastError);
  return null;
}

// Usage
await sendEmailWithRetry(() =>
  sendVerificationEmail(email, token)
);
```

## Rate Limiting Considerations

Resend has rate limits:
- Free tier: 100 emails/day during trial
- Paid tier: Higher limits based on plan

Implement rate limiting for user-triggered emails:

```typescript
// Example: Limit verification email resends
const RESEND_COOLDOWN = 60 * 1000; // 1 minute
const lastSentMap = new Map<string, number>();

export async function sendVerificationEmailWithRateLimit(
  email: string,
  token: string
) {
  const lastSent = lastSentMap.get(email);
  const now = Date.now();

  if (lastSent && now - lastSent < RESEND_COOLDOWN) {
    throw new Error('Please wait before requesting another email');
  }

  const result = await sendVerificationEmail(email, token);
  lastSentMap.set(email, now);

  return result;
}
```

## Production Best Practices

### Domain Verification

1. Add your domain in Resend dashboard
2. Add DNS records (Resend provides them)
3. Verify domain
4. Update `FROM_EMAIL` to use your domain

### Email Reputation

- Monitor bounce rates (keep < 5%)
- Handle unsubscribes
- Don't send to invalid addresses
- Authenticate your domain (DKIM/SPF/DMARC)

### Logging and Monitoring

```typescript
// Track email metrics
const emailMetrics = {
  sent: 0,
  failed: 0,
  bounced: 0,
};

// In sendEmail function:
if (result.data) {
  emailMetrics.sent++;
} else {
  emailMetrics.failed++;
}

// Expose metrics endpoint
server.get('/metrics/email', async (request, reply) => {
  return emailMetrics;
});
```

### Environment-Specific Behavior

```typescript
const isDevelopment = process.env.NODE_ENV === 'development';
const isTest = process.env.NODE_ENV === 'test';

// Don't send real emails in test environment
if (isTest) {
  console.log('TEST MODE: Email not sent');
  return { id: 'test-email-id' };
}

// In development, log email content
if (isDevelopment) {
  console.log('DEV MODE: Email content:', { to, subject, html });
}
```

## Template Customization

### Branding

Update email templates with your brand:

```typescript
// Add your logo
const logoUrl = 'https://yourdomain.com/logo.png';

// In email HTML:
<img src="${logoUrl}" alt="Logo" style="height: 40px; margin-bottom: 20px;" />

// Brand colors
const brandColor = '#0070f3'; // Replace with your primary color
```

### Email Footer

Add consistent footer to all emails:

```typescript
const emailFooter = `
  <tr>
    <td style="padding: 20px 40px; border-top: 1px solid #eee; background-color: #f9f9f9;">
      <table width="100%" cellpadding="0" cellspacing="0">
        <tr>
          <td style="color: #999; font-size: 12px; text-align: center;">
            <p style="margin: 0 0 10px;">
              © ${new Date().getFullYear()} ${fromName}. All rights reserved.
            </p>
            <p style="margin: 0;">
              <a href="${appUrl}/unsubscribe" style="color: #999; text-decoration: underline;">Unsubscribe</a>
              |
              <a href="${appUrl}/privacy" style="color: #999; text-decoration: underline;">Privacy Policy</a>
            </p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
`;
```

## Troubleshooting

### Emails not arriving

- Check Resend dashboard for delivery status
- Verify FROM_EMAIL domain is verified
- Check spam folder
- Ensure RESEND_API_KEY is correct
- Review DNS records (DKIM/SPF/DMARC)

### High bounce rate

- Validate email addresses before sending
- Remove bounced addresses from database
- Use double opt-in for signups

### Emails marked as spam

- Verify your domain
- Set up DKIM/SPF/DMARC
- Don't use spam trigger words
- Include physical address in footer
- Provide unsubscribe link

### Rate limit errors

- Implement rate limiting in your app
- Upgrade Resend plan if needed
- Use batch sending for multiple recipients

## Advanced Usage

### Batch Sending

```typescript
// Send to multiple recipients
export async function sendBatchEmail(
  recipients: string[],
  subject: string,
  html: string
) {
  const results = await Promise.allSettled(
    recipients.map(email =>
      sendEmail({ to: email, subject, html })
    )
  );

  const succeeded = results.filter(r => r.status === 'fulfilled').length;
  const failed = results.filter(r => r.status === 'rejected').length;

  console.log(`Batch email: ${succeeded} succeeded, ${failed} failed`);
  return { succeeded, failed };
}
```

### Email Templates with Variables

```typescript
function renderTemplate(template: string, variables: Record<string, string>) {
  return template.replace(/\{\{(\w+)\}\}/g, (match, key) => {
    return variables[key] || match;
  });
}

const template = 'Hello {{name}}, your order {{orderId}} is ready!';
const rendered = renderTemplate(template, {
  name: 'John',
  orderId: '12345'
});
```

## Next Steps

- **Analytics:** Track email open rates (Resend provides webhooks)
- **A/B Testing:** Test different email subject lines
- **Personalization:** Use user data for personalized content
- **Unsubscribe Management:** Implement preference center
- **Email Queue:** Use job queue (Bull, BullMQ) for high-volume sending

## Useful Resources

- Resend Documentation: https://resend.com/docs
- Resend API Reference: https://resend.com/docs/api-reference
- Email HTML Best Practices: https://www.campaignmonitor.com/dev-resources/guides/coding/
- Can I Email: https://www.caniemail.com - Email client support reference
