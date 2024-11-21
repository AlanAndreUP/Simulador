package ui

import (
	"fmt"
	"sync"
	"time"

	application "simulador/App"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	gridSize      = 20
	cellSize      = 60
	gridRows      = 4
	gridCols      = 5
	doorWidth     = 100
	doorHeight    = 60
	marginTop     = 80
	marginLeft    = 100
	laneWidth     = 40
	cornerRadius  = 5
	animationStep = 50 * time.Millisecond
)

type PathPoint struct {
	x, y float32
}

type ParkingSpot struct {
	rect         *canvas.Rectangle
	occupied     bool
	carImage     *canvas.Image
	carContainer *fyne.Container
	position     fyne.Position
	spotLabel    *canvas.Text
	carID        int
}

type ParkingUI struct {
	spots         [gridSize]ParkingSpot
	entryDoor     *fyne.Container
	container     *fyne.Container
	service       *application.ParkingLotService
	updateChannel chan application.UpdateInfo
	statusLabel   *widget.Label
	app           fyne.App
	window        fyne.Window
	driveLanes    []PathPoint
	background    *canvas.Rectangle
	spotMutex     sync.RWMutex
}

// Update implements parking.Observer.
func (ui *ParkingUI) Update(info application.UpdateInfo) {
	ui.updateChannel <- info
}

// createCarWithContainer creates a container with a car image
func createCarWithContainer(filePath string) (*fyne.Container, *canvas.Image, error) {
	carImg := canvas.NewImageFromFile(filePath)
	carImg.FillMode = canvas.ImageFillOriginal
	carImg.Resize(fyne.NewSize(40, 40))

	carContainer := container.NewWithoutLayout(carImg)
	carContainer.Resize(fyne.NewSize(5, 5))

	return carContainer, carImg, nil
}

// createGradientDoor creates a door with a gradient and label
func createGradientDoor(isEntry bool) *fyne.Container {
	doorColor := theme.PrimaryColor()
	if !isEntry {
		doorColor = theme.ErrorColor()
	}

	mainRect := canvas.NewRectangle(doorColor)
	border := canvas.NewRectangle(theme.ShadowColor())
	border.StrokeWidth = 2
	border.StrokeColor = theme.ForegroundColor()

	label := canvas.NewText(map[bool]string{true: "ENTRADA", false: "SALIDA"}[isEntry], theme.ForegroundColor())
	label.TextSize = 12
	label.TextStyle.Bold = true

	door := container.NewWithoutLayout(mainRect, border, label)

	mainRect.Resize(fyne.NewSize(doorWidth, doorHeight))
	border.Resize(fyne.NewSize(doorWidth, doorHeight))
	label.Move(fyne.NewPos(doorWidth/4, doorHeight/3))

	return door
}

// NewParkingUI initializes the parking lot UI
func NewParkingUI(service *application.ParkingLotService) *ParkingUI {
	a := app.New()
	w := a.NewWindow("Parking System")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(800, 600))
	ui := &ParkingUI{
		service:       service,
		window:        w,
		container:     container.NewWithoutLayout(),
		driveLanes:    make([]PathPoint, 0),
		updateChannel: make(chan application.UpdateInfo, 100), // Buffered channel
	}

	ui.initializeUI()
	return ui
}

// initializeUI sets up the user interface components
func (ui *ParkingUI) initializeUI() {
	ui.background = canvas.NewRectangle(theme.BackgroundColor())
	ui.container.Add(ui.background)

	ui.createDriveLanes()
	ui.setupParkingSpots()
	ui.drawDriveLanes()

	ui.entryDoor = createGradientDoor(true)
	entryPos := fyne.NewPos(float32(marginLeft/2), float32(marginTop+cellSize*2))
	ui.entryDoor.Move(entryPos)
	ui.container.Add(ui.entryDoor)
}

