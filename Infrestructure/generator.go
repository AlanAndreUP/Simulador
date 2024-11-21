package infrastructure

import (
	"fmt"
	"math/rand"
	application "simulador/App"
	domain "simulador/Domain"
	"sync"
	"time"
)

func StartParkingControl(service *application.ParkingLotService, numVehicles int) {
	var wg sync.WaitGroup
	wg.Add(numVehicles)

	entryChannel := service.GetEntryChannel()
	exitChannel := service.GetExitChannel()

	go func() {
		for i := 0; i < numVehicles; i++ {

			car := domain.Car{
				ID:     i,
				Image:  "/Users/alansanchez/Desktop/Repos/Simulador/Assets/car.png",
				State:  "Waiting",
				Spot:   i,
				Cancel: make(chan bool, 1),
			}

			fmt.Printf("Generando vehículo %d\n", car.ID)
			time.Sleep(time.Duration(rand.Intn(3)+3) * time.Second)
			service.EnterParking(&car)
			fmt.Printf("Vehicle %d entered parking lot.\n", car.ID)
			wg.Add(1)
		}

		close(entryChannel)
		fmt.Println("Canal de entrada cerrado.")
	}()

	go func() {
		for vehicle := range exitChannel {
			go func(v *domain.Car) {
				service.ExitParking(v)
				time.Sleep(time.Duration(rand.Intn(3)+3) * time.Second)
				fmt.Printf("Vehicle %d exited parking lot.\n", v.ID)
				wg.Done()
			}(vehicle)
		}
	}()

	wg.Wait()
	fmt.Println("All vehicles have exited.")
}
