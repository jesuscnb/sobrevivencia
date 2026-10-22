package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Arvore struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Existe  bool
}

const escalaArvore float32 = 3
const raioTronco float32 = 0.45

func criarArvoreNoTerreno(mundo *Mundo, modeloArvore rl.Model, x, z float32) (Arvore, bool) {
	posicao, ok := posicaoNoTerreno(mundo, modeloArvore, x, z, escalaArvore, 0.85)
	if !ok {
		return Arvore{}, false
	}
	if !soloNaturalPreservado(mundo, modeloArvore, posicao, escalaArvore) {
		return Arvore{}, false
	}

	return Arvore{Posicao: posicao, Existe: true}, true
}

func gerarArvores(mundo *Mundo, modeloArvore rl.Model, indice *IndiceOcupacao, regiao RegiaoGeracao, candidatos []CandidatoObjeto, quantidade int) []Arvore {
	arvores := make([]Arvore, 0, quantidade)

	var trechosEstrada [][2]PontoTerritorial
	if mundo.versaoGerador >= versaoGeradorComReservaEstrada {
		trechosEstrada = mundo.estradaInicial.trechosNaArea(
			regiao.Area,
			alcanceEstrada+raioTronco,
		)
	}

	tentar := func(candidato CandidatoObjeto) {

		if sobrepoeFaixaEstrada(trechosEstrada, candidato.X, candidato.Z, raioTronco) {
			return
		}

		arvore, encontrouChao := criarArvoreNoTerreno(mundo, modeloArvore, candidato.X, candidato.Z)
		if !encontrouChao {
			return
		}

		if podePosicionar(indice, TipoArvore, arvore.Posicao) {
			arvore.ID = IDObjeto{Tipo: TipoArvore, RegiaoX: regiao.X, RegiaoZ: regiao.Z, Ordem: candidato.Ordem}
			arvores = append(arvores, arvore)
			registrarOcupacao(indice, arvore.ID, TipoArvore, arvore.Posicao)
		}
	}

	for _, candidato := range candidatos {
		if len(arvores) >= quantidade {
			break
		}
		tentar(candidato)
	}

	return arvores
}

func (mundo *Mundo) carregarArvoresRegiao(coord CoordRegiaoTerreno, modelo rl.Model) {
	regiao := mundo.regioes[coord]
	if coord.Y != 0 || regiao == nil || regiao.ArvoresCarregadas {
		return
	}

	area := regiaoGeracaoParaTipo(coord, TipoArvore, mundo.versaoGerador)
	candidatos := gerarCandidatos(TipoArvore, int64(mundo.semente), area, 100)
	novas := gerarArvores(mundo, modelo, mundo.ocupacao, area, candidatos, 6)

	for i := range novas {
		if _, removida := mundo.recursosRemovidos[novas[i].ID]; removida {
			novas[i].Existe = false
			mundo.ocupacao.remover(novas[i].ID)
		}
	}

	mundo.arvores = append(mundo.arvores, novas...)
	regiao.ArvoresCarregadas = true
}

func (mundo *Mundo) descarregarArvoresRegiao(coord CoordRegiaoTerreno) {
	if coord.Y != 0 {
		return
	}

	restantes := mundo.arvores[:0]
	for _, arvore := range mundo.arvores {
		if arvore.ID.RegiaoX == coord.X && arvore.ID.RegiaoZ == coord.Z {
			mundo.ocupacao.remover(arvore.ID)
			continue
		}
		restantes = append(restantes, arvore)
	}
	mundo.arvores = restantes
}

func caixaDoTronco(posicao rl.Vector3) rl.BoundingBox {
	return rl.NewBoundingBox(
		rl.NewVector3(posicao.X-raioTronco, posicao.Y, posicao.Z-raioTronco),
		rl.NewVector3(posicao.X+raioTronco, posicao.Y+2.4, posicao.Z+raioTronco),
	)
}
