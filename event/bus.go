package event

type HandlerFunc func(eventName string, payload any)

type Bus interface {
	Emit(eventName string, payload any)
	Subscribe(eventName string, handlers ...HandlerFunc)
}

type eventBus struct {
	h map[string][]HandlerFunc
}

func NewBus() Bus {
	return &eventBus{
		h: make(map[string][]HandlerFunc),
	}
}

func (b *eventBus) Subscribe(eventName string, handlers ...HandlerFunc) {
	b.h[eventName] = append(b.h[eventName], handlers...)
}

func (b *eventBus) Emit(eventName string, payload any) {
	for _, handler := range b.h[eventName] {
		go handler(eventName, payload)
	}
}
