package example

import (
	"math"
	"time"
)

type Dynamics struct {
	Position            Vec2
	Velocity            Vec2
	MaxSpeed            float64
	Acceleration        Vec2
	Angle               float64
	AngularVelocity     float64
	AngularAcceleration float64
	MaxAngularSpeed     float64
}

func (d *Dynamics) Direction() Vec2 {
	return Vec2{
		math.Sin(d.Angle),
		-math.Cos(d.Angle),
	}.Normalize()
}

type Behavior struct {
	Kind string
}

type Creature struct {
	Dynamics *Dynamics
	Behavior *Behavior
}

type (
	Vec2 [2]float64
)

func (v1 Vec2) Add(v2 Vec2) Vec2 {
	return Vec2{v1[0] + v2[0], v1[1] + v2[1]}
}

func (v1 Vec2) Mul(c float64) Vec2 {
	return Vec2{v1[0] * c, v1[1] * c}
}

func (v1 Vec2) Len() float64 {
	return float64(math.Hypot(float64(v1[0]), float64(v1[1])))
}

func (v1 Vec2) Normalize() Vec2 {
	l := 1.0 / v1.Len()
	return Vec2{v1[0] * l, v1[1] * l}
}

type DynamicsSystem struct {
	em *EM
}

func (c *DynamicsSystem) Update(dt time.Duration) error {
	secs := dt.Seconds()
	return c.em.EachDynamics(func(d *Dynamics) error {
		d.Velocity = d.Velocity.Add(d.Acceleration.Mul(secs))
		if d.Velocity.Len() > d.MaxSpeed {
			d.Velocity = d.Direction().Mul(d.MaxSpeed)
		}
		d.Position = d.Position.Add(d.Velocity.Mul(secs))

		d.AngularVelocity += d.AngularAcceleration * secs
		if d.AngularVelocity > d.MaxAngularSpeed {
			d.AngularVelocity = d.MaxAngularSpeed
		} else if d.AngularVelocity < -d.MaxAngularSpeed {
			d.AngularVelocity = -d.MaxAngularSpeed
		}
		d.Angle += d.AngularVelocity * secs
		return nil
	})
}
