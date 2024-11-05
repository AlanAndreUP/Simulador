package main

import (
	application "simulador/App"
	infrastructure "simulador/Infrestructure"
	ui "simulador/UI"
)

func main() {
	numVehicles := 100

	parkingService := application.NewParkingService(20)

	done := make(chan bool)
	go infrastructure.StartParkingControl(parkingService, numVehicles)
	go infrastructure.GenerateVehicles(parkingService.EntryChannel, numVehicles)
	ui.StartUI(parkingService)

	<-done
}
