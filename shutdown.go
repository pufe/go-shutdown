package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Service struct {
	name     string
	function func() error
}

func (manager *Manager) PingService(serviceName string, pingFunc func() error) *Manager {
	manager.servicesToPing = append(manager.servicesToPing,
		Service{name: serviceName, function: pingFunc},
	)
	return manager
}

func (manager *Manager) PingCloseService(serviceName string, pingFunc func() error, closeFunc func() error) *Manager {
	return manager.PingService(serviceName, pingFunc).CloseService(serviceName, closeFunc)
}

func (manager *Manager) CloseService(serviceName string, closeFunc func() error) *Manager {
	manager.servicesToClose = append(manager.servicesToClose,
		Service{name: serviceName, function: closeFunc},
	)
	return manager
}

func (manager *Manager) Listener(
	name string,
	listenFunc func() error,
	shutdownFunc func(ctx context.Context) error,
) *Manager {
	manager.listeners = append(manager.listeners,
		&Listener{name: name, listenFunc: listenFunc, shutdownFunc: shutdownFunc},
	)

	return manager
}

type Manager struct {
	osDone          chan os.Signal
	appDone         chan bool
	isClosing       bool
	servicesToPing  []Service
	servicesToClose []Service
	context         context.Context
	cancel func()
	listeners       []*Listener
}

func Manage(timeout time.Duration) *Manager {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	app := &Manager{
		context:         ctx,
		cancel: cancel,
		osDone:          make(chan os.Signal, 1),
		appDone:         make(chan bool, 1),
		servicesToClose: make([]Service, 0),
		servicesToPing:  make([]Service, 0),
		listeners:       make([]*Listener, 0),
	}

	signal.Notify(app.osDone, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	return app
}

func (manager *Manager) runListeners() {
	closedListenerChan := make(chan *Listener, 1)
	for _, listener := range manager.listeners {
		go func(listener *Listener) {
			if err := listener.Listen(); err != nil {
				logger.Error(err, fmt.Sprintf("Listener \"%s\", stop working", listener.Name()))
			}
			listener.SetDown()
			closedListenerChan <- listener
		}(listener)
	}
	<-closedListenerChan
	if manager.isClosing == false {
		manager.appDone <- true
	}
}

func (manager *Manager) ping() (failures bool) {
	wg := new(sync.WaitGroup)

	for _, service := range manager.servicesToPing {
		wg.Add(1)
		go func(service Service) {
			if err := service.function(); err != nil {
				logger.Error(err, fmt.Sprintf("error when pinging %s error", service.name))
				failures = true
			} else {
				logger.Info(fmt.Sprintf("successful ping to service %s", service.name))
			}
			wg.Done()
		}(service)
	}

	wg.Wait()

	return
}

func (manager *Manager) close() {
	wg := new(sync.WaitGroup)

	for _, service := range manager.servicesToClose {
		wg.Add(1)
		go func(service Service) {
			if err := service.function(); err != nil {
				logger.Error(err, fmt.Sprintf("error when closing %s error", service.name))
			} else {
				logger.Info(fmt.Sprintf("%s was closed gracefully", service.name))
			}
			wg.Done()
		}(service)
	}

	wg.Wait()
}

func (manager *Manager) shutdown() {
	wg := new(sync.WaitGroup)
	for _, listener := range manager.listeners {
		if listener.IsDown() {
			continue
		}
		wg.Add(1)
		go func(listener *Listener) {
			if err := listener.Shutdown(manager.context); err != nil {
				logger.Error(err, fmt.Sprintf("error when shutting down %s", listener.Name()))
			}
			wg.Done()
		}(listener)
	}
	wg.Wait()
}

func (manager *Manager) Listen() {
	if manager.ping() == true {
		return
	}
	go manager.runListeners()
	select {
	case <-manager.osDone:
	case <-manager.appDone:
	}
	manager.isClosing = true

	logger.Info("gracefully stopping...")
	manager.shutdown()
	manager.close()

	logger.Info("application stop working, services ended")
}
