package main

import "math"

func planosNaArea(semente int64, area AreaXZ) []PlanoTerritorial {
	if area.MinX > area.MaxX || area.MinZ > area.MaxZ {
		return nil
	}

	// A borda máxima fica fora da área, exceto quando a área é um ponto.
	fimX, fimZ := area.MaxX, area.MaxZ
	if area.MaxX > area.MinX {
		fimX = math.Nextafter32(area.MaxX, float32(math.Inf(-1)))
	}
	if area.MaxZ > area.MinZ {
		fimZ = math.Nextafter32(area.MaxZ, float32(math.Inf(-1)))
	}

	inicio := coordenadaPlano(area.MinX, area.MinZ)
	fim := coordenadaPlano(fimX, fimZ)

	var planos []PlanoTerritorial
	for x := inicio.X; x <= fim.X; x++ {
		for z := inicio.Z; z <= fim.Z; z++ {
			planos = append(planos, gerarPlanoTerritorial(
				semente, CoordPlanoTerritorial{X: x, Z: z},
			))
		}
	}
	return planos
}
