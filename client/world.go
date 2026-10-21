package main

import (
	fastnoise "github.com/Auburn/FastNoiseLite/Go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mundo guarda o terreno e as instâncias presentes na partida.
type Mundo struct {
	semente       int
	versaoGerador int
	campo         [17][9][17]float32
	modeloTerreno rl.Model
	cubos         []CuboColetavel
	arvores       []Arvore
	rochas        []Rocha
	gramas        []Grama
	construcoes   []PecaConstrucao
	ocupacao      *IndiceOcupacao
	limitesRocha  rl.BoundingBox
	inimigos      []Inimigo
}

type PecaConstrucao struct {
	ID      IDObjeto
	Tipo    TipoObjeto
	Posicao rl.Vector3
	Escala  float32
	Rotacao float32
}

func novoMundo(ruido *fastnoise.State[float32], semente int, catalogo *Catalogo) Mundo {
	mundo := Mundo{
		semente:       semente,
		versaoGerador: versaoGerador,
		ocupacao:      novoIndiceOcupacao(2),
	}
	preencherCampo(&mundo.campo, ruido)
	mundo.modeloTerreno = criarModeloTerreno(&mundo.campo)
	mundo.cubos = gerarCubos(ruido, semente)

	registrarOcupacao(mundo.ocupacao, IDObjeto{Tipo: TipoInicioJogador}, TipoInicioJogador, rl.NewVector3(4, 0, 4))
	for _, cubo := range mundo.cubos {
		registrarOcupacao(mundo.ocupacao, cubo.ID, TipoCubo, cubo.Posicao)
	}

	modeloArvore := catalogo.modeloObrigatorio(TipoArvore)
	modeloRocha := catalogo.modeloObrigatorio(TipoRocha)
	modeloGrama := catalogo.modeloObrigatorio(TipoGrama)
	mundo.arvores = gerarArvores(mundo.modeloTerreno, modeloArvore, mundo.ocupacao, semente)
	mundo.limitesRocha = rl.GetModelBoundingBox(modeloRocha)
	mundo.rochas = gerarRochas(mundo.modeloTerreno, modeloRocha, mundo.ocupacao, semente)
	mundo.gramas = gerarGrama(mundo.modeloTerreno, modeloGrama, mundo.ocupacao, semente)
	return mundo
}

func (mundo *Mundo) adicionarPeca(tipo TipoObjeto, posicao rl.Vector3, escala float32, rotacao float32, modelo rl.Model) {
	id := IDObjeto{Tipo: tipo, Ordem: uint32(len(mundo.construcoes))}
	area := areaPecaXZ(modelo, posicao, escala, rotacao)

	registro := RegistroOcupacao{
		ID:      id,
		Tipo:    tipo,
		Posicao: posicao,
		Area:    area,
	}

	if !mundo.ocupacao.adicionar(registro) {
		panic("não foi possível registrar peça de construção")
	}
	mundo.construcoes = append(mundo.construcoes, PecaConstrucao{
		ID: id, Tipo: tipo, Posicao: posicao, Escala: escala, Rotacao: rotacao,
	})

	if tipo == TipoPiso {
		mundo.removerGramaNaArea(area)
	}

}
