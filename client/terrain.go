package main

import (
	"math"

	fastnoise "github.com/Auburn/FastNoiseLite/Go"
)

func densidadeNoPonto(ruido *fastnoise.State[float32], x, y, z float32) float32 {
	altura := ruido.GetNoise2D(x, z) * amplitudeTerrenoAtual
	return altura - y
}

func novoRuido(semente int) *fastnoise.State[float32] {
	ruido := fastnoise.New[float32]()
	ruido.Seed = semente
	ruido.Frequency = frequenciaRuidoAtual
	ruido.NoiseType(fastnoise.Perlin)
	return ruido
}

func densidadePonto(campo *[17][9][17]float32, x, y, z int) float32 {
	return (*campo)[x+8][y+4][z+8]
}

func cortarEsfera(campo *[17][9][17]float32, cx, cy, cz, raio float32, aoAlterar func(x, y, z int, densidade float32)) bool {
	alterado := false

	for x := -8; x <= 8; x++ {
		for y := -4; y <= 4; y++ {
			for z := -8; z <= 8; z++ {
				dx := float32(x) - cx
				dy := float32(y) - cy
				dz := float32(z) - cz

				distancia := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
				densidadeCorte := distancia - raio

				ix, iy, iz := x+8, y+4, z+8
				if densidadeCorte < (*campo)[ix][iy][iz] {
					(*campo)[ix][iy][iz] = densidadeCorte
					aoAlterar(x, y, z, densidadeCorte)
					alterado = true
				}
			}
		}
	}

	return alterado
}
