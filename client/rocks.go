package main

import rl "github.com/gen2brain/raylib-go/raylib"

const escalaRocha float32 = 2

type Rocha struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Existe  bool
}

func gerarRochas(
	mundo *Mundo, modeloRocha rl.Model,
	indice *IndiceOcupacao,
	regiao RegiaoGeracao,
	candidatos []CandidatoObjeto,
	quantidade int,
) []Rocha {
	rochas := make([]Rocha, 0, quantidade)

	raioRocha := raioHorizontalLimites(mundo.limitesRocha, escalaRocha)

	var trechosEstrada [][2]PontoTerritorial
	if mundo.versaoGerador >= versaoGeradorComReservaEstrada {
		trechosEstrada = mundo.estradaInicial.trechosNaArea(
			regiao.Area,
			alcanceEstrada+raioRocha,
		)
	}

	tentar := func(candidato CandidatoObjeto) {

		if sobrepoeFaixaEstrada(trechosEstrada, candidato.X, candidato.Z, raioRocha) {
			return
		}

		posicao, ok := posicaoNoTerreno(mundo, modeloRocha, candidato.X, candidato.Z, escalaRocha, 0.9)
		if !ok {
			return
		}

		if !soloNaturalPreservado(mundo, modeloRocha, posicao, escalaRocha) {
			return
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

	for _, candidato := range candidatos {
		if len(rochas) >= quantidade {
			break
		}
		tentar(candidato)
	}

	return rochas
}

func (mundo *Mundo) carregarRochasRegiao(coord CoordRegiaoTerreno, modelo rl.Model) {
	regiao := mundo.regioes[coord]
	if coord.Y != 0 || regiao == nil || regiao.RochasCarregadas {
		return
	}

	area := regiaoGeracaoParaTipo(coord, TipoRocha, mundo.versaoGerador)
	candidatos := gerarCandidatos(TipoRocha, int64(mundo.semente), area, 200)
	novas := gerarRochas(mundo, modelo, mundo.ocupacao, area, candidatos, 4)

	for i := range novas {
		if _, removida := mundo.recursosRemovidos[novas[i].ID]; removida {
			novas[i].Existe = false
			mundo.ocupacao.remover(novas[i].ID)
		}
	}

	mundo.rochas = append(mundo.rochas, novas...)
	regiao.RochasCarregadas = true
}

func (mundo *Mundo) descarregarRochasRegiao(coord CoordRegiaoTerreno) {
	if coord.Y != 0 {
		return
	}

	restantes := mundo.rochas[:0]
	for _, rocha := range mundo.rochas {
		if rocha.ID.RegiaoX == coord.X && rocha.ID.RegiaoZ == coord.Z {
			mundo.ocupacao.remover(rocha.ID)
			continue
		}
		restantes = append(restantes, rocha)
	}
	mundo.rochas = restantes
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
