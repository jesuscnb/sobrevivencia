package main

import rl "github.com/gen2brain/raylib-go/raylib"

const escalaRocha float32 = 2

type Rocha struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Existe  bool
}

func gerarRochas(
	terreno, modeloRocha rl.Model,
	indice *IndiceOcupacao,
	semente int,
) []Rocha {
	const quantidade = 4
	const maxTentativas = 200

	rochas := make([]Rocha, 0, quantidade)
	regiao := regiaoInicial()
	candidatos := gerarCandidatos(TipoRocha, int64(semente), regiao, maxTentativas)

	for _, candidato := range candidatos {
		if len(rochas) >= quantidade {
			break
		}
		posicao, ok := posicaoNoTerreno(terreno, modeloRocha, candidato.X, candidato.Z, escalaRocha, 0.9)
		if !ok {
			continue
		}
		if podePosicionar(indice, TipoRocha, posicao) {
			rocha := Rocha{
				ID:      IDObjeto{Tipo: TipoRocha, RegiaoX: regiao.X, RegiaoZ: regiao.Z, Ordem: candidato.Ordem},
				Posicao: posicao,
				Existe:  true,
			}
			rochas = append(rochas, rocha)
			registrarOcupacao(indice, rocha.ID, TipoRocha, posicao)
		}
	}

	return rochas
}

func caixaDaRocha(posicao rl.Vector3, limites rl.BoundingBox) rl.BoundingBox {
	minimo := rl.NewVector3(
		posicao.X+limites.Min.X*escalaRocha,
		posicao.Y+limites.Min.Y*escalaRocha,
		posicao.Z+limites.Min.Z*escalaRocha,
	)
	maximo := rl.NewVector3(
		posicao.X+limites.Max.X*escalaRocha,
		posicao.Y+limites.Max.Y*escalaRocha,
		posicao.Z+limites.Max.Z*escalaRocha,
	)

	return rl.NewBoundingBox(minimo, maximo)
}
