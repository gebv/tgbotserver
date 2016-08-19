package models

import (
	"context"
	"sync"
)

type Handler interface {
	Name() string
	Match(context.Context, *Services) bool
	Handler(context.Context, *Services) error
}

func NewHandlers() *Handlers {
	return &Handlers{}
}

type Handlers struct {
	list   []Handler
	mutext sync.RWMutex
}

func (h *Handlers) Add(handler Handler) *Handlers {
	h.mutext.Lock()
	defer h.mutext.Unlock()

	h.list = append(h.list, handler)

	return h
}

func (h *Handlers) List() []Handler {
	h.mutext.RLock()
	defer h.mutext.RUnlock()

	arr := make([]Handler, len(h.list))
	copy(arr, h.list)

	return arr
}
