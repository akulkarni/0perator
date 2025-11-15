---
title: Add Stripe Payments
description: Integrate Stripe for one-time payments and recurring subscriptions with webhooks and customer management
tags: [payments, stripe, checkout, subscriptions, billing, webhooks]
category: payments
dependencies: [database_tiger]
related: [create_web_app, auth_jwt, email_resend]
---

# Add Stripe Payments

Integrate Stripe to accept one-time payments and recurring subscriptions using Stripe Checkout with secure webhook handling.

## Overview

This guide implements:
- One-time payments via Stripe Checkout
- Recurring subscriptions (monthly, yearly, etc.)
- Webhook handling for payment events
- Customer management
- Generic payment tracking patterns
- Refund handling
- Customer portal for subscription management

**Prerequisites:**
- Existing Node.js/TypeScript application (see `create_web_app`)
- Database setup with Drizzle ORM (see `database_tiger`)

**Why Stripe Checkout:**
- Hosted payment page (PCI compliant out of the box)
- Handles complex payment flows
- Supports 135+ currencies
- Mobile optimized
- Built-in fraud prevention

## Step 1: Create Stripe Account

1. Sign up at https://stripe.com
2. Verify your email
3. Get API keys at https://dashboard.stripe.com/test/apikeys
4. Copy **Publishable key** (starts with `pk_test_`)
5. Copy **Secret key** (starts with `sk_test_`)

**Note:** Use test mode keys for development.

## Step 2: Install Stripe SDK

```bash
npm install stripe
```

Use the `execute` tool with `run_command` operation.

## Step 3: Configure Environment Variables

Add to `.env`:

```bash
# Stripe Configuration
STRIPE_SECRET_KEY="sk_test_your_secret_key_here"
STRIPE_PUBLISHABLE_KEY="pk_test_your_publishable_key_here"
STRIPE_WEBHOOK_SECRET="whsec_your_webhook_secret_here"

# App URL
APP_URL="http://localhost:3000"

# Existing variables from other templates
DATABASE_URL="..."
```

**Security:** Never expose `STRIPE_SECRET_KEY` or `STRIPE_WEBHOOK_SECRET` to the client.

## Step 4: Initialize Stripe Client

Create `src/services/stripe.ts`:

```typescript
import Stripe from 'stripe';

const stripeSecretKey = process.env.STRIPE_SECRET_KEY;

if (!stripeSecretKey) {
  throw new Error('STRIPE_SECRET_KEY environment variable is required');
}

export const stripe = new Stripe(stripeSecretKey, {
  apiVersion: '2024-11-20.acacia',
  typescript: true,
});

export const STRIPE_WEBHOOK_SECRET = process.env.STRIPE_WEBHOOK_SECRET || '';
```

## Step 5: Database Schema for Payments

Extend your database schema to track payments. These are **generic patterns** - adapt to your specific use case.

Update `src/db/schema.ts`:

