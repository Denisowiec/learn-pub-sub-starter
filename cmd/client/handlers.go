package main

import (
	"fmt"

	"github.com/Denisowiec/learn-pub-sub-starter/internal/gamelogic"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/routing"
)

func HandlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
	}
}

func HandlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(am gamelogic.ArmyMove) {
		defer fmt.Print("> ")

		gs.HandleMove(am)
	}
}
