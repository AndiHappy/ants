// MIT License

// Copyright (c) 2018 Andy Pan

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ants

import (
	"fmt"
	"runtime/debug"
	"time"
)

// goWorker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type goWorker struct {
	// 为了区分具体的 worker
	uid string

	// pool who owns this worker.
	pool *Pool

	// task is a job should be done.
	task chan func()

	// lastUsed will be updated when putting a worker back into queue.
	lastUsed time.Time
}

func (w *goWorker) String() string {
	return w.uid
}

func (w *goWorker) setUid(uid string) {
	w.uid = uid
}

func (w *goWorker) getUid() string {
	return w.uid
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *goWorker) run() {
	w.pool.addRunning(1) // 首先是 pool 的执行中的 worker +1
	go func() {          // 启动 goroutine
		defer func() { // 使用 defer 方法，来确定goWorker 负者的 task 执行完毕之后的内容
			w.pool.addRunning(-1)         // goWorker 负者的 task 执行完毕,正在执行的 worker -1
			w.pool.workerCache.Put(w)     // goWorker 放到 goWorker 的 cache pool 中
			if p := recover(); p != nil { //异常 panic 处理
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.options.Logger.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
			fmt.Printf("worker:%s run over!\n", w)
			w.pool.cond.Signal() //通知等待着，现在已经有空闲的 worker
		}()

		//goroutine 的执行过程，w.task为chan，task的队列
		var taskRunTime = 1
		for f := range w.task {
			if f == nil {
				fmt.Println("worker run over!:", taskRunTime)
				return
			}
			f() //执行 goWorker 中任务
			fmt.Printf("worker:%s第%d 执行 range w.task \n", w, taskRunTime)
			//为什么这里，恢复worker 在 for 循环里面？
			//任务执行完成之后调用池对象的revertWorker()方法将该goWorker对象放回池中，以便下次取出处理新的任务
			ok := w.pool.revertWorker(w)
			if !ok {
				fmt.Println("worker运行结束:", taskRunTime)
				return
			}
			taskRunTime++
		}
		fmt.Printf("goroutine over!")
	}()
}

func (w *goWorker) finish() {
	w.task <- nil
}

func (w *goWorker) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorker) inputFunc(fn func()) {
	w.task <- fn
}

func (w *goWorker) inputParam(interface{}) {
	panic("unreachable")
}
