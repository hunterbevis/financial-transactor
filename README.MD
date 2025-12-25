# Financial Transactor Engine

![financial transactor dashboard](https://github.com/user-attachments/assets/4c0d8631-c230-4b36-be5a-e02ba9d47872)

A high-concurrency atomic ledger simulation built with go and sveltekit. This project demonstrates Golang's capability to handle massive transaction volumes using its native concurrency primitives and memory management. You the frontend to send 10k - 100k transactions to the Golang server. Tweak the worker delay time, and number of Goroutines to see how Goroutines impact performance.

## Technical Features

- **High-Concurrency Worker Pool**: Dynamic scaling of goroutines to process transaction batches based on system load.
- **Sharded Ledger**: Utilizes mutex-sharding to minimize lock contention across 10,000 accounts, allowing for parallel writes.
- **Real-Time Heatmap**: A visual representation of ledger state changes streamed via websockets. The heatmap has 10k slots. Each slot represents a bank account each containing `$10k` USD. The first 0-5k accounts (the top half) are sending `$100` per transaction to accounts 5001-10k (bottom half). As account balances go below `$10k`, the box gets more and more red until zero (black at zero). As they get above `$10k` they get more green.
- **Latency Simulation**: Adjustable worker delay to observe how go handles i/o-bound vs cpu-bound tasks. Values are 0-50ms.
- **Atomic Operations**: Ensures data integrity and prevents race conditions during rapid-fire balance transfers.

## Tech Stack

- **Backend**: Golang (goroutines, channels, sharded mutexes)
- **Frontend**: Sveltekit, tailwind css, lucide svelte
- **Communication**: High-frequency websockets for real-time metrics streaming

## Concurrency metrics and monitoring

The engine tracks and visualizes the relationship between different layers of the runtime:

- **CPU Threads**: Physical hardware limits defined by `$GOMAXPROCS$`.
- **Goroutines**: The total number of living routines managed by the go scheduler.
- **Worker Pool**: The active application-level workers pulling from the transaction queue.
- **Shard Lock Contention**: Visual feedback on mutex locks and state synchronization across the ledger shards.

## Getting started

1. **clone the repository**

   ```bash
   git clone https://github.com/hunterbevis/financial-transactor
   cd financial-transactor
   ```

2. Start the go backend

   ```bash
    go run main.go
   ```

3. Start the sveltekit frontend

   ```bash
    cd frontend
    bun install
    bun run dev --open
   ```

4. Access the dashboard navigate to http://localhost:5173
   NOTE: Frontend should open automatically with --open flag