```typescript
import { pgTable, bigserial, text, timestamp, numeric, boolean, index } from 'drizzle-orm/pg-core';

// Stripe customers (links your users to Stripe)
export const stripeCustomers = pgTable('stripe_customers', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  userId: bigserial('user_id', { mode: 'number' }).notNull().references(() => users.id, { onDelete: 'cascade' }),
  stripeCustomerId: text('stripe_customer_id').notNull().unique(),
  createdAt: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
}, (table) => ({
  userIdIdx: index('stripe_customers_user_id_idx').on(table.userId),
  stripeCustomerIdIdx: index('stripe_customers_stripe_customer_id_idx').on(table.stripeCustomerId),
}));

// Orders (one-time payments)
export const orders = pgTable('orders', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  userId: bigserial('user_id', { mode: 'number' }).notNull().references(() => users.id),
  stripeCheckoutSessionId: text('stripe_checkout_session_id').unique(),
  stripePaymentIntentId: text('stripe_payment_intent_id').unique(),
  amount: numeric('amount', { precision: 10, scale: 2 }).notNull(),
  currency: text('currency').notNull().default('usd'),
  status: text('status').notNull().default('pending'), // pending, completed, failed, refunded
  metadata: text('metadata'), // JSON string for flexible data
  createdAt: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
  completedAt: timestamp('completed_at', { withTimezone: true }),
}, (table) => ({
  userIdIdx: index('orders_user_id_idx').on(table.userId),
  statusIdx: index('orders_status_idx').on(table.status),
}));

// Subscriptions (recurring payments)
export const subscriptions = pgTable('subscriptions', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  userId: bigserial('user_id', { mode: 'number' }).notNull().references(() => users.id),
  stripeSubscriptionId: text('stripe_subscription_id').notNull().unique(),
  stripeCustomerId: text('stripe_customer_id').notNull(),
  stripePriceId: text('stripe_price_id').notNull(),
  status: text('status').notNull(), // active, canceled, past_due, etc.
  currentPeriodStart: timestamp('current_period_start', { withTimezone: true }).notNull(),
  currentPeriodEnd: timestamp('current_period_end', { withTimezone: true }).notNull(),
  cancelAtPeriodEnd: boolean('cancel_at_period_end').notNull().default(false),
  createdAt: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
  updatedAt: timestamp('updated_at', { withTimezone: true }).defaultNow().notNull(),
}, (table) => ({
  userIdIdx: index('subscriptions_user_id_idx').on(table.userId),
  stripeSubscriptionIdIdx: index('subscriptions_stripe_subscription_id_idx').on(table.stripeSubscriptionId),
  statusIdx: index('subscriptions_status_idx').on(table.status),
}));
```

Push schema changes:

```bash
npx drizzle-kit push
```

## Step 6: Create Products and Prices in Stripe

You can create products via Stripe Dashboard or API. Here's the API approach:

Create `src/scripts/setup-stripe-products.ts`:

```typescript
import 'dotenv/config';
import { stripe } from '../services/stripe';

async function setupProducts() {
  console.log('Creating Stripe products and prices...\n');

  // One-time payment product
  const product1 = await stripe.products.create({
    name: 'Premium Widget',
    description: 'A one-time purchase product',
  });

  const price1 = await stripe.prices.create({
    product: product1.id,
    unit_amount: 4999, // $49.99 in cents
    currency: 'usd',
  });

  console.log(`✓ Created one-time product: ${product1.name}`);
  console.log(`  Price ID: ${price1.id}\n`);

  // Subscription product - Monthly
  const product2 = await stripe.products.create({
    name: 'Pro Plan',
    description: 'Monthly subscription',
  });

  const priceMonthly = await stripe.prices.create({
    product: product2.id,
    unit_amount: 1999, // $19.99/month
    currency: 'usd',
    recurring: {
      interval: 'month',
    },
  });

  console.log(`✓ Created subscription product: ${product2.name}`);
  console.log(`  Monthly Price ID: ${priceMonthly.id}\n`);

  // Subscription product - Yearly
  const priceYearly = await stripe.prices.create({
    product: product2.id,
    unit_amount: 19999, // $199.99/year
    currency: 'usd',
    recurring: {
      interval: 'year',
    },
  });

  console.log(`  Yearly Price ID: ${priceYearly.id}\n`);

  console.log('✓ Setup complete! Use these price IDs in your checkout routes.');
}

setupProducts().catch(console.error);
```

Run once to set up products:

```bash
npx tsx src/scripts/setup-stripe-products.ts
```

**Alternative:** Create products in Stripe Dashboard at https://dashboard.stripe.com/test/products

## Step 7: Create Checkout Routes

Create `src/routes/payments.ts`:

