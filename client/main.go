package main

import rl "github.com/gen2brain/raylib-go/raylib"

func main() {
	rl.InitWindow(1920, 720, "Game of Survive")
	rl.DisableCursor()

	gameLoop()

	rl.EnableCursor()
	rl.CloseWindow()
}
