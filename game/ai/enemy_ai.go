// EnemyAI package - separate from main
package ai

import "math/rand"

type EnemyAI struct{ moveSpeed float64 }

func NewEnemyAI() *EnemyAI { return &EnemyAI{moveSpeed: 50.0} }

func (ai *EnemyAI) Update(interface{}, interface{}, float64) { }
