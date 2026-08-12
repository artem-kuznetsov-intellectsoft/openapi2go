package generated

// Car is generated from components.schemas.Car.
type Car struct {
	LicensePlate    string `json:"license_plate"`
	SeatingCapacity *int32 `json:"seating_capacity,omitempty"`
	VehicleType     string `json:"vehicle_type"`
}