```typescript
import { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import { stripe } from '../services/stripe';
import { authenticateUser } from '../auth/hook';
import { eq } from 'drizzle-orm';
import { users, stripeCustomers } from '../db/schema';

const appUrl = process.env.APP_URL || 'http://localhost:3000';

// Validation schemas
const createCheckoutSchema = z.object({
  priceId: z.string(), // Stripe price ID
  mode: z.enum(['payment', 'subscription']),
  successUrl: z.string().optional(),
  cancelUrl: z.string().optional(),
});

const paymentsRoutes: FastifyPluginAsync = async (server) => {
  // Create Checkout Session (one-time or subscription)
  server.post('/payments/create-checkout', {
    onRequest: authenticateUser,
  }, async (request, reply) => {
    try {
      const data = createCheckoutSchema.parse(request.body);
      const userId = request.user!.userId;

      // Get or create Stripe customer
      let stripeCustomerId: string;

      const [existingCustomer] = await server.db
        .select()
        .from(stripeCustomers)
        .where(eq(stripeCustomers.userId, userId))
        .limit(1);

      if (existingCustomer) {
        stripeCustomerId = existingCustomer.stripeCustomerId;
      } else {
        // Get user details
        const [user] = await server.db
          .select()
          .from(users)
          .where(eq(users.id, userId))
          .limit(1);

        if (!user) {
          reply.status(404).send({ error: 'User not found' });
          return;
        }

        // Create Stripe customer
        const customer = await stripe.customers.create({
          email: user.email,
          metadata: {
            userId: userId.toString(),
          },
        });

        stripeCustomerId = customer.id;

        // Save to database
        await server.db.insert(stripeCustomers).values({
          userId,
          stripeCustomerId,
        });
      }

      // Create Checkout Session
      const session = await stripe.checkout.sessions.create({
        customer: stripeCustomerId,
        mode: data.mode,
        line_items: [
          {
            price: data.priceId,
            quantity: 1,
          },
        ],
        success_url: data.successUrl || `${appUrl}/payment/success?session_id={CHECKOUT_SESSION_ID}`,
        cancel_url: data.cancelUrl || `${appUrl}/payment/canceled`,
        metadata: {
          userId: userId.toString(),
        },
      });

      reply.send({ url: session.url });
    } catch (error) {
      if (error instanceof z.ZodError) {
        reply.status(400).send({ error: 'Invalid input', details: error.errors });
        return;
      }

      server.log.error(error);
      reply.status(500).send({ error: 'Failed to create checkout session' });
    }
  });

  // Get customer portal URL (for subscription management)
  server.get('/payments/customer-portal', {
    onRequest: authenticateUser,
  }, async (request, reply) => {
    try {
      const userId = request.user!.userId;

      // Get Stripe customer
      const [customer] = await server.db
        .select()
        .from(stripeCustomers)
        .where(eq(stripeCustomers.userId, userId))
        .limit(1);

      if (!customer) {
        reply.status(404).send({ error: 'No payment account found' });
        return;
      }

      // Create portal session
      const session = await stripe.billingPortal.sessions.create({
        customer: customer.stripeCustomerId,
        return_url: `${appUrl}/settings/billing`,
      });

      reply.send({ url: session.url });
    } catch (error) {
      server.log.error(error);
      reply.status(500).send({ error: 'Failed to create portal session' });
    }
  });
};

export default paymentsRoutes;
```

Register routes in `src/server.ts`:

```typescript
import paymentsRoutes from './routes/payments';

server.register(paymentsRoutes);
```

## Step 8: Implement Webhook Handler

Webhooks are critical - they notify your app when payments succeed, subscriptions update, etc.

Create `src/routes/webhooks.ts`:

