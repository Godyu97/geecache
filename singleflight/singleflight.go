package singleflight

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

type Group struct {
	l sync.Mutex
	m map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.l.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.l.Unlock()
		// 如果请求正在进行中，则等待
		c.wg.Wait()
		// 请求结束，返回结果
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.l.Unlock()
	// 调用 fn，发起请求
	c.val, c.err = fn()
	c.wg.Done()

	g.l.Lock()
	delete(g.m, key)
	g.l.Unlock()

	// 返回结果
	return c.val, c.err
}
