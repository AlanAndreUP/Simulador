package infrastructure

import (
	"fmt"
	application "simulador/App"
	domain "simulador/Domain"
	"sync"
)

func StartParkingControl(service *application.ParkingService, numVehicles int) {
	var wg sync.WaitGroup
	wg.Add(numVehicles)

	go func() {
		for {
			select {
			case vehicle := <-service.EntryChannel:
				go service.EnterParking(vehicle)
			case vehicle := <-service.ExitChannel:
				go func(v domain.Car) {

					service.ExitParking(v)
					wg.Done()
				}(vehicle)

			}
		}
	}()

	wg.Wait()
	close(service.UpdateChannel)
	fmt.Println("Todos los vehículos han salido.")

}