```typescript
import { FastifyPluginAsync } from 'fastify';
import { stripe, STRIPE_WEBHOOK_SECRET } from '../services/stripe';
import { eq } from 'drizzle-orm';
import { orders, subscriptions } from '../db/schema';
import { sendOrderConfirmationEmail } from '../services/email';

const webhooksRoutes: FastifyPluginAsync = async (server) => {
  // Stripe webhook endpoint
  server.post('/webhooks/stripe', {
    config: {
      // Disable body parsing - we need raw body for signature verification
      rawBody: true,
    },
  }, async (request, reply) => {
    const signature = request.headers['stripe-signature'];

    if (!signature) {
      reply.status(400).send({ error: 'Missing stripe-signature header' });
      return;
    }

    let event;

    try {
      // Verify webhook signature
      event = stripe.webhooks.constructEvent(
        request.rawBody!,
        signature,
        STRIPE_WEBHOOK_SECRET
      );
    } catch (error) {
      server.log.error('Webhook signature verification failed:', error);
      reply.status(400).send({ error: 'Invalid signature' });
      return;
    }

    // Handle event (idempotent - safe to process multiple times)
    try {
      switch (event.type) {
        case 'checkout.session.completed': {
          const session = event.data.object;
          await handleCheckoutCompleted(server, session);
          break;
        }

        case 'payment_intent.succeeded': {
          const paymentIntent = event.data.object;
          server.log.info(`PaymentIntent succeeded: ${paymentIntent.id}`);
          break;
        }

        case 'payment_intent.payment_failed': {
          const paymentIntent = event.data.object;
          await handlePaymentFailed(server, paymentIntent);
          break;
        }

        case 'customer.subscription.created':
        case 'customer.subscription.updated': {
          const subscription = event.data.object;
          await handleSubscriptionUpdate(server, subscription);
          break;
        }

        case 'customer.subscription.deleted': {
          const subscription = event.data.object;
          await handleSubscriptionDeleted(server, subscription);
          break;
        }

        case 'invoice.paid': {
          const invoice = event.data.object;
          server.log.info(`Invoice paid: ${invoice.id}`);
          break;
        }

        case 'invoice.payment_failed': {
          const invoice = event.data.object;
          server.log.warn(`Invoice payment failed: ${invoice.id}`);
          break;
        }

        default:
          server.log.info(`Unhandled event type: ${event.type}`);
      }

      reply.send({ received: true });
    } catch (error) {
      server.log.error('Webhook handler error:', error);
      reply.status(500).send({ error: 'Webhook handler failed' });
    }
  });
};

// Handle completed checkout
async function handleCheckoutCompleted(server: any, session: any) {
  const userId = parseInt(session.metadata?.userId || '0');

  if (!userId) {
    server.log.error('Missing userId in checkout session metadata');
    return;
  }

  if (session.mode === 'payment') {
    // One-time payment
    const [order] = await server.db
      .select()
      .from(orders)
      .where(eq(orders.stripeCheckoutSessionId, session.id))
      .limit(1);

    if (!order) {
      // Create order
      await server.db.insert(orders).values({
        userId,
        stripeCheckoutSessionId: session.id,
        stripePaymentIntentId: session.payment_intent,
        amount: (session.amount_total / 100).toFixed(2),
        currency: session.currency,
        status: 'completed',
        metadata: JSON.stringify(session.metadata || {}),
        completedAt: new Date(),
      });

      server.log.info(`Order created for session: ${session.id}`);

      // Send confirmation email (optional)
      // await sendOrderConfirmationEmail(...);
    }
  } else if (session.mode === 'subscription') {
    // Subscription - will be handled by subscription.created event
    server.log.info(`Subscription checkout completed: ${session.id}`);
  }
}

// Handle payment failure
async function handlePaymentFailed(server: any, paymentIntent: any) {
  const [order] = await server.db
    .select()
    .from(orders)
    .where(eq(orders.stripePaymentIntentId, paymentIntent.id))
    .limit(1);

  if (order) {
    await server.db
      .update(orders)
      .set({ status: 'failed' })
      .where(eq(orders.id, order.id));

    server.log.info(`Order marked as failed: ${order.id}`);
  }
}

// Handle subscription update
async function handleSubscriptionUpdate(server: any, subscription: any) {
  const userId = parseInt(subscription.metadata?.userId || '0');

  if (!userId) {
    server.log.error('Missing userId in subscription metadata');
    return;
  }

  // Upsert subscription
  const [existing] = await server.db
    .select()
    .from(subscriptions)
    .where(eq(subscriptions.stripeSubscriptionId, subscription.id))
    .limit(1);

  const subscriptionData = {
    userId,
    stripeSubscriptionId: subscription.id,
    stripeCustomerId: subscription.customer,
    stripePriceId: subscription.items.data[0].price.id,
    status: subscription.status,
    currentPeriodStart: new Date(subscription.current_period_start * 1000),
    currentPeriodEnd: new Date(subscription.current_period_end * 1000),
    cancelAtPeriodEnd: subscription.cancel_at_period_end,
    updatedAt: new Date(),
  };

  if (existing) {
    await server.db
      .update(subscriptions)
      .set(subscriptionData)
      .where(eq(subscriptions.id, existing.id));

    server.log.info(`Subscription updated: ${subscription.id}`);
  } else {
    await server.db.insert(subscriptions).values(subscriptionData);
    server.log.info(`Subscription created: ${subscription.id}`);
  }
}

// Handle subscription deletion
async function handleSubscriptionDeleted(server: any, subscription: any) {
  await server.db
    .update(subscriptions)
    .set({
      status: 'canceled',
      updatedAt: new Date(),
    })
    .where(eq(subscriptions.stripeSubscriptionId, subscription.id));

  server.log.info(`Subscription canceled: ${subscription.id}`);
}

export default webhooksRoutes;
```

