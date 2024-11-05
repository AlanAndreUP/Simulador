package application

import (
	"fmt"
	domain "simulador/Domain"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type UpdateInfo struct {
	Car      domain.Car
	Entering bool
	Spot     int
}

type ParkingService struct {
	Parking       domain.Parking
	EntryChannel  chan domain.Car
	ExitChannel   chan domain.Car
	UpdateChannel chan UpdateInfo
	entryQueue    []domain.Car
	spotAvailable sync.Cond
	queueMutex    sync.Mutex
}

func NewParkingService(capacity int) *ParkingService {
	spots := make([]int, capacity)
	ps := &ParkingService{
		Parking:       domain.Parking{Capacity: capacity, Spots: spots},
		EntryChannel:  make(chan domain.Car),
		ExitChannel:   make(chan domain.Car),
		UpdateChannel: make(chan UpdateInfo),
		entryQueue:    make([]domain.Car, 0),
	}
	ps.spotAvailable.L = &sync.Mutex{}
	return ps
}

func (ps *ParkingService) EnterParking(vehicle domain.Car) {
	ps.spotAvailable.L.Lock()
	defer ps.spotAvailable.L.Unlock()

	if ps.Parking.Vehicles < ps.Parking.Capacity {
		ps.assignSpotAndEnter(vehicle)
	} else {
		ps.queueMutex.Lock()
		ps.entryQueue = append(ps.entryQueue, vehicle)
		ps.queueMutex.Unlock()
		fmt.Printf("Estacionamiento lleno. Carro %d en espera.\n", vehicle.ID)

		ps.waitForSpot(vehicle)
	}
}

func (ps *ParkingService) waitForSpot(vehicle domain.Car) {
	for {
		ps.spotAvailable.L.Lock()
		spotIndex := ps.findAvailableSpot()
		ps.spotAvailable.L.Unlock()

		if spotIndex != -1 {
			ps.queueMutex.Lock()
			ps.removeFromQueue(vehicle)
			ps.queueMutex.Unlock()

			ps.assignSpotAndEnter(vehicle)
			return
		}

		ps.spotAvailable.L.Lock()
		ps.spotAvailable.Wait()
		ps.spotAvailable.L.Unlock()

		ps.queueMutex.Lock()
		inQueue := ps.vehicleInQueue(vehicle)
		ps.queueMutex.Unlock()

		if !inQueue {
			return
		}
	}
}

func (ps *ParkingService) handleEntryTimeout(vehicle domain.Car) {
	timeout := time.After(10 * time.Second)
	select {
	case <-ps.tryEnterWithTimeout(vehicle):
	case <-timeout:
		ps.spotAvailable.L.Lock()
		ps.queueMutex.Lock()
		ps.removeFromQueue(vehicle)
		ps.queueMutex.Unlock()

		vehicle.State = "Timeout"
		ps.UpdateChannel <- UpdateInfo{Car: vehicle, Entering: false, Spot: -1}
		ps.spotAvailable.L.Unlock()

		fmt.Printf("Carro %d timeout después de 10 segundos. Saliendo.\n", vehicle.ID)
	}
}

func (ps *ParkingService) tryEnterWithTimeout(vehicle domain.Car) chan struct{} {
	entered := make(chan struct{})
	go func() {
		ps.spotAvailable.L.Lock()
		defer ps.spotAvailable.L.Unlock()

		if !ps.vehicleInQueue(vehicle) {
			return
		}

		if spotIndex := ps.findAvailableSpot(); spotIndex != -1 {
			ps.assignSpotAndEnter(vehicle)
			close(entered)
		}
	}()
	return entered
}

func (ps *ParkingService) assignSpotAndEnter(vehicle domain.Car) {
	vehicle.Spot = ps.findAvailableSpot()
	if vehicle.Spot == -1 {
		fmt.Printf("Error: No se pudo asignar un lugar a %d\n", vehicle.ID)
		return
	}

	ps.Parking.Vehicles++
	vehicle.State = "Estacionamiento"
	ps.UpdateChannel <- UpdateInfo{Car: vehicle, Entering: true, Spot: vehicle.Spot}
	go ps.scheduleExit(vehicle)
}

func (ps *ParkingService) scheduleExit(vehicle domain.Car) {
	sleepDuration := time.Duration(rand.Intn(10)) * time.Second
	time.Sleep(sleepDuration)
	ps.ExitChannel <- vehicle
}

func (ps *ParkingService) ExitParking(vehicle domain.Car) {
	ps.spotAvailable.L.Lock()
	defer ps.spotAvailable.L.Unlock()

	if vehicle.Spot < 0 || vehicle.Spot >= ps.Parking.Capacity {
		fmt.Printf("El vehículo %d no puede salir: Lugar no válido.\n", vehicle.ID)
		return
	}

	ps.Parking.Vehicles--
	vehicle.State = "Exiting"
	ps.UpdateChannel <- UpdateInfo{Car: vehicle, Entering: false, Spot: -1}
	ps.freeSpot(vehicle.Spot)
	ps.spotAvailable.Signal()
}

func (ps *ParkingService) removeFromQueue(vehicle domain.Car) {
	for i, v := range ps.entryQueue {
		if v.ID == vehicle.ID {
			ps.entryQueue = append(ps.entryQueue[:i], ps.entryQueue[i+1:]...)
			return
		}
	}
}

func (ps *ParkingService) vehicleInQueue(vehicle domain.Car) bool {
	for _, v := range ps.entryQueue {
		if v.ID == vehicle.ID {
			return true
		}
	}
	return false
}

func (ps *ParkingService) findAvailableSpot() int {
	for i := 0; i < ps.Parking.Capacity; i++ {
		if ps.Parking.Spots[i] == 0 {
			ps.Parking.Spots[i] = 1
			return i
		}
	}
	return -1
}

func (ps *ParkingService) freeSpot(spotIndex int) {
	if spotIndex >= 0 && spotIndex < ps.Parking.Capacity {
		ps.Parking.Spots[spotIndex] = 0
	}
}
