package main

import (
	"fmt"
	application "simulador/App"
	infrastructure "simulador/Infrestructure"
	ui "simulador/UI"
	"time"
)

func main() {
	numVehicles := 100
	timeout := 10 * time.Second
	parkingLotService := application.NewParkingLotService(20, timeout)
	parkingUI := ui.NewParkingUI(parkingLotService)
	parkingLotService.RegisterObserver(parkingUI)

	if parkingLotService == nil {
		fmt.Println("Error: parkingService es nil")
		return
	}

	go infrastructure.StartParkingControl(parkingLotService, numVehicles)
	ui.StartUI(parkingLotService)
}
