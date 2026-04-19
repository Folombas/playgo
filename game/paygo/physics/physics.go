package physics

import "math"

type Vec2 struct{ X, Y float64 }

type TransformComponent struct{ Position, Size Vec2 }
type BodyComponent struct{ Body interface{} }

type Body interface {
	Position() Vec2
	SetPosition(Vec2)
	Velocity() Vec2
	SetVelocity(Vec2)
	Mass() float64
}
type BodyStatic struct{}
func (b *BodyStatic) Position() Vec2     { return Vec2{} }
func (b *BodyStatic) SetPosition(Vec2)   {}
func (b *BodyStatic) Velocity() Vec2     { return Vec2{} }
func (b *BodyStatic) SetVelocity(Vec2)   {}
func (b *BodyStatic) Mass() float64      { return 0 }
type BodyDynamic struct{}
func (b *BodyDynamic) Position() Vec2     { return Vec2{} }
func (b *BodyDynamic) SetPosition(Vec2)   {}
func (b *BodyDynamic) Velocity() Vec2     { return Vec2{} }
func (b *BodyDynamic) SetVelocity(Vec2)   {}
func (b *BodyDynamic) Mass() float64      { return 1 }
func NewBodyStatic(interface{}) Body      { return &BodyStatic{} }
func NewBodyDynamic(interface{}, float64) Body { return &BodyDynamic{} }
