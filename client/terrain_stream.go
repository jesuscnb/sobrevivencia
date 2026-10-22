package main

import (
	"fmt"
	"time"
)

const (
	raioCarregamento    int32 = 1
	raioDescarregamento int64 = 2
)

func (mundo *Mundo) carregarProximaRegiao(centro CoordRegiaoTerreno, catalogo *Catalogo) {
	for _, coord := range regioesVizinhas(centro, raioCarregamento) {
		if _, carregada := mundo.regioes[coord]; carregada {
			continue
		}

		inicio := time.Now()
		mundo.carregarRegiaoTerreno(coord)
		mundo.carregarArvoresRegiao(coord, catalogo.modeloObrigatorio(TipoArvore))
		mundo.carregarRochasRegiao(coord, catalogo.modeloObrigatorio(TipoRocha))
		mundo.carregarGramaRegiao(coord, catalogo.modeloObrigatorio(TipoGrama))
		mundo.carregarBordaVegetacaoRegiao(coord, catalogo)
		mundo.ultimoCarregamento = time.Since(inicio)

		fmt.Printf("Região carregada: %+v | ativas: %d | tempo: %s\n",
			coord, len(mundo.regioes), mundo.ultimoCarregamento)
		return
	}
}

func regioesVizinhas(centro CoordRegiaoTerreno, raio int32) []CoordRegiaoTerreno {
	regioes := []CoordRegiaoTerreno{centro}

	for dz := -raio; dz <= raio; dz++ {
		for dx := -raio; dx <= raio; dx++ {
			if dx == 0 && dz == 0 {
				continue
			}

			regioes = append(regioes, CoordRegiaoTerreno{
				X: centro.X + dx,
				Y: centro.Y,
				Z: centro.Z + dz,
			})
		}
	}

	return regioes
}

func (mundo *Mundo) descarregarRegiaoDistante(centro CoordRegiaoTerreno) {
	for coord := range mundo.regioes {
		dx := int64(coord.X) - int64(centro.X)
		dz := int64(coord.Z) - int64(centro.Z)

		dentroDaMargem := coord.Y == centro.Y &&
			dx >= -raioDescarregamento && dx <= raioDescarregamento &&
			dz >= -raioDescarregamento && dz <= raioDescarregamento

		if dentroDaMargem {
			continue
		}

		inicio := time.Now()
		mundo.descarregarRegiaoTerreno(coord)
		mundo.ultimaDescarga = time.Since(inicio)
		fmt.Printf("Região descarregada: %+v | ativas: %d | tempo: %s\n",
			coord, len(mundo.regioes), mundo.ultimaDescarga)
		return
	}
}
