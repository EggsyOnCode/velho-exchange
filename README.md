# Velho Exchange

A minimal crypto centralized exchange (CEX)-style matching engine with an HTTP API, in-memory order books, a simple market-making loop, and optional ETH transfers against a local dev chain. The project boots an API server, a demo client, and a market maker to continuously seed liquidity and tighten spreads.

## Requirements

- Go 1.22+ (tested with 1.22.2)
- Make (optional but recommended)
- Running JSON-RPC Ethereum node on `http://localhost:8545` (Anvil, Hardhat, or Ganache) if you want token and USD transfer flows to execute fully. Without this, blockchain-dependent calls will fail.

  - Example (recommended) using Anvil:
    ```bash
    anvil -p 8545
    ```

## Repository layout

- `main.go`: Entry point. Starts the HTTP server, creates a demo client, registers users, starts a market maker, and a background market-order placer.
- `api/`
  - `api.go`: Echo HTTP server setup, middleware (CORS), and route registration.
  - `handlers/orderbook.go`: Request/response types and HTTP handlers for users, orders, books, trades, and best bid/ask.
- `core/`
  - `exchange.go`: Exchange state: users, order books per market, and user order indexing. Provides `AddUser`, `AddOrder`, and `GetOrders`.
  - `orderbook.go`: Matching engine and data structures. Defines `Order`, `Limit`, `OrderBook`, `Trade`, and matching logic for LIMIT and MARKET orders; token/USD transfer hooks; best bid/ask; trade history; and current price.
- `auth/`
  - `user.go`: `User` model with ECDSA keypair and USD balance, utilities to generate dev users, and ETH balance queries.
- `internals/`
  - `utils.go`: Utilities for ECDSA keys, Ethereum address derivation, unit conversions, RPC client, gas price, and raw ETH transfers via go-ethereum.
- `client/`
  - `client.go`: Simple HTTP client wrapper for calling the API from Go (used by the market maker and demo flow in `main.go`).
- `market_maker/`
  - `mm.go`: A basic market maker: seeds an initial two-sided book and tightens the spread at an interval using LIMIT orders.
- `bin/`: Build artifacts (`make build` outputs `bin/vleho`).
- `Makefile`: Convenience targets to build, run, and test.

## Runtime architecture

- On `main.go` run:
  1. Starts the API server on `:3000`.
  2. Instantiates a `client.Client` that talks to the local server.
  3. Registers a few users with initial USD balances.
  4. Starts a `market_maker.MarketMaker` which places LIMIT orders around the mid.
  5. Starts a background goroutine that periodically submits MARKET orders to exercise matching.

- Exchange state:
  - Two markets pre-initialized: `BTC` and `ETH` (see `core/exchange.go`). The demo market maker and flows use `ETH`.
  - Each `OrderBook` maintains:
    - `Asks` (ascending by price) and `Bids` (descending by price), both AVL trees of price levels.
    - Each price level (`Limit`) stores FIFO orders keyed by timestamp (implemented as a tree ordered by descending timestamp for efficient head iteration).
    - `OrdersMap` for direct order lookups by UUID.
    - Trade tape (`Trades`) and the latest traded price (`CurrentPrice`).

- Matching & settlement (see `core/orderbook.go`):
  - LIMIT orders rest on the book and adjust aggregate bid/ask volume.
  - MARKET orders sweep opposite-side limits from best price outward until filled or volume exhausted.
  - Post-match, the engine calls settlement hooks to move USD between users and transfer tokens between user and exchange wallets.
  - Token and USD flows:
    - USD is tracked off-chain in memory (`User.USD` and `Exchange.UsdPool`).
    - Token “custody” is simulated via ETH transfers to/from the exchange/user wallets using the dev chain at `:8545`.

## HTTP API

Base URL: `http://localhost:3000`

- Users
  - POST `/user`
    - Body: `{ "private_key": string (hex) | "", "usd": number }`
    - If `private_key` is empty, a new ECDSA key is generated. Returns `{ status, user: <userID> }`.
  - GET `/user/:id`
    - Returns the full user object (including USD; ETH balance is on-chain and not included).

