package main

import "math/rand"

const versaoGerador = 2

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

func regiaoInicial() RegiaoGeracao {
	return RegiaoGeracao{
		X:    0,
		Z:    0,
		Area: AreaXZ{MinX: -6, MaxX: 6, MinZ: -6, MaxZ: 6},
	}
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