**Important:** Configure Fastify to preserve raw body for webhook signature verification.

Update `src/server.ts`:

```typescript
const server = Fastify({
  logger: true,
  // Enable raw body for webhook routes
  bodyLimit: 1048576, // 1MB
  // This requires @fastify/raw-body plugin
});

// Install plugin: npm install @fastify/raw-body
import rawBody from '@fastify/raw-body';

await server.register(rawBody, {
  field: 'rawBody',
  global: false,
  encoding: 'utf8',
  runFirst: true,
  routes: ['/webhooks/stripe'], // Only for webhook route
});

// Register webhook routes
import webhooksRoutes from './routes/webhooks';
server.register(webhooksRoutes);
```

Install raw body plugin:

```bash
npm install @fastify/raw-body
```

## Step 9: Configure Stripe Webhook

### Development (Stripe CLI)

1. Install Stripe CLI: https://stripe.com/docs/stripe-cli
2. Login: `stripe login`
3. Forward webhooks to local server:

```bash
stripe listen --forward-to localhost:3000/webhooks/stripe
```

4. Copy the webhook signing secret (starts with `whsec_`) to `.env`:

```bash
STRIPE_WEBHOOK_SECRET="whsec_..."
```

### Production

1. Go to https://dashboard.stripe.com/webhooks
2. Click "Add endpoint"
3. Enter your webhook URL: `https://yourdomain.com/webhooks/stripe`
4. Select events to listen to (or select "all events")
5. Copy the webhook signing secret to production environment

## Step 10: Frontend Integration

### Client-Side Checkout Flow

```html
<!-- Example: Payment button -->
<button id="checkout-button">Subscribe to Pro Plan</button>

<script>
document.getElementById('checkout-button').addEventListener('click', async () => {
  try {
    const response = await fetch('/payments/create-checkout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`, // From auth_jwt
      },
      body: JSON.stringify({
        priceId: 'price_1234...', // Your Stripe price ID
        mode: 'subscription', // or 'payment' for one-time
      }),
    });

    const { url } = await response.json();

    // Redirect to Stripe Checkout
    window.location.href = url;
  } catch (error) {
    console.error('Checkout failed:', error);
  }
});
</script>
```

### Customer Portal (Manage Subscriptions)

```html
<!-- Example: Manage subscription button -->
<button id="portal-button">Manage Subscription</button>

<script>
document.getElementById('portal-button').addEventListener('click', async () => {
  try {
    const response = await fetch('/payments/customer-portal', {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });

    const { url } = await response.json();

    // Redirect to Stripe Customer Portal
    window.location.href = url;
  } catch (error) {
    console.error('Portal failed:', error);
  }
});
</script>
```

## Step 11: Query Payment Data

```typescript
// Get user's orders
const userOrders = await server.db
  .select()
  .from(orders)
  .where(eq(orders.userId, userId))
  .orderBy(orders.createdAt);

// Get user's active subscription
const [activeSubscription] = await server.db
  .select()
  .from(subscriptions)
  .where(
    and(
      eq(subscriptions.userId, userId),
      eq(subscriptions.status, 'active')
    )
  )
  .limit(1);

// Check if user has active subscription
const hasActiveSubscription = !!activeSubscription;
```

## Step 12: Refund Handling

Create refund endpoint:

```typescript
// In src/routes/payments.ts

