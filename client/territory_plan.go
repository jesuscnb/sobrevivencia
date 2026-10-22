package main

import (
	"encoding/binary"
	"hash/fnv"
	"math"
)

const (
	tamanhoPlanoTerritorial = 5 * tamanhoRegiaoXZ // 80 unidades
	metadePlanoTerritorial  = tamanhoPlanoTerritorial / 2
)

const (
	finalidadeMarcoXBase byte = 16
	finalidadeMarcoZBase byte = 32
)

const (
	finalidadePortalX byte = 48
	finalidadePortalZ byte = 49
)

const (
	finalidadeCentroX      byte = 50
	finalidadeCentroZ      byte = 51
	finalidadeAssentamento byte = 52
)

type EixoFronteira byte

type LadoPlano byte

const (
	LadoNorte LadoPlano = iota
	LadoLeste
	LadoSul
	LadoOeste
)

const (
	EixoFronteiraX EixoFronteira = iota
	EixoFronteiraZ
)

type PontoTerritorial struct {
	X, Z float32
}

type CoordPlanoTerritorial struct {
	X, Z int64
}

func coordenadaPlano(x, z float32) CoordPlanoTerritorial {
	return CoordPlanoTerritorial{
		X: int64(math.Floor(
			(float64(x) + float64(metadePlanoTerritorial)) /
				float64(tamanhoPlanoTerritorial),
		)),
		Z: int64(math.Floor(
			(float64(z) + float64(metadePlanoTerritorial)) /
				float64(tamanhoPlanoTerritorial),
		)),
	}
}

func (coord CoordPlanoTerritorial) area() AreaXZ {
	minX := float32(coord.X*tamanhoPlanoTerritorial - metadePlanoTerritorial)
	minZ := float32(coord.Z*tamanhoPlanoTerritorial - metadePlanoTerritorial)

	return AreaXZ{
		MinX: minX,
		MaxX: minX + tamanhoPlanoTerritorial,
		MinZ: minZ,
		MaxZ: minZ + tamanhoPlanoTerritorial,
	}
}

const versaoPlanejamentoTerritorial = 1

func hashTerritorial(
	semente int64,
	coord CoordPlanoTerritorial,
	finalidade byte,
) uint64 {
	var dados [26]byte

	binary.LittleEndian.PutUint64(dados[0:8], uint64(semente))
	binary.LittleEndian.PutUint64(dados[8:16], uint64(coord.X))
	binary.LittleEndian.PutUint64(dados[16:24], uint64(coord.Z))
	dados[24] = byte(versaoPlanejamentoTerritorial)
	dados[25] = finalidade

	hash := fnv.New64a()
	_, _ = hash.Write(dados[:])
	return hash.Sum64()
}

func marcosIniciais(semente int64) [4]PontoTerritorial {
	// Ordem: noroeste, nordeste, sudeste, sudoeste.
	// Norte corresponde a Z negativo.
	marcos := [4]PontoTerritorial{
		{X: -26, Z: -26},
		{X: 26, Z: -26},
		{X: 26, Z: 26},
		{X: -26, Z: 26},
	}

	coord := CoordPlanoTerritorial{}
	for i := range marcos {
		deslocamentoX := int(hashTerritorial(
			semente, coord, finalidadeMarcoXBase+byte(i),
		)%9) - 4

		deslocamentoZ := int(hashTerritorial(
			semente, coord, finalidadeMarcoZBase+byte(i),
		)%9) - 4

		marcos[i].X += float32(deslocamentoX)
		marcos[i].Z += float32(deslocamentoZ)
	}

	return marcos
}

func portalFronteira(
	semente int64,
	coord CoordPlanoTerritorial,
	eixo EixoFronteira,
) PontoTerritorial {
	area := coord.area()

	switch eixo {
	case EixoFronteiraX:
		// Borda leste: X é fixo; Z varia ao longo da borda.
		deslocamento := 16 + float32(
			hashTerritorial(semente, coord, finalidadePortalX)%49,
		)
		return PontoTerritorial{
			X: area.MaxX,
			Z: area.MinZ + deslocamento,
		}

	case EixoFronteiraZ:
		// Borda sul: Z é fixo; X varia ao longo da borda.
		deslocamento := 16 + float32(
			hashTerritorial(semente, coord, finalidadePortalZ)%49,
		)
		return PontoTerritorial{
			X: area.MinX + deslocamento,
			Z: area.MaxZ,
		}

	default:
		panic("eixo de fronteira inválido")
	}
}

func portalDoLado(
	semente int64,
	coord CoordPlanoTerritorial,
	lado LadoPlano,
) PontoTerritorial {
	switch lado {
	case LadoLeste:
		return portalFronteira(semente, coord, EixoFronteiraX)

	case LadoOeste:
		vizinho := CoordPlanoTerritorial{X: coord.X - 1, Z: coord.Z}
		return portalFronteira(semente, vizinho, EixoFronteiraX)

	case LadoSul:
		return portalFronteira(semente, coord, EixoFronteiraZ)

	case LadoNorte:
		vizinho := CoordPlanoTerritorial{X: coord.X, Z: coord.Z - 1}
		return portalFronteira(semente, vizinho, EixoFronteiraZ)

	default:
		panic("lado territorial inválido")
	}
}

type PlanoTerritorial struct {
	ID                    CoordPlanoTerritorial
	Area                  AreaXZ
	CentroX, CentroZ      float32
	CandidatoAssentamento bool
}

func gerarPlanoTerritorial(
	semente int64,
	coord CoordPlanoTerritorial,
) PlanoTerritorial {
	area := coord.area()

	offsetX := float32(hashTerritorial(semente, coord, finalidadeCentroX) % 41)
	offsetZ := float32(hashTerritorial(semente, coord, finalidadeCentroZ) % 41)

	return PlanoTerritorial{
		ID:      coord,
		Area:    area,
		CentroX: area.MinX + 20 + offsetX,
		CentroZ: area.MinZ + 20 + offsetZ,
		CandidatoAssentamento: hashTerritorial(
			semente, coord, finalidadeAssentamento,
		)%8 == 0,
	}
}
