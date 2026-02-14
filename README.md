# Solana Liquidity Migration Indexer

A lightweight real-time Solana indexing engine for detecting Raydium liquidity pool migration events and tracking newly launched tokens.

This project explores event-driven ingestion, normalization of on-chain data, and analytics-ready storage design using a minimal backend architecture.

---

## Overview

The engine monitors the Solana blockchain for Raydium `initialize2` liquidity pool migration events. When detected, it extracts transaction metadata, indexes associated token addresses, and enriches them with market data for tracking and analysis.

The system was built as an experimental backend to explore Solana log subscriptions, ingestion pipelines, and storage modeling patterns.

---

## Architecture

### 1. Real-Time Log Subscription

* Subscribes to Solana `logsSubscribe` via Helius WebSocket
* Filters for Raydium `initialize2: InitializeInstruction2` program logs
* Persists relevant transaction signatures to `rpc_logs` table
* Uses concurrent goroutines for message handling

---

### 2. Transaction Processing

* Fetches full transaction data via Helius Transactions API
* Extracts contract addresses from relevant instructions
* Normalizes token data into structured relational tables
* Persists results in SQLite database

---

### 3. Market Data Enrichment

* Integrates Dexscreener API
* Updates token metadata (symbol, creation date, market cap)
* Stores time-series snapshots in `market_data` table
* Supports periodic metadata refresh jobs

---

### 4. Data Lifecycle Management

* Removes processed log entries
* Applies configurable filtering rules to prune low-signal tokens
* Default filtering:

  * Token age > 48 hours
  * Market cap < $50,000

Configuration is defined in `config.json`.

---

## System Design Principles

* Event-driven ingestion
* Idempotent log processing
* Asynchronous worker routines
* Clear separation between ingestion, enrichment, and cleanup jobs
* Lightweight relational storage for rapid iteration

---

## Data Model

Core tables:

* `rpc_logs` — tracked event signatures
* `tokens` — indexed token metadata
* `market_data` — time-series market metrics

The schema is designed for:

* Fast token lookup
* Market cap filtering
* Historical tracking
* Analytical querying

---

## Tech Stack

* Go (concurrent worker routines)
* SQLite
* Solana RPC (Helius WebSocket + Transactions API)
* Dexscreener API

---

## Purpose

This project was built to explore Solana indexing patterns and analytics modeling around liquidity migration events (e.g., Pump.fun → Raydium transitions). It focuses on ingestion mechanics and storage design rather than production deployment concerns.

---

## Future Improvements

* Replace SQLite with PostgreSQL for higher write throughput
* Introduce persistent job queue for improved fault tolerance
* Add metrics and structured observability
* Containerized deployment
* Horizontal scaling support