// setupParkingSpots initializes parking spots
func (ui *ParkingUI) setupParkingSpots() {
	for i := 0; i < gridSize; i++ {
		row := i / gridCols
		col := i % gridCols

		spot := &ui.spots[i]
		spot.rect = canvas.NewRectangle(theme.DisabledButtonColor())
		spot.rect.StrokeWidth = 1
		spot.rect.StrokeColor = theme.PrimaryColor()
		spot.rect.Resize(fyne.NewSize(cellSize-4, cellSize-4))

		spot.position = fyne.NewPos(
			float32(marginLeft+col*cellSize),
			float32(marginTop+row*cellSize),
		)
		spot.rect.Move(spot.position)

		spot.spotLabel = canvas.NewText(fmt.Sprintf("%d", i+1), theme.ForegroundColor())
		spot.spotLabel.TextSize = 12
		spot.spotLabel.Move(fyne.NewPos(
			spot.position.X+5,
			spot.position.Y+5,
		))

		ui.container.Add(spot.rect)
		ui.container.Add(spot.spotLabel)
	}
}

// createHeader creates the header for the parking system
func (ui *ParkingUI) createHeader() *fyne.Container {
	// Crear el label para el encabezado
	headerLabel := widget.NewLabel("Sistema de Estacionamiento")
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}
	headerLabel.Alignment = fyne.TextAlignCenter

	// Crear el label de estado para el encabezado
	ui.statusLabel = widget.NewLabel("Estado: 0/20 espacios ocupados")
	ui.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Crear un contenedor con el encabezado y el label de estado
	headerContainer := container.NewVBox(
		headerLabel,
		ui.statusLabel,
	)

	return headerContainer
}

// getOccupiedSpaces counts the number of occupied parking spots
func (ui *ParkingUI) getOccupiedSpaces() int {
	occupied := 0
	for i := 0; i < gridSize; i++ {
		if ui.spots[i].occupied {
			occupied++
		}
	}
	return occupied
}

// createDriveLanes creates points for driving lanes
func (ui *ParkingUI) createDriveLanes() {
	startX := marginLeft/2 + doorWidth
	endX := marginLeft + cellSize*gridCols
	laneY := marginTop + cellSize*2 + doorHeight/2
	for x := startX; x <= endX; x += 10 {
		ui.driveLanes = append(ui.driveLanes, PathPoint{float32(x), float32(laneY)})
	}
	for col := 0; col < gridCols; col++ {
		x := marginLeft + col*cellSize + cellSize/2
		for y := marginTop; y <= marginTop+cellSize*gridRows; y += 10 {
			ui.driveLanes = append(ui.driveLanes, PathPoint{float32(x), float32(y)})
		}
	}
}

// drawDriveLanes draws the driving lanes on the UI
func (ui *ParkingUI) drawDriveLanes() {
	laneColor := theme.DisabledButtonColor()
	mainLane := canvas.NewRectangle(laneColor)
	mainLane.Resize(fyne.NewSize(float32(gridCols*cellSize+doorWidth), laneWidth))
	mainLane.Move(fyne.NewPos(
		float32(marginLeft/2+doorWidth/2),
		float32(marginTop+cellSize*2+doorHeight/2-laneWidth/2),
	))
	ui.container.Add(mainLane)
	for col := 0; col < gridCols; col++ {
		vertLane := canvas.NewRectangle(laneColor)
		vertLane.Resize(fyne.NewSize(laneWidth, float32(gridRows*cellSize)))
		vertLane.Move(fyne.NewPos(
			float32(marginLeft+col*cellSize+cellSize/2-laneWidth/2),
			float32(marginTop),
		))
		ui.container.Add(vertLane)
	}
}

