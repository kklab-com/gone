package concurrent

import (
	"context"
	"sync"
	"sync/atomic"

	kkpanic "github.com/kklab-com/goth-panic"
)

const FutureRun = 0
const FutureSuccess = 1
const FutureCancel = 2

type Future interface {
	Get() interface{}
	Await() Future
	IsDone() bool
	IsSuccess() bool
	IsCancelled() bool
	Error() error
	AddListener(listener FutureListener) Future
	AddListeners(listeners ...FutureListener) Future
}

type ManualFuture interface {
	Success()
	Cancel()
}

type DefaultFuture struct {
	ctx               context.Context
	cancel            context.CancelFunc
	op                sync.Mutex
	state             int32
	result            interface{}
	err               error
	listeners         []FutureListener
	callListenersOnce sync.Once
}

func NewFuncFuture(f func(f Future) interface{}, ctx context.Context) Future {
	future := &DefaultFuture{}
	if ctx == nil {
		ctx = context.Background()
	}

	future.ctx, future.cancel = context.WithCancel(ctx)
	future.do(f)
	return future
}

func NewFuture(ctx context.Context) Future {
	future := &DefaultFuture{}
	if ctx == nil {
		ctx = context.Background()
	}

	future.ctx, future.cancel = context.WithCancel(ctx)
	return future
}

func NewCancelFuture() Future {
	return &DefaultFuture{
		state: FutureCancel,
	}
}

func NewSuccessFuture() Future {
	return &DefaultFuture{
		state: FutureSuccess,
	}
}

func (d *DefaultFuture) Get() interface{} {
	if d.IsDone() {
		return d.result
	}

	<-d.ctx.Done()
	if !d.IsDone() {
		atomic.StoreInt32(&d.state, FutureCancel)
		d.err = d.ctx.Err()
		d.callListeners()
	}

	return d.result
}

func (d *DefaultFuture) Await() Future {
	d.Get()
	return d
}

func (d *DefaultFuture) IsDone() bool {
	return atomic.LoadInt32(&d.state) > FutureRun
}

func (d *DefaultFuture) IsSuccess() bool {
	return d.state == FutureSuccess
}

func (d *DefaultFuture) IsCancelled() bool {
	return d.state == FutureCancel
}

func (d *DefaultFuture) Error() error {
	return d.err
}

func (d *DefaultFuture) Cancel() {
	d.op.Lock()
	defer d.op.Unlock()
	if !d.IsDone() {
		d.cancel()
	}
}

func (d *DefaultFuture) Success() {
	d.op.Lock()
	defer d.op.Unlock()
	if !d.IsDone() {
		atomic.StoreInt32(&d.state, FutureSuccess)
		d.callListeners()
		d.cancel()
	}
}

func (d *DefaultFuture) AddListener(listener FutureListener) Future {
	d.listeners = append(d.listeners, listener)
	return d
}

func (d *DefaultFuture) AddListeners(listener ...FutureListener) Future {
	d.listeners = append(d.listeners, listener...)
	return d
}

func (d *DefaultFuture) do(f func(f Future) interface{}) Future {
	go func() {
		defer kkpanic.Call(func(r kkpanic.Caught) {
			d.err = r
		})

		d.result = f(d)
		d.op.Lock()
		defer d.op.Unlock()
		if !d.IsDone() {
			atomic.StoreInt32(&d.state, FutureSuccess)
			d.callListeners()
		}

		d.cancel()
	}()

	return d
}

func (d *DefaultFuture) callListeners() {
	d.callListenersOnce.Do(func() {
		for _, listener := range d.listeners {
			listener.OperationCompleted(d)
		}
	})
}
