package models

// DeliveryStatus is where a won opportunity's work has reached. It exists only
// once the sale is won; the opportunity is the project, so there is no separate
// project record to carry it.
type DeliveryStatus string

const (
	DeliveryPlanning   DeliveryStatus = "planning"
	DeliveryInProgress DeliveryStatus = "in-progress"
	DeliveryOnHold     DeliveryStatus = "on-hold"
	DeliveryCompleted  DeliveryStatus = "completed"
	DeliveryCancelled  DeliveryStatus = "cancelled"
)
