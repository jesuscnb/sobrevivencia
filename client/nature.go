package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Arvore struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Existe  bool
}

const escalaArvore float32 = 3

func criarArvoreNoTerreno(terreno, modeloArvore rl.Model, x, z float32) (Arvore, bool) {
	posicao, ok := posicaoNoTerreno(terreno, modeloArvore, x, z, escalaArvore, 0.85)
	if !ok {
		return Arvore{}, false
	}

	return Arvore{Posicao: posicao, Existe: true}, true
}

func gerarArvores(terreno, modeloArvore rl.Model, indice *IndiceOcupacao, semente int) []Arvore {
	const quantidade = 6
	const maxTentativas = 100

	arvores := make([]Arvore, 0, quantidade)
	regiao := regiaoInicial()
	candidatos := gerarCandidatos(TipoArvore, int64(semente), regiao, maxTentativas)

	for _, candidato := range candidatos {
		if len(arvores) >= quantidade {
			break
		}
		arvore, encontrouChao := criarArvoreNoTerreno(terreno, modeloArvore, candidato.X, candidato.Z)
		if !encontrouChao {
			continue
		}

		if podePosicionar(indice, TipoArvore, arvore.Posicao) {
			arvore.ID = IDObjeto{Tipo: TipoArvore, RegiaoX: regiao.X, RegiaoZ: regiao.Z, Ordem: candidato.Ordem}
			arvores = append(arvores, arvore)
			registrarOcupacao(indice, arvore.ID, TipoArvore, arvore.Posicao)
		}
	}

	return arvores
}

func caixaDoTronco(posicao rl.Vector3) rl.BoundingBox {
	const raio float32 = 0.45

	return rl.NewBoundingBox(
		rl.NewVector3(posicao.X-raio, posicao.Y, posicao.Z-raio),
		rl.NewVector3(posicao.X+raio, posicao.Y+2.4, posicao.Z+raio),
	)
}
