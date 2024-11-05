package infrastructure

import (
	"math/rand"
	"time"

	domain "simulador/Domain"
)

func GenerateVehicles(entryChannel chan domain.Car, numVehicles int) {
	id := 1
	for i := 0; i < numVehicles; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		vehicle := domain.Car{ID: id, State: "Esperando"}
		entryChannel <- vehicle
		id++
	}
}
