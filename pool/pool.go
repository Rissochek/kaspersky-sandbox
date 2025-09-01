package pool

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type PoolInterface interface {
	Submit(task func()) error
	Stop() error
}

type HookInterface interface {
	ActOnComplete()
}

type Hook struct {
	//sth
}

func NewHook() *Hook{
	return &Hook{}
}

type Pool struct {
	taskQueue    chan func()
	wg           *sync.WaitGroup
	initWg 		 *sync.WaitGroup
	shutdownChan chan struct{}
	hook         *Hook
}

//инициализируем пул + запускаем воркеров
func InitPool(workers int, queueSize int) PoolInterface{
	taskQueue := make(chan func(), queueSize)
	shutdownChan := make(chan struct{})
	wg := &sync.WaitGroup{}
	initWg := &sync.WaitGroup{}
	hook := NewHook()
	pool := NewPool(taskQueue, wg, shutdownChan, hook, initWg)

	for i := range workers {
		wg.Add(1)
		initWg.Add(1)
		go pool.HandleWorker(i)
	}

	log.Println("waiting for all workers are ready")
	initWg.Wait()
	log.Println("all workers are ready to work")
	
	return pool
}

func NewPool(taskQueue chan func(), wg *sync.WaitGroup, shutdownChan chan struct{}, hook *Hook, initWg *sync.WaitGroup) *Pool {
	return &Pool{taskQueue: taskQueue, wg: wg, shutdownChan: shutdownChan, hook: hook, initWg: initWg}
}

func (pool *Pool) Submit(task func()) error {
	select {
	case pool.taskQueue <- task:

	default:
		log.Printf("queue is full")
		return fmt.Errorf("query is full")
	}

	return nil
}

func (pool *Pool) Stop() error {
	log.Printf("stopping accepting new tasks")
	close(pool.shutdownChan)
	pool.wg.Wait()
	return nil
}

func (pool *Pool) HandleWorker(workerId int) {
	log.Printf("worker %v is starting", workerId)
	pool.initWg.Done()
	for {
		//если shutdown не закрыт, то всегда будет выполняться первый case
		select {
		case task, ok := <-pool.taskQueue:
			if !ok {
				log.Printf("task queue closed, worker %v finished", workerId)
				pool.wg.Done()
				return
			}

			pool.HandleTask(task)
		//при закрытии shutdown будет выполняться случайный case если очередь не пуста. Это нас устраивает, так как внутри есть еще проверка на наличие задач в очереди
		//Если же очередь с задачами пуста, то попадая в case с shutdown мы попадем в default, который сообщит о завершении работы горутины с помощью wg.Done.
		case <-pool.shutdownChan:
			select {
			case task, ok := <-pool.taskQueue:
				if !ok {
					log.Printf("task queue closed, worker %v finished", workerId)
					pool.wg.Done()
					return
				}

				pool.HandleTask(task)

			default:
				log.Printf("worker %v stopping due to queue is empty", workerId)
				pool.wg.Done()
				return
			}
		}
	}
}

//обертка для обработки функции. Также здесь вызывает функция из интерфейса хука
func (pool *Pool) HandleTask(task func()) {
	task()
	pool.hook.ActOnComplete()
}

//пример хука (если я правильно понял что подразумевается под хуком)
func (hook *Hook) ActOnComplete() {
	fmt.Println("task is done")
}

//некая логика обработки задачи
func SomeTask() {
	executingTime := time.Duration((rand.Intn(2) + 3)) * time.Second
	time.Sleep(executingTime)
	fmt.Println("Hello from task")
}