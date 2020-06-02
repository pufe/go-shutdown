package shutdown
//
//import "sync"
//
//type CloseGroup struct {
//	closableItems []Service
//}
//
//func Group() *CloseGroup {
//	return &CloseGroup{closableItems: make([]Service, 0)}
//}
//
//func (group *CloseGroup) AddService(serviceName string, closeFunc func() error) *CloseGroup {
//	service := Service{name: serviceName, closeFunc: closeFunc}
//	group.closableItems = append(group.closableItems, service)
//
//	return group
//}
//
//func (group *CloseGroup) CloseGroup() {
//	wg := new(sync.WaitGroup)
//
//	for _, service := range group.closableItems {
//		wg.Add(1)
//		go func(service Service) {
//			service.closeService(wg)
//			wg.Done()
//		}(service)
//	}
//
//	wg.Wait()
//}