- Orders
  - POST `/order?user=<userID>`
    - Body: `{ "order_type": "LIMIT"|"MARKET", "price": number, "size": int, "bid": bool, "market": "ETH"|"BTC" }`
    - LIMIT returns `{ status: "success", id: <orderID> }`.
    - MARKET returns `{ status: "success", matches: [...] }` or expectation-failed with an error if insufficient volume.
  - DELETE `/order?id=<orderID>&market=<ETH|BTC>`
    - Cancels a resting LIMIT order by ID.
  - GET `/order?userID=<userID>`
    - Returns active orders for the user segregated into `Asks` and `Bids`.

- Order book & prices
  - GET `/orderbook?market=<ETH|BTC>`
    - Returns full book snapshot with `Asks`, `Bids`, and total bid/ask volumes.
  - GET `/book/bid?market=<ETH|BTC>` → `{ price: number }` (best bid; 0 if none).
  - GET `/book/ask?market=<ETH|BTC>` → `{ price: number }` (best ask; 0 if none).
  - GET `/trade?market=<ETH|BTC>`
    - Returns recent trades recorded by the engine.
  - GET `/marketPrice/:id?market=<ETH|BTC>`
    - Returns `{ status, price }` representing the last traded price.

## Build, run, and test

- Build
  ```bash
  make build
  # binary: ./bin/vleho
  ```

- Run (starts server, demo client, market maker, and market-order loop)
  ```bash
  # Ensure a dev Ethereum node is running on :8545 (see Requirements)
  make run
  ```

- Test
  ```bash
  make test
  ```

## Using the HTTP API manually

- Register a user (auto-generate key)
  ```bash
  curl -s -X POST http://localhost:3000/user \
    -H 'Content-Type: application/json' \
    -d '{"private_key":"","usd":100000}'
  ```

- Place a LIMIT bid for 100 units of ETH at price 995
  ```bash
  curl -s -X POST 'http://localhost:3000/order?user=<USER_ID>' \
    -H 'Content-Type: application/json' \
    -d '{"order_type":"LIMIT","price":995,"size":100,"bid":true,"market":"ETH"}'
  ```

- Place a MARKET sell order for 50 units of ETH
  ```bash
  curl -s -X POST 'http://localhost:3000/order?user=<USER_ID>' \
    -H 'Content-Type: application/json' \
    -d '{"order_type":"MARKET","price":0,"size":50,"bid":false,"market":"ETH"}'
  ```

- Get best bid/ask
  ```bash
  curl -s 'http://localhost:3000/book/bid?market=ETH'
  curl -s 'http://localhost:3000/book/ask?market=ETH'
  ```

- Get the order book snapshot
  ```bash
  curl -s 'http://localhost:3000/orderbook?market=ETH'
  ```

- Cancel a LIMIT order
  ```bash
  curl -s -X DELETE 'http://localhost:3000/order?id=<ORDER_ID>&market=ETH'
  ```

## Notable implementation details

- Data structures: price-time priority via AVL trees (`github.com/zyedidia/generic/avl`).
- Matching semantics: MARKET orders walk the book; LIMIT orders rest. After matching, trades are recorded and `CurrentPrice` is updated to the last execution price.
- Settlement:
  - USD ledger: in-memory adjustments between users and the exchange pool.
  - Token transfers: ETH transfers through `internals.TransferETH` on a dev chain. This requires funded keys and a running RPC node.
- Keys used in the demo: `auth.GenerateMM`/`main.initMMs` include static private keys intended for local dev only. Do not use them on public networks.

## Configuration and defaults

- Server: listens on `:3000` (see `api/api.go`).
- Client: uses `http://localhost:3000` (see `client/client.go`).
- Markets: `ETH` and `BTC` are initialized; the demo uses `ETH`.
- Dev chain: expected at `http://localhost:8545` (see `internals/utils.go`).
- Make targets: `build`, `run`, `test`.

## Caveats

- This is an in-memory demo service; no persistence or durability.
- If no Ethereum node is running on `:8545` or keys are unfunded, ETH transfer calls may fail or cause panics due to unchecked errors in utility calls. Run a local node as described or avoid flows that require token movement.
- Not production grade; for learning and experimentation.
