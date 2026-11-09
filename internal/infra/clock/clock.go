// Package clock provides the wall clock used by the services.
package clock

import (
	"fmt"
	"time"
	_ "time/tzdata" // Windows has no system zone database.
)

const planLocation = "Europe/Vilnius"

// Clock reports the current time in the zone plan weeks are counted in.
type Clock struct {
	location *time.Location
}

func New() (*Clock, error) {
	location, err := time.LoadLocation(planLocation)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", planLocation, err)
	}
	return &Clock{location: location}, nil
}

func (c *Clock) Now() time.Time {
	return time.Now().In(c.location)
}
