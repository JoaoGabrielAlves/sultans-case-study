# Shopify Case Study

Two tasks using Shopify's GraphQL API.

## Setup

Create a `.env` file:

```bash
SHOPIFY_URL=https://interview-00x037s7zm.myshopify.com
SHOPIFY_ACCESS_TOKEN=your_access_token_here
SHOPIFY_USER_EMAIL=your-email@domain.com
SHOPIFY_STORE_PASSWORD=shopify
```

Run:

```bash
go mod tidy
go run main.go
```

## Task 1: Customer Leaderboard

Finds customers with `task1` and `level:3` tags, sorts by spending, exports top 50 to CSV.

Uses `customerSegmentMembers` query with tag filtering.

## Task 2: Capture the Flag

Finds the flag product, checks for existing orders, provides manual ordering instructions if needed.

Attempts to disable inventory tracking to allow ordering sold-out items.

## Structure

```
├── main.go
├── pkg/shopify/client.go       # GraphQL client
├── tasks/task1/                # Customer leaderboard
└── tasks/task2/                # Flag capture
```

The client handles GraphQL requests and config loading. Each task is self-contained.
