package main

import (
	"math"

	fastnoise "github.com/Auburn/FastNoiseLite/Go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	tamanhoRegiaoXZ = 16
	tamanhoRegiaoY  = 8
)

type CoordRegiaoTerreno struct {
	X, Y, Z int32
}

type CoordAmostraTerreno struct {
	X, Y, Z int
}

type RegiaoTerreno struct {
	Coord             CoordRegiaoTerreno
	Campo             [17][9][17]float32
	Modelo            rl.Model
	ArvoresCarregadas bool
	RochasCarregadas  bool
	GramaCarregada    bool
	BordaCarregada    bool
	MascaraEstrada    rl.Texture2D
}

func coordenadaRegiao(x, y, z float32) CoordRegiaoTerreno {
	return CoordRegiaoTerreno{
		X: int32(math.Floor((float64(x) + tamanhoRegiaoXZ/2) / tamanhoRegiaoXZ)),
		Y: int32(math.Floor((float64(y) + tamanhoRegiaoY/2) / tamanhoRegiaoY)),
		Z: int32(math.Floor((float64(z) + tamanhoRegiaoXZ/2) / tamanhoRegiaoXZ)),
	}
}

func preencherCampoRegiao(
	campo *[17][9][17]float32,
	coord CoordRegiaoTerreno,
	ruido *fastnoise.State[float32],
) {
	for ix := 0; ix <= tamanhoRegiaoXZ; ix++ {
		for iy := 0; iy <= tamanhoRegiaoY; iy++ {
			for iz := 0; iz <= tamanhoRegiaoXZ; iz++ {
				xGlobal := int64(coord.X)*tamanhoRegiaoXZ + int64(ix) - tamanhoRegiaoXZ/2
				yGlobal := int64(coord.Y)*tamanhoRegiaoY + int64(iy) - tamanhoRegiaoY/2
				zGlobal := int64(coord.Z)*tamanhoRegiaoXZ + int64(iz) - tamanhoRegiaoXZ/2

				(*campo)[ix][iy][iz] = densidadeNoPonto(
					ruido,
					float32(xGlobal),
					float32(yGlobal),
					float32(zGlobal),
				)
			}
		}
	}
}

func (regiao *RegiaoTerreno) posicaoMundo() rl.Vector3 {
	return rl.NewVector3(
		float32(regiao.Coord.X)*tamanhoRegiaoXZ,
		float32(regiao.Coord.Y)*tamanhoRegiaoY,
		float32(regiao.Coord.Z)*tamanhoRegiaoXZ,
	)
}

func (regiao *RegiaoTerreno) indicesDaAmostraGlobal(x, y, z int) (int, int, int, bool) {
	minX := int64(regiao.Coord.X)*tamanhoRegiaoXZ - tamanhoRegiaoXZ/2
	minY := int64(regiao.Coord.Y)*tamanhoRegiaoY - tamanhoRegiaoY/2
	minZ := int64(regiao.Coord.Z)*tamanhoRegiaoXZ - tamanhoRegiaoXZ/2

	gx, gy, gz := int64(x), int64(y), int64(z)
	if gx < minX || gx > minX+tamanhoRegiaoXZ ||
		gy < minY || gy > minY+tamanhoRegiaoY ||
		gz < minZ || gz > minZ+tamanhoRegiaoXZ {
		return 0, 0, 0, false
	}

	return int(gx - minX), int(gy - minY), int(gz - minZ), true
}

func (coord CoordRegiaoTerreno) areaXZ() AreaXZ {
	centroX := float32(coord.X) * tamanhoRegiaoXZ
	centroZ := float32(coord.Z) * tamanhoRegiaoXZ
	metade := float32(tamanhoRegiaoXZ) / 2

	return AreaXZ{
		MinX: centroX - metade,
		MaxX: centroX + metade,
		MinZ: centroZ - metade,
		MaxZ: centroZ + metade,
	}
}
