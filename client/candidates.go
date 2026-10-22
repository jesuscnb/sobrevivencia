package main

import "math/rand"

const (
	versaoGeradorAnterior          = 2
	versaoGeradorComReservaEstrada = 4
	versaoGerador                  = versaoGeradorComReservaEstrada
)

type RegiaoGeracao struct {
	X, Z int32
	Area AreaXZ
}

type CandidatoObjeto struct {
	Tipo            TipoObjeto
	Ordem           uint32
	X, Z            float32
	Escala, Rotacao float32
}

func regiaoGeracaoNaCoord(coord CoordRegiaoTerreno) RegiaoGeracao {
	centroX := float32(coord.X) * tamanhoRegiaoXZ
	centroZ := float32(coord.Z) * tamanhoRegiaoXZ
	metade := float32(tamanhoRegiaoXZ) / 2

	return RegiaoGeracao{
		X: coord.X,
		Z: coord.Z,
		Area: AreaXZ{
			MinX: centroX - metade,
			MaxX: centroX + metade,
			MinZ: centroZ - metade,
			MaxZ: centroZ + metade,
		},
	}
}

// A versão anterior usava apenas o quadrado central de 12x12. Mantê-lo
// permite que saves antigos recriem os recursos nas posições originais.
func regiaoGeracaoAnteriorNaCoord(coord CoordRegiaoTerreno) RegiaoGeracao {
	regiao := regiaoGeracaoNaCoord(coord)
	regiao.Area.MinX += 2
	regiao.Area.MaxX -= 2
	regiao.Area.MinZ += 2
	regiao.Area.MaxZ -= 2
	return regiao
}

func regiaoGeracaoParaTipo(coord CoordRegiaoTerreno, tipo TipoObjeto, versao int) RegiaoGeracao {
	if versao <= versaoGeradorAnterior {
		return regiaoGeracaoAnteriorNaCoord(coord)
	}

	regiao := regiaoGeracaoNaCoord(coord)
	var margem float32
	switch tipo {
	case TipoArvore:
		margem = 1.25 // Duas árvores vizinhas mantêm 2,5 unidades entre troncos.
	case TipoRocha:
		margem = 1 // Duas rochas vizinhas mantêm 2 unidades.
	case TipoGrama:
		margem = 0.2 // Fecha a faixa visual sem sobrepor grama entre regiões.
	case TipoInimigo:
		margem = 2 // Preserva a área inicial de surgimento dos inimigos.
	default:
		return regiao
	}

	regiao.Area.MinX += margem
	regiao.Area.MaxX -= margem
	regiao.Area.MinZ += margem
	regiao.Area.MaxZ -= margem
	return regiao
}

// Acrescenta candidatos somente na faixa que o gerador antigo deixava vazia.
// As ordens começam depois das tentativas antigas para manter os IDs salvos.
func gerarCandidatosBorda(tipo TipoObjeto, semente int64, coord CoordRegiaoTerreno, tentativas int) []CandidatoObjeto {
	areaAntiga := regiaoGeracaoAnteriorNaCoord(coord).Area
	areaNova := regiaoGeracaoParaTipo(coord, tipo, versaoGerador)
	candidatos := gerarCandidatos(tipo, semente, areaNova, tentativas)
	borda := make([]CandidatoObjeto, 0, len(candidatos))

	for _, candidato := range candidatos {
		if candidato.X >= areaAntiga.MinX && candidato.X < areaAntiga.MaxX &&
			candidato.Z >= areaAntiga.MinZ && candidato.Z < areaAntiga.MaxZ {
			continue
		}
		candidato.Ordem += uint32(tentativas)
		borda = append(borda, candidato)
	}
	return borda
}

// O sorteio de cada candidato é independente da aceitação dos anteriores.
// As coordenadas da região entram na semente para permitir expandir o mundo.
func gerarCandidatos(tipo TipoObjeto, semente int64, regiao RegiaoGeracao, tentativas int) []CandidatoObjeto {
	var deslocamentoTipo int64
	switch tipo {
	case TipoArvore:
		deslocamentoTipo = 1
	case TipoRocha:
		deslocamentoTipo = 2
	case TipoGrama:
		deslocamentoTipo = 3
	case TipoInimigo:
		deslocamentoTipo = 4
	default:
		return nil
	}

	sementeRegiao := semente + deslocamentoTipo +
		int64(regiao.X)*73856093 + int64(regiao.Z)*19349663
	sorteio := rand.New(rand.NewSource(sementeRegiao))
	candidatos := make([]CandidatoObjeto, 0, tentativas)

	for ordem := 0; ordem < tentativas; ordem++ {
		candidato := CandidatoObjeto{
			Tipo:  tipo,
			Ordem: uint32(ordem),
			X:     regiao.Area.MinX + sorteio.Float32()*(regiao.Area.MaxX-regiao.Area.MinX),
			Z:     regiao.Area.MinZ + sorteio.Float32()*(regiao.Area.MaxZ-regiao.Area.MinZ),
		}
		switch tipo {
		case TipoArvore:
			candidato.Escala = escalaArvore
		case TipoRocha:
			candidato.Escala = escalaRocha
		case TipoGrama:
			candidato.Escala = 2.4 + sorteio.Float32()*1.2
			candidato.Rotacao = sorteio.Float32() * 360
		}
		candidatos = append(candidatos, candidato)
	}
	return candidatos
}
