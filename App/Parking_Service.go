package parking

import (
	"fmt"
	domain "simulador/Domain"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type Observer interface {
	Update(info UpdateInfo)
}

type UpdateInfo struct {
	Car       *domain.Car
	Entering  bool
	Spot      int
	EventType string
}

type ParkingLotService struct {
	ParkingLot      *domain.ParkingLot
	entryChannel    chan *domain.Car
	exitChannel     chan *domain.Car
	entryQueue      []*domain.Car
	spotAvailable   *sync.Cond
	queueMutex      sync.Mutex
	UpdateChannel   chan UpdateInfo
	timeoutDuration time.Duration
	observers       []Observer
	observerMutex   sync.Mutex
}

func NewParkingLotService(capacity int, timeout time.Duration) *ParkingLotService {
	spots := make([]int, capacity)
	parkingLot := &domain.ParkingLot{Capacity: capacity, Spots: spots}
	ps := &ParkingLotService{
		ParkingLot:      parkingLot,
		entryChannel:    make(chan *domain.Car),
		exitChannel:     make(chan *domain.Car),
		entryQueue:      make([]*domain.Car, 20),
		timeoutDuration: timeout,
		UpdateChannel:   make(chan UpdateInfo),
	}
	ps.spotAvailable = sync.NewCond(&sync.Mutex{})
	return ps
}

func (ps *ParkingLotService) RegisterObserver(o Observer) {
	ps.observerMutex.Lock()
	defer ps.observerMutex.Unlock()
	ps.observers = append(ps.observers, o)
}

func (ps *ParkingLotService) RemoveObserver(o Observer) {
	ps.observerMutex.Lock()
	defer ps.observerMutex.Unlock()
	for i, observer := range ps.observers {
		if observer == o {
			ps.observers = append(ps.observers[:i], ps.observers[i+1:]...)
			break
		}
	}
}

func (ps *ParkingLotService) notifyObservers(info UpdateInfo) {
	fmt.Printf("Notificando: Car %d - Estado: %s, Spot: %d, Tipo de Evento: %s\n", info.Car.ID, info.Spot, info.Car.Spot, info.EventType)

	ps.observerMutex.Lock()
	defer ps.observerMutex.Unlock()
	for _, observer := range ps.observers {
		go observer.Update(info)
	}
}

func (ps *ParkingLotService) EnterParking(car *domain.Car) {
	fmt.Printf("Car %d intentando  entrar\n", car.ID)
	ps.spotAvailable.L.Lock()
	defer ps.spotAvailable.L.Unlock()

	if ps.ParkingLot.IsFull() {
		ps.queueMutex.Lock()
		ps.entryQueue = append(ps.entryQueue, car)
		ps.queueMutex.Unlock()
		fmt.Printf("Parking lleno. Carro %d Esperando.\n", car.ID)
		ps.notifyObservers(UpdateInfo{Car: car, Entering: true, Spot: -1, EventType: "CarWaiting"})
		go ps.handleEntryTimeout(car)
		ps.waitForSpot(car)
	} else {
		ps.assignSpotAndEnter(car)
	}
}

func (ps *ParkingLotService) assignSpotAndEnter(car *domain.Car) {
	spot := ps.ParkingLot.FindAvailableSpot()
	if spot == -1 {
		fmt.Printf("Error: No hay lugar que se pueda asignar  %d\n", car.ID)
		return
	}
	ps.ParkingLot.ParkCar(car, spot)
	car.State = "Parked"
	fmt.Printf("Carro %d Parkeado en el lugar %d\n", car.ID, spot)
	ps.notifyObservers(UpdateInfo{Car: car, Entering: true, Spot: spot, EventType: "CarParked"})
	go ps.scheduleExit(car)
}

func (ps *ParkingLotService) ExitParking(car *domain.Car) {
	ps.spotAvailable.L.Lock()
	defer ps.spotAvailable.L.Unlock()

	if !ps.ParkingLot.IsCarParked(car) {
		fmt.Printf("Vehicle %d no puede salir: Lugar Invalido.\n", car.ID)
		return
	}

	ps.ParkingLot.RemoveCar(car)
	car.State = "Exiting"
	ps.notifyObservers(UpdateInfo{Car: car, Entering: false, Spot: -1, EventType: "CarExiting"})
	ps.spotAvailable.Broadcast()
}

func (ps *ParkingLotService) GetEntryChannel() chan<- *domain.Car {
	fmt.Printf("Carro %d entrando al  parking \n")
	return ps.entryChannel
}

func (ps *ParkingLotService) GetExitChannel() chan *domain.Car {
	return ps.exitChannel
}

func (ps *ParkingLotService) scheduleExit(car *domain.Car) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	ps.exitChannel <- car
}

func (ps *ParkingLotService) waitForSpot(car *domain.Car) {
	fmt.Printf("Car %d Esperando por un lugar \n", car.ID)
	for {
		ps.spotAvailable.L.Lock()
		if !ps.isCarInQueue(car) {
			ps.spotAvailable.L.Unlock()
			return
		}
		if !ps.ParkingLot.IsFull() {
			ps.assignSpotAndEnter(car)
			ps.spotAvailable.L.Unlock()
			return
		}
		ps.spotAvailable.Wait()
		ps.spotAvailable.L.Unlock()
	}
}

func (ps *ParkingLotService) handleEntryTimeout(car *domain.Car) {
	timer := time.NewTimer(ps.timeoutDuration)
	select {
	case <-timer.C:
		ps.spotAvailable.L.Lock()
		ps.removeFromQueue(car)
		car.State = "Timeout"
		ps.notifyObservers(UpdateInfo{Car: car, Entering: false, Spot: -1, EventType: "CarTimeout"})
		ps.spotAvailable.L.Unlock()
		fmt.Printf("Car %d se acabo el tiempo de espera %v. saliendo.\n", car.ID, ps.timeoutDuration)

	case <-car.Cancel:
		timer.Stop()
		ps.spotAvailable.L.Lock()
		ps.removeFromQueue(car)
		car.State = "Cancelled"
		ps.notifyObservers(UpdateInfo{Car: car, Entering: false, Spot: -1, EventType: "CarCancelled"})
		ps.spotAvailable.L.Unlock()
		fmt.Printf("Car %d se cancela la espera . Saliendo.\n", car.ID)
	}
}

func (ps *ParkingLotService) removeFromQueue(car *domain.Car) {
	ps.queueMutex.Lock()
	defer ps.queueMutex.Unlock()
	for i, c := range ps.entryQueue {
		if c.ID == car.ID {
			ps.entryQueue = append(ps.entryQueue[:i], ps.entryQueue[i+1:]...)
			return
		}
	}
}

func (ps *ParkingLotService) isCarInQueue(car *domain.Car) bool {
	ps.queueMutex.Lock()
	defer ps.queueMutex.Unlock()
	for _, c := range ps.entryQueue {
		if c.ID == car.ID {
			return true
		}
	}
	return false
}

func NewParkingService(parkingSize int) *ParkingLotService {
	ps := &ParkingLotService{
		entryChannel: make(chan *domain.Car, parkingSize),
		exitChannel:  make(chan *domain.Car, parkingSize),
	}
	go ps.handleCarEntry()
	go ps.handleCarExit()
	return ps
}

func (ps *ParkingLotService) handleCarEntry() {
	for car := range ps.entryChannel {

		ps.EnterParking(car)

	}

}

func (ps *ParkingLotService) handleCarExit() {
	for car := range ps.exitChannel {
		ps.ExitParking(car)
	}
}
