package generated

// Bicycle is generated from components.schemas.Bicycle.
type Bicycle struct {
	FrameSize   string `json:"frame_size,omitempty"`
	HasBasket   bool   `json:"has_basket"`
	VehicleType string `json:"vehicle_type"`
}

// Car is generated from components.schemas.Car.
type Car struct {
	LicensePlate    string `json:"license_plate"`
	SeatingCapacity int32  `json:"seating_capacity,omitempty"`
	VehicleType     string `json:"vehicle_type"`
}

// Vehicle is generated from components.schemas.Vehicle.
type Vehicle = Discriminated[Car, Bicycle]

// VehicleUnion is generated from components.schemas.VehicleUnion.
type VehicleUnion struct {
	BackupVehicles []Vehicle `json:"backup_vehicles,omitempty"`
	MainVehicle    Vehicle   `json:"main_vehicle"`
}
