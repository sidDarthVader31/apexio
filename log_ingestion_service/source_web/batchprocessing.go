package main

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string           `json:"message"`
	Level     string           `json:"level"`
	Metadata  map[string]string `json:"metadata"`
}

type BatchProcessor struct {
	producer    sarama.SyncProducer
	topic       string
	batchSize   int
	batchWindow time.Duration
	buffer      []LogEntry
	bufferMu    sync.Mutex
	done        chan struct{}
	
	// Channel to receive logs
	logChan chan LogEntry
}

func NewBatchProcessor(producer sarama.SyncProducer, topic string, batchSize int, batchWindow time.Duration) *BatchProcessor {
	bp := &BatchProcessor{
		producer:    producer,
		topic:       topic,
		batchSize:   batchSize,
		batchWindow: batchWindow,
		buffer:      make([]LogEntry, 0, batchSize),
		logChan:     make(chan LogEntry, batchSize*2), // Buffer channel to handle spikes
		done:        make(chan struct{}),
	}

	// Start the background processor
	go bp.processLogs()
	return bp
}

func (bp *BatchProcessor) processLogs() {
	ticker := time.NewTicker(bp.batchWindow)
	defer ticker.Stop()

	for {
		select {
		case log := <-bp.logChan:
			bp.bufferMu.Lock()
			bp.buffer = append(bp.buffer, log)
			
			// If buffer is full, flush immediately
			if len(bp.buffer) >= bp.batchSize {
				bp.flush()
			}
			bp.bufferMu.Unlock()

		case <-ticker.C:
			// Time window expired, flush if there are any logs
			bp.bufferMu.Lock()
			if len(bp.buffer) > 0 {
				bp.flush()
			}
			bp.bufferMu.Unlock()

		case <-bp.done:
			// Final flush before shutting down
			bp.bufferMu.Lock()
			if len(bp.buffer) > 0 {
				bp.flush()
			}
			bp.bufferMu.Unlock()
			return
		}
	}
}

func (bp *BatchProcessor) flush() {
	if len(bp.buffer) == 0 {
		return
	}

	// Convert logs to Kafka messages
	messages := make([]*sarama.ProducerMessage, len(bp.buffer))
	for i, log := range bp.buffer {
		value, err := json.Marshal(log)
		if err != nil {
			// Handle error - could add to error channel or dead letter queue
			continue
		}

		messages[i] = &sarama.ProducerMessage{
			Topic: bp.topic,
			Value: sarama.ByteEncoder(value),
			// Optional: Add key for partitioning
			Key: sarama.StringEncoder(log.Metadata["service"]),
		}
	}

	// Send batch to Kafka
	err := bp.producer.SendMessages(messages)
	if err != nil {
		// Handle error - could implement retry logic here
		// For now, just log the error
		log.Printf("Failed to send messages: %v", err)
	}

	// Clear the buffer
	bp.buffer = make([]LogEntry, 0, bp.batchSize)
}

func (bp *BatchProcessor) AddLog(log LogEntry) error {
	select {
	case bp.logChan <- log:
		return nil
	default:
		// Channel is full - could implement backpressure here
		return errors.New("log channel is full")
	}
}

func (bp *BatchProcessor) Close() error {
	close(bp.done)
	return bp.producer.Close()
}

// HTTP Handler implementation
type LogHandler struct {
	batchProcessor *BatchProcessor
}

func NewLogHandler(producer sarama.SyncProducer, topic string) *LogHandler {
	return &LogHandler{
		batchProcessor: NewBatchProcessor(
			producer,
			topic,
			1000,              // Batch size
			time.Second*5,     // Batch window
		),
	}
}

func (h *LogHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	var log LogEntry
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add timestamp if not present
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Add to batch processor
	if err := h.batchProcessor.AddLog(log); err != nil {
		http.Error(w, "Failed to process log", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
