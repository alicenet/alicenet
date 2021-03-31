package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/sirupsen/logrus"
)

// ========================================================
// Task interfaces and structs
// ========================================================

// Task the interface requirements of a task
type Task interface {
	DoDone(*logrus.Logger)
	DoRetry(context.Context, *logrus.Logger, blockchain.Ethereum) bool
	DoWork(context.Context, *logrus.Logger, blockchain.Ethereum) bool
	ShouldRetry(context.Context, *logrus.Logger, blockchain.Ethereum) bool
}

type TaskWrapper struct {
	TaskName string
	TaskRaw  []byte
}

// TaskHandler required functionality of a task
type TaskHandler interface {
	Cancel()
	Start()
	Complete() bool
	Successful() bool
}

// TaskDoFunc is shorthand for func(context.Context) bool, which is what the Do*() task functions are
// -- Return value indicates if task work has completed succesfully
type TaskDoFunc func(context.Context) bool

// TaskShouldFunc is shorthand for func(context.Context) bool, which is the ShouldRetry() type
type TaskShouldFunc func(context.Context) bool

// TaskDoneFunc is shorthand for func() bool, which is the DoDone() type
// -- This is executed as a cleanup, so can't be canceled
type TaskDoneFunc func()

// TaskHandlerDetails contains all the data required to implment a task
type TaskHandlerDetails struct {
	complete   bool
	count      int
	eth        blockchain.Ethereum
	ID         int
	label      string
	logger     *logrus.Logger
	retryDelay time.Duration
	successful bool
	task       Task
	taskCancel context.CancelFunc
	timeout    time.Duration
	wg         *sync.WaitGroup
}

// ========================================================
// Here are the standard functions that orchestrate tasks
// ========================================================

// Start begins orchestrating the task in a goroutine
func (td *TaskHandlerDetails) Start() {

	td.Logger().Infof("Starting...")
	td.wg.Add(1)

	var taskContext context.Context
	taskContext, td.taskCancel = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		if td.wg != nil {
			defer td.wg.Done()
		}
		defer td.wrappedDoDone()
		defer td.taskCancel() // TODO double check

		td.successful = td.wrappedDoWork(ctx)

		for !td.successful && td.wrappedShouldRetry(ctx) {

			err := blockchain.SleepWithContext(ctx, td.retryDelay)
			if err != nil {
				td.Logger().Warnf("Task interupted: %v", ctx.Err())
				return
			}

			td.successful = td.wrappedDoRetry(ctx)
		}

	}(taskContext)
}

// Cancel stops the task's goroutine
func (td *TaskHandlerDetails) Cancel() {
	td.Logger().Infof("Canceling ...")
	if td.taskCancel != nil {
		td.taskCancel()
	}
}

// Successful returns indication of whether the task was eventually successful
func (td *TaskHandlerDetails) Successful() bool {
	return td.successful
}

// Complete returns indication of whether has completed
func (td *TaskHandlerDetails) Complete() bool {
	return td.complete
}

func (td *TaskHandlerDetails) Logger() *logrus.Entry {
	return td.logger.WithField("TaskID", td.label)
}

// ========================================================
// Here are the methods that invoke task specific functionality
// ========================================================

func (td *TaskHandlerDetails) wrappedDoWork(ctx context.Context) bool {

	td.count++

	var resp bool

	ctx, cancelFunc := context.WithTimeout(ctx, td.timeout)
	defer cancelFunc()
	resp = td.task.DoWork(ctx, td.logger, td.eth)

	td.Logger().Infof("wrappedDoWork(...) try %d: %v", td.count, resp)

	return resp
}

func (td *TaskHandlerDetails) wrappedDoRetry(ctx context.Context) bool {

	td.count++

	var resp bool

	ctx, cancelFunc := context.WithTimeout(ctx, td.timeout)
	defer cancelFunc()
	resp = td.task.DoRetry(ctx, td.logger, td.eth)

	td.Logger().Infof("wrappedDoRetry(...) try %d: %v", td.count, resp)

	return resp
}

