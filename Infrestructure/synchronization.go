package infrastructure

import (
	"fmt"
	"math/rand"
	"time"

	domain "simulador/Domain"
)

func GenerateVehicles(entryChannel chan<- *domain.Car, numVehicles int) {
	id := 1
	for i := 0; i < numVehicles; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		vehicle := &domain.Car{State: "Waiting"}
		fmt.Printf("Generando vehículo %d\n", id)
		entryChannel <- vehicle
		fmt.Printf("Vehículo %d enviado al canal\n", id)
		id++
	}

	close(entryChannel)
	fmt.Println("Todos los vehículos han sido generados.")
}
