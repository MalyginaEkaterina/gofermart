package service

import (
	"context"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
	"log"
	"time"
)

const (
	processAfter = time.Second
)

type OrderWorker struct {
	client AccrualClient
	store  storage.OrderStorage
}

func NewOrderWorker(client AccrualClient, store storage.OrderStorage) *OrderWorker {
	return &OrderWorker{client: client, store: store}
}

func (w *OrderWorker) Run(ctx context.Context) {
	processTick := time.NewTicker(processAfter)
	for {
		select {
		case <-processTick.C:
			w.process(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *OrderWorker) process(ctx context.Context) {
	orders, err := w.store.GetNotProcessedOrders(ctx)
	if err != nil {
		log.Println("Get not processed orders error: ", err)
		return
	}
	log.Printf("Processing of %v orders\n", len(orders))
	for _, order := range orders {
		processedOrder, err := w.processOrder(order)
		if err != nil {
			log.Println("Process order error: ", err)
			continue
		}
		if processedOrder.Accrual != nil {
			err = w.store.UpdateOrderAccrual(ctx, processedOrder)
			if err != nil {
				log.Println("Update order accrual error: ", err)
			}
		} else {
			err = w.store.UpdateOrderStatus(ctx, processedOrder)
			if err != nil {
				log.Println("Update order status error: ", err)
			}
		}
	}
}

func (w *OrderWorker) processOrder(order internal.ProcessingOrder) (internal.ProcessingOrder, error) {
	var result internal.ProcessingOrder
	accrual, err := w.client.GetAccrual(string(order.Number))
	if err != nil {
		return result, err
	}
	result = order
	switch accrual.Status {
	case StatusInvalid:
		result.Status = internal.Invalid
	case StatusProcessing:
		result.Status = internal.Processing
	case StatusProcessed:
		result.Status = internal.Processed
		result.Accrual = accrual.Accrual
	default:
	}
	return result, nil
}
