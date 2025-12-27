package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// constants
const (
	shardcount     = 1024
	accountcount   = 10_000
	initialbalance = 10_000
	transferamount = 1
	maxqueuesize   = 10_000_000
)

// transaction
type transaction struct {
	ID          uint64 `json:"id"`
	From        uint32 `json:"from"`
	To          uint32 `json:"to"`
	Amount      int64  `json:"amount"`
	SubmittedBy string `json:"submitted_by"`
	Ts          int64  `json:"ts"`
}

type account struct {
	mu      sync.Mutex
	balance int64
}

type ledgershard struct {
	accounts map[uint32]*account
	mu       sync.RWMutex
}

// global state
var (
	ledger [shardcount]*ledgershard

	txcounter      atomic.Uint64
	processedtx    atomic.Int64
	failedtx       atomic.Int64
	inflighttx     atomic.Int64 // this is the queue backlog
	activeworkers  atomic.Int64 // tracking current execution state
	globalengine   *engine
	workerpoolsize atomic.Int64
	txevents       = make(chan transaction, 100000)

	clients   = make(map[*client]struct{})
	clientsmu sync.Mutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type client struct {
	send chan []transaction
}

func getshardindex(id uint32) uint32 {
	return id % shardcount
}

func processtransaction(tx transaction) {
	if tx.From == tx.To {
		processedtx.Add(1)
		return
	}

	idxa := getshardindex(tx.From)
	idxb := getshardindex(tx.To)

	firstidx, secondidx := idxa, idxb
	if idxa > idxb {
		firstidx, secondidx = idxb, idxa
	}

	if idxa != idxb {
		ledger[firstidx].mu.RLock()
		ledger[secondidx].mu.RLock()
		defer ledger[firstidx].mu.RUnlock()
		defer ledger[secondidx].mu.RUnlock()
	} else {
		ledger[firstidx].mu.RLock()
		defer ledger[firstidx].mu.RUnlock()
	}

	accta := ledger[idxa].accounts[tx.From]
	acctb := ledger[idxb].accounts[tx.To]

	if tx.From < tx.To {
		accta.mu.Lock()
		acctb.mu.Lock()
	} else {
		acctb.mu.Lock()
		accta.mu.Lock()
	}
	defer accta.mu.Unlock()
	defer acctb.mu.Unlock()

	if accta.balance < tx.Amount {
		failedtx.Add(1)
		return
	}

	accta.balance -= tx.Amount
	acctb.balance += tx.Amount
	processedtx.Add(1)

	select {
	case txevents <- tx:
	default:
	}
}

type engine struct {
	jobs chan transaction
}

func newengine(workers int) *engine {
	e := &engine{
		jobs: make(chan transaction, maxqueuesize),
	}
	workerpoolsize.Store(int64(workers))
	for i := 0; i < workers; i++ {
		go e.worker()
	}
	return e
}

func (e *engine) worker() {
	for tx := range e.jobs {
		activeworkers.Add(1) // increment when starting work
		processtransaction(tx)
		inflighttx.Add(-1)
		activeworkers.Add(-1) // decrement when done
		
		time.Sleep(50 * time.Microsecond)
	}
}

func submitbatch(eng *engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request", 400)
			return
		}

		if inflighttx.Load()+int64(payload.Count) > maxqueuesize {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		inflighttx.Add(int64(payload.Count))
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		go func(count int, submitter string) {
			for i := 0; i < count; i++ {
				eng.jobs <- transaction{
					ID:          txcounter.Add(1),
					From:        uint32(rand.Intn(accountcount)),
					To:          uint32(rand.Intn(accountcount)),
					Amount:      transferamount,
					SubmittedBy: submitter,
					Ts:          time.Now().UnixMilli(),
				}
			}
		}(payload.Count, ip)

		w.WriteHeader(http.StatusAccepted)
	}
}

func wsmetrics(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		stats := map[string]interface{}{
			"cpu_threads": runtime.GOMAXPROCS(0),
			"goroutines":  runtime.NumGoroutine(),
			"worker_pool": workerpoolsize.Load(),
			"failed":      failedtx.Load(),
			"processed":   processedtx.Load(),
			"queue_len":   inflighttx.Load(),
			"queue_cap":   maxqueuesize,
			// active_workers is kept in the engine but not sent as "in_flight"
		}
		if err := conn.WriteJSON(stats); err != nil {
			break
		}
	}
}

func wstxstream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &client{send: make(chan []transaction, 128)}
	clientsmu.Lock()
	clients[c] = struct{}{}
	clientsmu.Unlock()

	defer func() {
		clientsmu.Lock()
		delete(clients, c)
		clientsmu.Unlock()
		conn.Close()
	}()

	for batch := range c.send {
		if err := conn.WriteJSON(batch); err != nil {
			break
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	for i := range shardcount {
		ledger[i] = &ledgershard{accounts: make(map[uint32]*account)} 
	}

	for i := range accountcount {
		id := uint32(i)
		idx := getshardindex(id)
		ledger[idx].accounts[id] = &account{balance: initialbalance}
	}

	optimalworkers := runtime.NumCPU() * 4
	globalengine = newengine(optimalworkers)

	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		var batch []transaction
		for {
			select {
			case tx := <-txevents:
				batch = append(batch, tx)
				if len(batch) > 5000 {
					flushbatch(&batch)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flushbatch(&batch)
				}
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", submitbatch(globalengine))
	mux.HandleFunc("/ws/metrics", wsmetrics)
	mux.HandleFunc("/ws/tx", wstxstream)
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		processedtx.Store(0)
		failedtx.Store(0)
		inflighttx.Store(0)
		w.WriteHeader(http.StatusOK)
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		mux.ServeHTTP(w, r)
	})

	fmt.Printf("engine live on :8080 [workers: %d] [cores: %d]\n", optimalworkers, runtime.NumCPU())
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func flushbatch(batch *[]transaction) {
	clientsmu.Lock()
	payload := make([]transaction, len(*batch))
	copy(payload, *batch)
	for c := range clients {
		select {
		case c.send <- payload:
		default:
		}
	}
	clientsmu.Unlock()
	*batch = nil
}
