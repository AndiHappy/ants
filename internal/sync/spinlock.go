// Copyright 2019 Andy Pan & Dietoad. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sync

import (
	"runtime"
	"sync"
	"sync/atomic"
)

/*
自旋锁，主要是：atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1)，主要就是 CAS 的操作。
*/
type spinLock uint32

const maxBackoff = 16

func (sl *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		// Leverage the exponential backoff algorithm, see https://en.wikipedia.org/wiki/Exponential_backoff.
		// 指数退避算法
		for i := 0; i < backoff; i++ {
			//Gosched
			//暂停当前goroutine，使其他goroutine先行运算。只是暂停，
			//不是挂起，当时间片轮转到该协程时，Gosched()后面的操作将自动恢复
			//类似 sleep
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff <<= 1
		}
	}
}

func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

// NewSpinLock instantiates a spin-lock.
func NewSpinLock() sync.Locker {
	return new(spinLock)
}
