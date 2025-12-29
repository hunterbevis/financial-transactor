# Financial Transactor Engine

![financial transactor dashboard](https://github.com/user-attachments/assets/4c0d8631-c230-4b36-be5a-e02ba9d47872)

A high-concurrency atomic ledger simulation built with go and sveltekit. This project demonstrates Golang's capability to handle massive transaction volumes using its native concurrency primitives and memory management. All connected users can send 10k - 5m transactions to the Golang server and watch everyone's transactions get executed in real-time!

## Technical Features

- **High-Concurrency Worker Pool**: Worker pools are set to be 4x the size of threads of the CPU the server is running on.
- **Sharded Ledger**: Utilizes mutex-sharding to minimize lock contention across 10,000 accounts, allowing for parallel writes.
- **Latency Simulation**: Worker delay to observe how go handles i/o-bound vs cpu-bound tasks. There's a 0.5ms delay per transaction.
- **Atomic Operations**: Ensures data integrity and prevents race conditions during rapid-fire balance transfers.

## Tech Stack

- **Backend**: Golang (goroutines, channels, sharded mutexes)
- **Frontend**: Sveltekit, tailwind css, lucide svelte
- **Communication**: High-frequency websockets for real-time transaction and metrics streaming

## Concurrency metrics and monitoring

the dashboard provides a direct window into the go runtime performance:

- **Concurrency State**: real-time tracking of living **goroutines** vs. active **worker pool** size.
- **System Throughput**: live **velocity (tps)** and **queue** depth monitoring to visualize backpressure handling.
- **Live Transaction Log**: a high-speed stream of incoming transactions showing source/destination accounts and system ray ids.
- **System Metrics**: a view into the shards and accounts on the backend server.

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
