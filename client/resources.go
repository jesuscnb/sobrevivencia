package main

import (
	"math/rand"

	fastnoise "github.com/Auburn/FastNoiseLite/Go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type TipoAlvo uint8

const (
	AlvoNenhum TipoAlvo = iota
	AlvoCubo
	AlvoArvore
	AlvoRocha
)

type CuboColetavel struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Existe  bool
}

func novoCubo(ruido *fastnoise.State[float32], x, z float32) CuboColetavel {
	altura := densidadeNoPonto(ruido, x, 0, z)
	return CuboColetavel{
		Posicao: rl.NewVector3(x, altura+0.5, z),
		Existe:  true,
	}
}

func gerarCubos(ruido *fastnoise.State[float32], semente int) []CuboColetavel {
	pontos := [][2]float32{
		{-5, -5}, {-5, 0}, {-5, 5},
		{0, -5}, {0, 5},
		{5, -5}, {5, 0}, {5, 5},
	}
	sorteio := rand.New(rand.NewSource(int64(semente)))
	ordem := sorteio.Perm(len(pontos))
	cubos := make([]CuboColetavel, 0, 3)
	for _, indice := range ordem[:3] {
		ponto := pontos[indice]
		cubo := novoCubo(ruido, ponto[0], ponto[1])
		cubo.ID = IDObjeto{Tipo: TipoCubo, Ordem: uint32(len(cubos))}
		cubos = append(cubos, cubo)
	}
	return cubos
}

func caixaDoCubo(posicao rl.Vector3) rl.BoundingBox {
	return rl.NewBoundingBox(
		rl.NewVector3(posicao.X-0.5, posicao.Y-0.5, posicao.Z-0.5),
		rl.NewVector3(posicao.X+0.5, posicao.Y+0.5, posicao.Z+0.5),
	)
}

type ResultadoMira struct {
	colisaoTerreno   rl.RayCollision
	terrenoAoAlcance bool
	tipoAlvo         TipoAlvo
	indiceAlvo       int
}

func calcularMira(jogo *Jogo) ResultadoMira {

	centroTela := rl.NewVector2(
		float32(rl.GetScreenWidth())/2,
		float32(rl.GetScreenHeight())/2,
	)
	raioMira := rl.GetScreenToWorldRay(centroTela, jogo.camera)

	var colisaoTerreno rl.RayCollision
	malhas := jogo.mundo.modeloTerreno.GetMeshes()
	if len(malhas) > 0 {
		colisaoTerreno = rl.GetRayCollisionMesh(raioMira, malhas[0], jogo.mundo.modeloTerreno.Transform)
	}

	tipoAlvo := AlvoNenhum
	indiceAlvo := -1
	distanciaObjeto := float32(1000)

	// Guarda o objeto mais próximo atingido pelo raio da mira.
	considerar := func(tipo TipoAlvo, indice int, colisao rl.RayCollision) {
		if colisao.Hit && colisao.Distance < distanciaObjeto {
			tipoAlvo = tipo
			indiceAlvo = indice
			distanciaObjeto = colisao.Distance
		}
	}

	for i := range jogo.mundo.cubos {
		if jogo.mundo.cubos[i].Existe {
			colisao := rl.GetRayCollisionBox(raioMira, caixaDoCubo(jogo.mundo.cubos[i].Posicao))
			considerar(AlvoCubo, i, colisao)
		}
	}

	for i := range jogo.mundo.arvores {
		if jogo.mundo.arvores[i].Existe {
			colisao := rl.GetRayCollisionBox(raioMira, caixaDoTronco(jogo.mundo.arvores[i].Posicao))
			considerar(AlvoArvore, i, colisao)
		}
	}

	for i := range jogo.mundo.rochas {
		if jogo.mundo.rochas[i].Existe {
			caixa := caixaDaRocha(jogo.mundo.rochas[i].Posicao, jogo.mundo.limitesRocha)
			colisao := rl.GetRayCollisionBox(raioMira, caixa)
			considerar(AlvoRocha, i, colisao)
		}
	}

	objetoVisivel := tipoAlvo != AlvoNenhum &&
		distanciaObjeto <= 3 &&
		(!colisaoTerreno.Hit || distanciaObjeto < colisaoTerreno.Distance)

	if !objetoVisivel {
		tipoAlvo = AlvoNenhum
		indiceAlvo = -1
	}

	return ResultadoMira{
		colisaoTerreno: colisaoTerreno,
		terrenoAoAlcance: colisaoTerreno.Hit &&
			colisaoTerreno.Distance <= 4 &&
			colisaoTerreno.Distance < distanciaObjeto,
		tipoAlvo:   tipoAlvo,
		indiceAlvo: indiceAlvo,
	}

}
