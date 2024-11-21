package Domain

type Car struct {
	ID     int
	State  string
	Spot   int
	Image  string
	Cancel chan bool
}

func NewCar(id int, image string) *Car {
	return &Car{
		ID:     id,
		Image:  image,
		State:  "Waiting",
		Cancel: make(chan bool, 1),
	}
}

func (c *Car) GetId() int {
	return c.ID
}

func (c *Car) GetImage() string {
	return c.Image
}
