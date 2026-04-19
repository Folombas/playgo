package ecs

type World struct{}

type Entity struct{}

type Component interface{}

func NewWorld() *World { return &World{} }

func (w *World) NewEntity() *Entity { return &Entity{} }

func (e *Entity) AddComponent(c Component) {}

func (e *Entity) GetComponent(t interface{}) interface{} { return nil }

func (w *World) Clear() {}

func (w *World) AddEntity(e *Entity) {}

func (w *World) UpdateSystems(dt float64) {}

func (w *World) RenderSystems() {}
