package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ====================
// domain & setup
// ====================
const (
	shardCount     = 1024
	accountCount   = 10000
	initialBalance = 10_000
	transferAmount = 100
)

type Transaction struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

type Account struct {
	mu      sync.Mutex
	Balance int64
}

type LedgerShard struct {
	accounts map[string]*Account
	mu       sync.RWMutex
}

var (
	ledger            [shardCount]*LedgerShard
	txCounter         atomic.Uint64
	totalTx           atomic.Int64
	processedTx       atomic.Int64
	failedTx          atomic.Int64
	inFlightTx        atomic.Int64
	lastBatchDuration atomic.Int64
	latencyMs         atomic.Int64 
	batchWG           sync.WaitGroup
	workerStatus      []*atomic.Bool
	statusMu          sync.Mutex
	activeWorkerCount atomic.Int64 // tracked for the websocket
)

// ====================
// worker pool
// ====================
type WorkerPool struct {
	jobs    chan Transaction
	workers int
	quit    chan struct{}
	mu      sync.Mutex
}

func NewWorkerPool(size int, queueSize int) *WorkerPool {
	wp := &WorkerPool{
		jobs: make(chan Transaction, queueSize),
		quit: make(chan struct{}),
	}
	wp.resize(size)
	return wp
}

func (wp *WorkerPool) resize(size int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	statusMu.Lock()
	defer statusMu.Unlock()

	delta := size - wp.workers

	if delta > 0 {
		for range delta {
			status := &atomic.Bool{}
			workerStatus = append(workerStatus, status)
			go wp.worker(status)
		}
	} else if delta < 0 {
		for i := 0; i < -delta; i++ {
			wp.quit <- struct{}{}
		}
		workerStatus = workerStatus[:size]
	}
	wp.workers = size
	activeWorkerCount.Store(int64(size))
	log.Printf("pool_resize: %d active workers (goroutines)", size)
}

func (wp *WorkerPool) worker(status *atomic.Bool) {
	for {
		select {
		case tx := <-wp.jobs:
			status.Store(true)
			processTransaction(tx)
			processedTx.Add(1)
			inFlightTx.Add(-1)
			batchWG.Done()
			status.Store(false)
		case <-wp.quit:
			return
		}
	}
}

// ====================
// logic
// ====================
func getShardIndex(id string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(id); i++ {
		h ^= uint32(id[i])
		h *= 16777619
	}
	return h % shardCount
}

func processTransaction(tx Transaction) {
	idxA := getShardIndex(tx.From)
	idxB := getShardIndex(tx.To)

	if idxA < idxB {
		ledger[idxA].mu.RLock()
		ledger[idxB].mu.RLock()
		defer ledger[idxA].mu.RUnlock()
		defer ledger[idxB].mu.RUnlock()
	} else if idxA > idxB {
		ledger[idxB].mu.RLock()
		ledger[idxA].mu.RLock()
		defer ledger[idxB].mu.RUnlock()
		defer ledger[idxA].mu.RUnlock()
	} else {
		ledger[idxA].mu.RLock()
		defer ledger[idxA].mu.RUnlock()
	}

	from := ledger[idxA].accounts[tx.From]
	to := ledger[idxB].accounts[tx.To]

	from.mu.Lock()
	to.mu.Lock()
	defer from.mu.Unlock()
	defer to.mu.Unlock()

	if d := latencyMs.Load(); d > 0 {
		time.Sleep(time.Duration(d) * time.Millisecond)
	}

	if from.Balance < tx.Amount {
		failedTx.Add(1)
		return
	}

	from.Balance -= tx.Amount
	to.Balance += tx.Amount
}

