package singleflight

import "sync"

// call 表示正在处理或已完成的 Do 调用
type call struct {
	wg  sync.WaitGroup // 用于同步等待
	val interface{}    // 调用结果值
	err error          // 调用过程中的错误
}

// Group 代表一类工作，形成一个命名空间，可以在其中执行具有重复抑制的工作单元。
type Group struct {
	mu sync.Mutex       // 用于保护 m
	m  map[string]*call // 惰性初始化
}

// Do 执行给定的函数并返回其结果，确保同一键在某个时刻只有一个执行在进行中。
// 如果有重复调用，重复的调用方会等待原始调用完成并获得相同的结果。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() // 等待原始调用完成
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1) // 添加等待组计数
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn() // 执行函数
	c.wg.Done()         // 函数执行完成，减少等待组计数

	g.mu.Lock()
	delete(g.m, key) // 从 map 中删除已完成的调用
	g.mu.Unlock()

	return c.val, c.err
}
