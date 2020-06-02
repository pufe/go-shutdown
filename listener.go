package shutdown

import "context"

type IListener interface {
	Name() string
	Listen() error
	IsDown() bool
	SetDown()
	Shutdown(ctx context.Context) error
}

type Listener struct {
	name         string
	listenFunc   func() error
	down         bool
	shutdownFunc func(ctx context.Context) error
}


func (l *Listener) Name() string {
	return l.name
}
func (l *Listener) Listen() error {
	return l.listenFunc()
}
func (l *Listener) Shutdown(ctx context.Context) error {
	l.down = true
	return l.shutdownFunc(ctx)
}

func (l *Listener) IsDown() bool {
	return l.down
}

func (l *Listener) SetDown() {
	l.down = true
}
