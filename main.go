package main

import (
	"log"
	"time"

	"github.com/Rissochek/kaspersky-sandbox/utils"
	"github.com/Rissochek/kaspersky-sandbox/pool"
)

type Server struct {
	pool pool.PoolInterface
}

func NewServer(pool pool.PoolInterface) *Server {
	return &Server{pool: pool}
}

func main() {
	workers := utils.GetKeyFromEnv("WORKERS")
	queueSize := utils.GetKeyFromEnv("QUEUE_SIZE")
	workerPool := pool.InitPool(workers, queueSize)

	//количество задач которое нужно выполнить
	tasksTODO := utils.GetKeyFromEnv("TASKS")

	for i := range tasksTODO {
		log.Printf("%v task is submitting", i)
		time.Sleep(200 * time.Millisecond)
		if err := workerPool.Submit(pool.SomeTask); err != nil {
			return
		}
	}
	//проверяем корректную отработку логики остановки
	workerPool.Stop()
}
