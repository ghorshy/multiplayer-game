package states

import (
	"math"
	"server/internal/server/objects"
	"testing"
)

// TestPhysicsCalculations tests physics formulas for mass and radius
func TestPhysicsCalculations(t *testing.T) {
	t.Run("radToMass calculation", func(t *testing.T) {
		// masa = π × promień²
		tests := []struct {
			name     string
			radius   float64
			expected float64
		}{
			{"radius 10", 10.0, math.Pi * 100},
			{"radius 20", 20.0, math.Pi * 400},
			{"radius 5", 5.0, math.Pi * 25},
			{"radius 1", 1.0, math.Pi},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := radToMass(tt.radius)
				if math.Abs(result-tt.expected) > 0.0001 {
					t.Errorf("radToMass(%f) = %f, expected %f", tt.radius, result, tt.expected)
				}
			})
		}
	})

	t.Run("massToRad calculation", func(t *testing.T) {
		// promień = sqrt(masa / π)
		tests := []struct {
			name     string
			mass     float64
			expected float64
		}{
			{"mass π*100", math.Pi * 100, 10.0},
			{"mass π*400", math.Pi * 400, 20.0},
			{"mass π*25", math.Pi * 25, 5.0},
			{"mass π", math.Pi, 1.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := massToRad(tt.mass)
				if math.Abs(result-tt.expected) > 0.0001 {
					t.Errorf("massToRad(%f) = %f, expected %f", tt.mass, result, tt.expected)
				}
			})
		}
	})

	t.Run("radToMass and massToRad are inverse functions", func(t *testing.T) {
		radii := []float64{5.0, 10.0, 15.0, 20.0, 25.0, 50.0, 100.0}

		for _, radius := range radii {
			mass := radToMass(radius)
			backToRadius := massToRad(mass)

			if math.Abs(backToRadius-radius) > 0.0001 {
				t.Errorf("Round trip failed: radius %f -> mass %f -> radius %f", radius, mass, backToRadius)
			}
		}
	})

	t.Run("nextRadius after consumption", func(t *testing.T) {
		// Symulacja: gracz z promieniem 20 zjada spore z promieniem 5
		game := &InGame{
			player: &objects.Player{
				Radius: 20.0,
			},
		}

		sporeMass := radToMass(5.0) // masa spore
		newRadius := game.nextRadius(sporeMass)

		// Oczekiwana masa: masa_gracza + masa_spore
		expectedMass := radToMass(20.0) + radToMass(5.0)
		expectedRadius := massToRad(expectedMass)

		if math.Abs(newRadius-expectedRadius) > 0.0001 {
			t.Errorf("nextRadius with sporeMass %f = %f, expected %f", sporeMass, newRadius, expectedRadius)
		}
	})

	t.Run("newRadius is greater than oldRadius after gaining mass", func(t *testing.T) {
		game := &InGame{
			player: &objects.Player{
				Radius: 20.0,
			},
		}

		massDiff := 100.0
		newRadius := game.nextRadius(massDiff)

		if newRadius <= game.player.Radius {
			t.Errorf("After gaining mass, newRadius (%f) should be > oldRadius (%f)", newRadius, game.player.Radius)
		}
	})
}