// ====================
// handlers
// ====================
func submit(wp *WorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct{ Count int }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return
		}

		start := time.Now()
		count := payload.Count

		go func() {
			for range count {
				srcIdx := rand.Intn(accountCount / 2)
				snkIdx := (accountCount / 2) + rand.Intn(accountCount / 2)

				tx := Transaction{
					ID:     fmt.Sprintf("tx_%d", txCounter.Add(1)),
					From:   fmt.Sprintf("acct_%d", srcIdx),
					To:     fmt.Sprintf("acct_%d", snkIdx),
					Amount: transferAmount,
				}

				totalTx.Add(1)
				inFlightTx.Add(1)
				batchWG.Add(1)
				wp.jobs <- tx
			}

			batchWG.Wait()
			lastBatchDuration.Store(time.Since(start).Milliseconds())
		}()

		w.WriteHeader(http.StatusAccepted)
	}
}

func updateLatency(w http.ResponseWriter, r *http.Request) {
	var payload struct{ Latency int64 }
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	latencyMs.Store(payload.Latency)
	w.WriteHeader(http.StatusOK)
}

func resetLedger(w http.ResponseWriter, r *http.Request) {
	totalTx.Store(0)
	processedTx.Store(0)
	failedTx.Store(0)
	inFlightTx.Store(0)
	txCounter.Store(0)
	lastBatchDuration.Store(0)

	for i := range shardCount {
		ledger[i].mu.Lock()
		for _, acc := range ledger[i].accounts {
			acc.mu.Lock()
			acc.Balance = initialBalance
			acc.mu.Unlock()
		}
		ledger[i].mu.Unlock()
	}
	w.WriteHeader(http.StatusOK)
}

func metricsWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		bals := make(map[string]int64)
		activeShards := make([]int, 0)

		for i := range accountCount {
			id := fmt.Sprintf("acct_%d", i)
			idx := getShardIndex(id)
			ledger[idx].mu.RLock()
			bals[id] = ledger[idx].accounts[id].Balance
			ledger[idx].mu.RUnlock()
		}

		for i := range shardCount {
			if !ledger[i].mu.TryLock() {
				activeShards = append(activeShards, i)
			} else {
				ledger[i].mu.Unlock()
			}
		}

		statusMu.Lock()
		busyWorkers := make([]bool, len(workerStatus))
		for i, s := range workerStatus {
			busyWorkers[i] = s.Load()
		}
		statusMu.Unlock()

		msg := map[string]any{
			"goroutines":      runtime.NumGoroutine(),
			"worker_threads":  runtime.GOMAXPROCS(0), // actual OS threads limit
			"worker_count":    activeWorkerCount.Load(),
			"total_tx":        totalTx.Load(),
			"processed_tx":    processedTx.Load(),
			"failed_tx":       failedTx.Load(),
			"in_flight_tx":    inFlightTx.Load(),
			"balances":        bals,
			"last_batch_ms":   lastBatchDuration.Load(),
			"worker_status":   busyWorkers,
			"active_shards":   activeShards,
			"latency_ms":      latencyMs.Load(),
		}
		if err := conn.WriteJSON(msg); err != nil {
			break
		}
	}
}

func resizeWorkers(wp *WorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct{ Workers int }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return
		}
		if payload.Workers > 0 {
			wp.resize(payload.Workers)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	latencyMs.Store(1)

	for i := range shardCount {
		ledger[i] = &LedgerShard{accounts: make(map[string]*Account)}
	}

	for i := range accountCount {
		id := fmt.Sprintf("acct_%d", i)
		idx := getShardIndex(id)
		ledger[idx].accounts[id] = &Account{Balance: initialBalance}
	}

	wp := NewWorkerPool(runtime.NumCPU(), 1_000_000)

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", submit(wp))
	mux.HandleFunc("/resize", resizeWorkers(wp))
	mux.HandleFunc("/reset", resetLedger)
	mux.HandleFunc("/latency", updateLatency)
	mux.HandleFunc("/ws/metrics", metricsWS)

	log.Printf("Engine v1.0 running with %d usable cpu threads", runtime.NumCPU())
	http.ListenAndServe(":8080", corsMiddleware(mux))
}