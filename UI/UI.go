package ui

import (
	"fmt"
	"math"
	application "simulador/App"
	domain "simulador/Domain"
	"time"

	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	gridSize     = 20
	cellSize     = 60
	gridRows     = 4
	gridCols     = 5
	doorWidth    = 100
	doorHeight   = 60
	marginTop    = 80
	marginLeft   = 100
	laneWidth    = 40
	cornerRadius = 5
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
	carID        string
}

type ParkingUI struct {
	spots      [gridSize]ParkingSpot
	entryDoor  *fyne.Container
	container  *fyne.Container
	service    *application.ParkingService
	driveLanes []PathPoint
	background *canvas.Rectangle
	doorMutex  sync.Mutex
}

func createCarWithContainer(filePath string) (*fyne.Container, *canvas.Image, error) {
	carImg := canvas.NewImageFromFile(filePath)
	carImg.FillMode = canvas.ImageFillOriginal
	carImg.Resize(fyne.NewSize(40, 40))

	carContainer := container.NewWithoutLayout()
	carContainer.Add(carImg)
	carImg.Move(fyne.NewPos(0, 0))

	return carContainer, carImg, nil
}

func createGradientDoor(isEntry bool) *fyne.Container {
	mainRect := canvas.NewRectangle(theme.PrimaryColor())
	if !isEntry {
		mainRect.FillColor = theme.ErrorColor()
	}

	border := canvas.NewRectangle(theme.ShadowColor())
	border.StrokeWidth = 2
	border.StrokeColor = theme.ForegroundColor()

	label := canvas.NewText(map[bool]string{true: "ENTRADA"}[isEntry], theme.ForegroundColor())
	label.TextSize = 12
	label.TextStyle.Bold = true

	door := container.NewWithoutLayout()
	door.Add(mainRect)
	door.Add(border)
	door.Add(label)

	mainRect.Resize(fyne.NewSize(doorWidth, doorHeight))
	border.Resize(fyne.NewSize(doorWidth, doorHeight))
	label.Move(fyne.NewPos(doorWidth/4, doorHeight/3))

	return door
}

func NewParkingUI(service *application.ParkingService) *ParkingUI {
	ui := &ParkingUI{
		service:    service,
		container:  container.NewWithoutLayout(),
		driveLanes: make([]PathPoint, 0),
	}
	ui.background = canvas.NewRectangle(theme.BackgroundColor())
	ui.container.Add(ui.background)
	ui.createDriveLanes()
	ui.entryDoor = createGradientDoor(true)
	entryPos := fyne.NewPos(float32(marginLeft/2), float32(marginTop+cellSize*2))
	ui.entryDoor.Move(entryPos)

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
	ui.drawDriveLanes()

	ui.container.Add(ui.entryDoor)

	return ui
}

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

func (ui *ParkingUI) animateAlongPath(carContainer *fyne.Container, carImg *canvas.Image, path []PathPoint, done chan<- bool) {
	const stepsPerSegment = 10

	go func() {
		for i := 0; i < len(path)-1; i++ {
			start := path[i]
			end := path[i+1]

			dx := (end.x - start.x) / stepsPerSegment
			dy := (end.y - start.y) / stepsPerSegment
			direction := "horizontal"
			if math.Abs(float64(end.y-start.y)) > math.Abs(float64(end.x-start.x)) {
				direction = "vertical"
			}
			if direction == "vertical" {
				carImg.Resize(fyne.NewSize(40, 40))
			} else {
				carImg.Resize(fyne.NewSize(40, 40))
			}

			for step := 0; step < stepsPerSegment; step++ {
				pos := fyne.NewPos(
					start.x+dx*float32(step),
					start.y+dy*float32(step),
				)
				carContainer.Move(pos)
				carContainer.Refresh()
				time.Sleep(100 * time.Millisecond)
			}
		}
		done <- true
	}()
}
func (ui *ParkingUI) updateDisplay(car domain.Car, entering bool, spotIndex int) {
	ui.doorMutex.Lock()
	defer ui.doorMutex.Unlock()
	if entering {
		if spotIndex != -1 {
			carContainer, carImg, err := createCarWithContainer("assets/car.png")
			if err != nil {
				fmt.Printf("Error Cargando  Imagen: %v\n", err)
				return
			}

			ui.spots[spotIndex].carImage = carImg
			ui.spots[spotIndex].carContainer = carContainer
			ui.spots[spotIndex].occupied = true
			ui.spots[spotIndex].carID = string(car.ID)
			ui.entryDoor.Hide()

			path := ui.calculatePath(ui.entryDoor.Position(), ui.spots[spotIndex].position)
			done := make(chan bool)
			ui.container.Add(carContainer)
			ui.animateAlongPath(carContainer, carImg, path, done)

			go func(spotIndex int) {
				<-done
				ui.spots[spotIndex].carContainer.Move(ui.spots[spotIndex].position)
				ui.spots[spotIndex].carContainer.Refresh()
				ui.entryDoor.Show()
				ui.entryDoor.Refresh()
				time.Sleep(5 * time.Second)

			}(spotIndex)
		}
	} else {
		ui.entryDoor.Hide()

		for i := range ui.spots {
			if ui.spots[i].occupied && ui.spots[i].carID == string(car.ID) {
				path := ui.calculatePath(ui.spots[i].position, ui.entryDoor.Position())
				done := make(chan bool)
				ui.animateAlongPath(ui.spots[i].carContainer, ui.spots[i].carImage, path, done)

				go func(i int) {
					<-done
					ui.spots[i].carContainer.Hide()
					ui.spots[i].occupied = false
					ui.spots[i].carID = ""
					ui.entryDoor.Show()
					ui.entryDoor.Refresh()

				}(i)
				break
			}
		}
	}

	ui.container.Refresh()

}

func StartUI(service *application.ParkingService) {
	a := app.New()
	w := a.NewWindow("Sistema de Estacionamiento")
	w.Resize(fyne.NewSize(1000, 700))

	parkingUI := NewParkingUI(service)

	header := widget.NewLabel("Sistema de Estacionamiento")
	header.TextStyle = fyne.TextStyle{Bold: true}
	header.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabel("Estado: 0/20 espacios ocupados")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	go func() {
		for update := range service.UpdateChannel {
			if update.Entering {
				parkingUI.updateDisplay(update.Car, true, update.Spot)
			} else {
				parkingUI.updateDisplay(update.Car, false, -1)
			}
			statusLabel.SetText(fmt.Sprintf("Estado: %d/%d espacios ocupados",
				service.Parking.Vehicles, service.Parking.Capacity))
		}
	}()

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		statusLabel,
		widget.NewSeparator(),
		parkingUI.container,
	)

	w.SetContent(content)
	w.ShowAndRun()
}
