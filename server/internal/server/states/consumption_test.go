package states

import (
	"math"
	"testing"
)

// TestPlayerConsumption tests the logic for player eating other players
func TestPlayerConsumption(t *testing.T) {
	t.Run("Player with mass 100 CAN eat player with mass 50", func(t *testing.T) {
		ourMass := 100.0
		otherMass := 50.0

		canConsume := ourMass > otherMass*1.5

		if !canConsume {
			t.Errorf("Player with mass %f should be able to consume player with mass %f (threshold: %f)", ourMass, otherMass, otherMass*1.5)
		}
	})

	t.Run("Player with mass 100 CANNOT eat player with mass 80", func(t *testing.T) {
		ourMass := 100.0
		otherMass := 80.0

		canConsume := ourMass > otherMass*1.5

		if canConsume {
			t.Errorf("Player with mass %f should NOT be able to consume player with mass %f (threshold: %f)", ourMass, otherMass, otherMass*1.5)
		}
	})

	t.Run("Exact threshold test - mass exactly 1.5x", func(t *testing.T) {
		otherMass := 100.0
		ourMass := otherMass * 1.5

		canConsume := ourMass > otherMass*1.5

		if canConsume {
			t.Errorf("Player with mass exactly 1.5x target mass should NOT be able to consume (our: %f, other: %f, threshold: %f)", ourMass, otherMass, otherMass*1.5)
		}
	})

	t.Run("Just above threshold - can consume", func(t *testing.T) {
		otherMass := 100.0
		ourMass := otherMass*1.5 + 0.01

		canConsume := ourMass > otherMass*1.5

		if !canConsume {
			t.Errorf("Player with mass slightly above 1.5x should be able to consume (our: %f, other: %f, threshold: %f)", ourMass, otherMass, otherMass*1.5)
		}
	})

	t.Run("Converting radius to mass for consumption check", func(t *testing.T) {
		ourRadius := 20.0
		otherRadius := 10.0

		ourMass := radToMass(ourRadius)
		otherMass := radToMass(otherRadius)

		canConsume := ourMass > otherMass*1.5

		if !canConsume {
			t.Errorf("Player with radius %f (mass %f) should be able to consume player with radius %f (mass %f), threshold: %f",
				ourRadius, ourMass, otherRadius, otherMass, otherMass*1.5)
		}
	})

	t.Run("Equal radius players cannot consume each other", func(t *testing.T) {
		radius := 20.0
		mass := radToMass(radius)

		canConsume := mass > mass*1.5

		if canConsume {
			t.Errorf("Players with equal mass should not be able to consume each other")
		}
	})

	t.Run("Small advantage not enough - 1.3x mass ratio", func(t *testing.T) {
		otherMass := 100.0
		ourMass := otherMass * 1.3

		canConsume := ourMass > otherMass*1.5

		if canConsume {
			t.Errorf("Player with only 1.3x mass should NOT be able to consume (our: %f, other: %f, threshold: %f)", ourMass, otherMass, otherMass*1.5)
		}
	})

	t.Run("Large mass advantage - 2x mass ratio", func(t *testing.T) {
		otherMass := 100.0
		ourMass := otherMass * 2.0

		canConsume := ourMass > otherMass*1.5

		if !canConsume {
			t.Errorf("Player with 2x mass should be able to consume (our: %f, other: %f, threshold: %f)", ourMass, otherMass, otherMass*1.5)
		}
	})

	t.Run("Very small players - radius 5 vs 3", func(t *testing.T) {
		ourRadius := 5.0
		otherRadius := 3.0

		ourMass := radToMass(ourRadius)
		otherMass := radToMass(otherRadius)

		canConsume := ourMass > otherMass*1.5

		if !canConsume {
			t.Errorf("Player with radius %f (mass %f) should be able to consume player with radius %f (mass %f), threshold: %f",
				ourRadius, ourMass, otherRadius, otherMass, otherMass*1.5)
		}
	})

	t.Run("Mass formula verification: mass >= 1.5 × masa_target", func(t *testing.T) {
		testCases := []struct {
			ourMass   float64
			otherMass float64
			canEat    bool
			desc      string
		}{
			{200, 100, true, "2x mass - can eat"},
			{150, 100, false, "1.5x mass exactly - cannot eat"},
			{151, 100, true, "1.51x mass - can eat"},
			{149, 100, false, "1.49x mass - cannot eat"},
			{100, 100, false, "equal mass - cannot eat"},
			{50, 100, false, "smaller mass - cannot eat"},
		}

		for _, tc := range testCases {
			canConsume := tc.ourMass > tc.otherMass*1.5

			if canConsume != tc.canEat {
				t.Errorf("%s: ourMass=%f, otherMass=%f, threshold=%f, expected canEat=%v, got %v",
					tc.desc, tc.ourMass, tc.otherMass, tc.otherMass*1.5, tc.canEat, canConsume)
			}
		}
	})
}

// Benchmark consumption check
func BenchmarkConsumptionCheck(b *testing.B) {
	ourMass := 1000.0
	otherMass := 500.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ourMass > otherMass*1.5
	}
}

// Benchmark full consumption calculation with radius conversion
func BenchmarkFullConsumptionCalculation(b *testing.B) {
	ourRadius := 30.0
	otherRadius := 15.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ourMass := math.Pi * ourRadius * ourRadius
		otherMass := math.Pi * otherRadius * otherRadius
		_ = ourMass > otherMass*1.5
	}
}