server.post('/payments/refund', {
  onRequest: authenticateUser,
}, async (request, reply) => {
  try {
    const { orderId } = z.object({ orderId: z.number() }).parse(request.body);

    // Get order
    const [order] = await server.db
      .select()
      .from(orders)
      .where(eq(orders.id, orderId))
      .limit(1);

    if (!order || order.userId !== request.user!.userId) {
      reply.status(404).send({ error: 'Order not found' });
      return;
    }

    if (order.status !== 'completed') {
      reply.status(400).send({ error: 'Order cannot be refunded' });
      return;
    }

    // Create refund
    const refund = await stripe.refunds.create({
      payment_intent: order.stripePaymentIntentId!,
    });

    // Update order
    await server.db
      .update(orders)
      .set({ status: 'refunded' })
      .where(eq(orders.id, order.id));

    reply.send({ success: true, refund });
  } catch (error) {
    server.log.error(error);
    reply.status(500).send({ error: 'Refund failed' });
  }
});
```

## Testing Payments

### Test Cards

Use these test card numbers:
- **Success:** `4242 4242 4242 4242`
- **Decline:** `4000 0000 0000 0002`
- **Requires authentication:** `4000 0025 0000 3155`

Use any future expiry date, any CVC, and any postal code.

### Testing Webhooks

Trigger test webhooks with Stripe CLI:

```bash
# Test successful payment
stripe trigger payment_intent.succeeded

# Test subscription creation
stripe trigger customer.subscription.created

# Test payment failure
stripe trigger payment_intent.payment_failed
```

## Production Checklist

- [ ] Switch from test keys to live keys in production
- [ ] Configure production webhook endpoint
- [ ] Test webhook signature verification
- [ ] Enable Stripe Radar for fraud prevention
- [ ] Set up email receipts in Stripe Dashboard
- [ ] Configure tax calculation if needed
- [ ] Set up proper error monitoring
- [ ] Test refund flow
- [ ] Configure subscription billing settings

## Security Best Practices

**✅ Do:**
- Always verify webhook signatures
- Use HTTPS in production
- Never expose secret key to client
- Validate amounts server-side
- Use idempotent webhook handlers
- Log all payment events

**❌ Don't:**
- Trust client-side price data
- Skip webhook signature verification
- Store full credit card numbers
- Process webhooks without validation
- Expose Stripe secret key in frontend code

## Common Patterns

### Check Subscription Status

```typescript
async function hasActiveSubscription(userId: number): Promise<boolean> {
  const [sub] = await db
    .select()
    .from(subscriptions)
    .where(
      and(
        eq(subscriptions.userId, userId),
        eq(subscriptions.status, 'active')
      )
    )
    .limit(1);

  return !!sub;
}
```

### Gate Features by Subscription

```typescript
server.get('/premium-feature', {
  onRequest: authenticateUser,
}, async (request, reply) => {
  const hasSubscription = await hasActiveSubscription(request.user!.userId);

  if (!hasSubscription) {
    reply.status(403).send({ error: 'Premium subscription required' });
    return;
  }

  // Premium feature logic...
});
```

## Troubleshooting

### Webhooks not received

- Check webhook endpoint is accessible
- Verify webhook signing secret
- Check Stripe Dashboard webhook logs
- Ensure no firewall blocking

### Payment succeeded but order not created

- Check webhook handler logs
- Verify database connection
- Ensure metadata includes userId
- Check webhook signature verification

### Customer portal errors

- Verify customer exists in Stripe
- Check return_url is valid
- Ensure billing portal is enabled in Stripe Dashboard

### Refund failures

- Check payment status in Stripe Dashboard
- Verify payment_intent ID is correct
- Check Stripe account balance
- Some payment methods don't support refunds

## Next Steps

- **Usage-based billing:** Implement metered billing for API usage
- **Trial periods:** Add free trials to subscriptions
- **Proration:** Handle subscription upgrades/downgrades
- **Invoices:** Customize invoice generation
- **Payment methods:** Save payment methods for future use
- **Analytics:** Track revenue metrics

## Useful Resources

- Stripe API Docs: https://stripe.com/docs/api
- Stripe Checkout: https://stripe.com/docs/payments/checkout
- Stripe Webhooks: https://stripe.com/docs/webhooks
- Test Cards: https://stripe.com/docs/testing
- Stripe CLI: https://stripe.com/docs/stripe-cli
