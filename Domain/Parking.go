package Domain

import "sync"

type ParkingLot struct {
	Capacity int
	Vehicles int
	Spots    []int
	Mutex    sync.Mutex
}

func NewParkingLot(capacity int) *ParkingLot {
	return &ParkingLot{
		Capacity: capacity,
		Spots:    make([]int, capacity),
	}
}

func (p *ParkingLot) ParkCar(car *Car, spot int) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.Vehicles++
	p.Spots[spot] = car.ID
	car.Spot = spot
}

func (p *ParkingLot) RemoveCar(car *Car) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if car.Spot != -1 {
		p.Vehicles--
		p.Spots[car.Spot] = 0
		car.Spot = -1
	}
}

func (p *ParkingLot) IsCarParked(car *Car) bool {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	return car.Spot != -1 && p.Spots[car.Spot] == car.ID
}

func (p *ParkingLot) FindAvailableSpot() int {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	for i := 0; i < p.Capacity; i++ {
		if p.Spots[i] == 0 {
			return i
		}
	}
	return -1
}

func (p *ParkingLot) IsFull() bool {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	return p.Vehicles >= p.Capacity
}