// calculatePath determines the path for car movement
func (ui *ParkingUI) calculatePath(from, to fyne.Position) []PathPoint {
	path := make([]PathPoint, 0)
	path = append(path, PathPoint{from.X, from.Y})
	mainLaneY := marginTop + cellSize*2 + doorHeight/2
	if from.Y < float32(mainLaneY) {
		path = append(path, PathPoint{from.X + cellSize/2, from.Y})
		path = append(path, PathPoint{from.X + cellSize/2, float32(mainLaneY)})
	} else {
		path = append(path, PathPoint{from.X, float32(mainLaneY)})
	}
	targetX := to.X
	if to.Y < float32(mainLaneY) {
		targetX += cellSize / 2
	}
	path = append(path, PathPoint{targetX, float32(mainLaneY)})
	if to.Y < float32(mainLaneY) {
		path = append(path, PathPoint{targetX, to.Y})
		path = append(path, PathPoint{to.X, to.Y})
	} else {
		path = append(path, PathPoint{to.X, to.Y})
	}
	return path
}

// StartUI launches the parking lot user interface
func StartUI(service *application.ParkingLotService) {
	parkingUI := NewParkingUI(service)
	if parkingUI == nil {
		return
	}

	// Registrar la UI como observador
	service.RegisterObserver(parkingUI)

	header := parkingUI.createHeader()
	content := container.NewVBox(
		header,
		parkingUI.container,
	)

	parkingUI.window.SetContent(content)
	go parkingUI.processUpdates()
	parkingUI.window.ShowAndRun()
}

// processUpdates handles car park updates concurrently
func (ui *ParkingUI) processUpdates() {
	for update := range ui.updateChannel {
		ui.safeUpdate(update)
	}
}

// safeUpdate ensures thread-safe updates to the UI
func (ui *ParkingUI) safeUpdate(info application.UpdateInfo) {
	ui.spotMutex.Lock()
	defer ui.spotMutex.Unlock()

	switch info.EventType {
	case "CarParked":
		ui.parkCar(info)
	case "CarExiting":
		ui.removeCar(info)
	}
}

// Ajustes en parkCar y removeCar para resolver problemas de contenedores
// parkCar handles car parking logic
func (ui *ParkingUI) parkCar(info application.UpdateInfo) {
	spot := ui.findAvailableSpot()
	if spot == nil {
		fmt.Println("No hay spots disponibles")
		return
	}

	spot.carID = info.Car.GetId()
	carContainer, carImage, _ := createCarWithContainer(info.Car.GetImage())
	spot.carContainer = carContainer
	spot.carImage = carImage

	// Actualiza el color del spot para indicar ocupación
	spot.rect.FillColor = theme.PrimaryColor()
	spot.occupied = true

	// Mueve el auto al lugar correcto
	carContainer.Move(spot.position)

	// Agrega el auto al contenedor principal
	ui.container.Add(carContainer)
	ui.window.Content().Refresh()

	// Actualiza el estado del sistema
	ui.updateStatusLabel()
	spot.carContainer.Show()
}

func (ui *ParkingUI) removeCar(info application.UpdateInfo) {
	for i := range ui.spots {
		spot := &ui.spots[i]
		if spot.occupied && spot.carID == info.Car.GetId() {
			spot.rect.FillColor = theme.DisabledButtonColor()
			spot.occupied = false
			if spot.carContainer != nil {
				spot.carContainer.Hide()
				ui.container.Remove(spot.carContainer)
				spot.carContainer = nil
			}
			break
		}
	}

	ui.updateStatusLabel()
	ui.container.Refresh()
}

// Siempre mantener statusLabel visible
func (ui *ParkingUI) updateStatusLabel() {
	occupiedSpaces := ui.getOccupiedSpaces()
	ui.statusLabel.SetText(fmt.Sprintf("Estado: %d/%d espacios ocupados", occupiedSpaces, gridSize))
	ui.window.Content().Refresh() // Asegurar refresco visual
}

// findAvailableSpot finds the first unoccupied parking spot
func (ui *ParkingUI) findAvailableSpot() *ParkingSpot {
	for i := range ui.spots {
		if !ui.spots[i].occupied {
			return &ui.spots[i]
		}
	}
	return nil
}