func (td *TaskHandlerDetails) wrappedShouldRetry(ctx context.Context) bool {

	var resp bool

	ctx, cancelFunc := context.WithTimeout(ctx, td.timeout)
	defer cancelFunc()
	resp = td.task.ShouldRetry(ctx, td.logger, td.eth)

	td.Logger().Infof("wrappedShouldRetry(...): %v", resp)

	return resp
}

func (td *TaskHandlerDetails) wrappedDoDone() {
	td.Logger().Infof("wrappedDoDone(...) tries %d", td.count)
	td.task.DoDone(td.logger)
	td.complete = true
}

// ========================================================
// Manager interfaces and structs
// ========================================================

// Manager describtes the basic functionality of a task Manager
type Manager interface {
	NewTaskHandler(logger *logrus.Logger, eth blockchain.Ethereum, t Task) TaskHandler
	WaitForTasks()
}

// ManagerDetails contains information required for implmentation of task Manager
type ManagerDetails struct {
	wg sync.WaitGroup
}

// ========================================================
// Manager implementation
// ========================================================

// NewManager creates a new Manager
func NewManager() Manager {
	return &ManagerDetails{wg: sync.WaitGroup{}}
}

// NewTaskHandler creates a new task handler, where each phase can take upto 'timeout'
// duration and there is a delay of 'retryDelay' before a retry.
func (md *ManagerDetails) NewTaskHandler(logger *logrus.Logger, eth blockchain.Ethereum, task Task) TaskHandler {

	taskID := rand.Intn(0x10000)
	taskLabel := fmt.Sprintf("0x%04x", taskID)

	logger.Infof("Creating task %v with timeout of %v and retryDelay of %v", taskID, eth.Timeout(), eth.RetryDelay())

	return &TaskHandlerDetails{
		ID:         taskID,
		eth:        eth,
		task:       task,
		label:      taskLabel,
		logger:     logger,
		wg:         &md.wg,
		timeout:    eth.Timeout(),
		retryDelay: eth.RetryDelay()}
}

// WaitForTasks blocks until all tasks associated withis Manager have completed
func (md *ManagerDetails) WaitForTasks() {
	md.wg.Wait()
}

// ========================================================
// Custom Marshal/Unmarshal for tasks
// ========================================================
var taskRegistry struct {
	sync.RWMutex
	a map[reflect.Type]string
	b map[string]reflect.Type
}

func RegisterTask(t Task) {
	taskRegistry.Lock()
	defer taskRegistry.Unlock()

	if taskRegistry.a == nil {
		taskRegistry.a = make(map[reflect.Type]string)
	}

	if taskRegistry.b == nil {
		taskRegistry.b = make(map[string]reflect.Type)
	}

	tipe := reflect.TypeOf(t)
	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}

	taskRegistry.a[tipe] = tipe.String()
	taskRegistry.b[tipe.String()] = tipe
}

func lookupName(tipe reflect.Type) string {
	taskRegistry.RLock()
	defer taskRegistry.RUnlock()

	return taskRegistry.a[tipe]
}

func lookupType(name string) reflect.Type {
	taskRegistry.Lock()
	defer taskRegistry.Unlock()

	return taskRegistry.b[name]
}

func MarshalTask(t Task) ([]byte, error) {

	tipe := reflect.TypeOf(t)
	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}

	name := lookupName(tipe)
	fmt.Printf("Marshaling %v...\n", name)

	rawTask, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	wrapper := TaskWrapper{TaskName: name, TaskRaw: rawTask}

	return json.Marshal(wrapper)
}

func UnmarshalTask(b []byte) (Task, error) {
	wrapper := &TaskWrapper{}
	err := json.Unmarshal(b, wrapper)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Unmarshaling %v from %v\n", wrapper.TaskName, string(wrapper.TaskRaw))

	tipe := lookupType(wrapper.TaskName)
	val := reflect.New(tipe)

	err = json.Unmarshal(wrapper.TaskRaw, val.Interface())
	if err != nil {
		return nil, err
	}

	return val.Interface().(Task), nil
}
