package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func raioHorizontalLimites(limites rl.BoundingBox, escala float32) float32 {
	x := math.Max(
		math.Abs(float64(limites.Min.X)),
		math.Abs(float64(limites.Max.X)),
	)
	z := math.Max(
		math.Abs(float64(limites.Min.Z)),
		math.Abs(float64(limites.Max.Z)),
	)
	return float32(math.Hypot(x, z)) * escala
}

// Cada valor mantém a distância usada antes pelos geradores. A ordem dos
// argumentos importa: primeiro vem o candidato, depois o objeto já colocado.
func distanciaMinima(tipoNovo, tipoExistente TipoObjeto) float32 {
	switch tipoNovo {
	case TipoArvore:
		switch tipoExistente {
		case TipoInicioJogador:
			return 3
		case TipoCubo:
			return 2
		case TipoArvore:
			return 2.5
		}
	case TipoRocha:
		switch tipoExistente {
		case TipoInicioJogador:
			return 2.5
		case TipoCubo:
			return 1.6
		case TipoArvore:
			return 1.4
		case TipoRocha:
			return 2
		}
	case TipoGrama:
		switch tipoExistente {
		case TipoInicioJogador:
			return 1.5
		case TipoCubo:
			return 0.8
		case TipoArvore:
			return 0.7
		case TipoRocha:
			return 1.2
		case TipoGrama:
			return 0.35
		}
	case TipoInimigo:
		switch tipoExistente {
		case TipoInicioJogador:
			return 3
		case TipoCubo, TipoArvore, TipoRocha:
			return 1.5
		case TipoPiso, TipoParede, TipoTelhado:
			return 2
		}
	}
	return 0
}

func areaAoRedor(posicao rl.Vector3, raio float32) AreaXZ {
	return AreaXZ{
		MinX: posicao.X - raio,
		MaxX: posicao.X + raio,
		MinZ: posicao.Z - raio,
		MaxZ: posicao.Z + raio,
	}
}

// A área de influência é um filtro amplo; distanciaMinima faz a decisão final.
func areaInfluencia(tipo TipoObjeto, posicao rl.Vector3) AreaXZ {
	var raio float32
	switch tipo {
	case TipoInicioJogador:
		raio = 3
	case TipoCubo:
		raio = 2
	case TipoArvore:
		raio = 2.5
	case TipoRocha:
		raio = 2
	case TipoGrama:
		raio = 0.35
	}
	return areaAoRedor(posicao, raio)
}

func podePosicionar(indice *IndiceOcupacao, tipo TipoObjeto, posicao rl.Vector3) bool {
	ponto := areaAoRedor(posicao, 0)
	for _, vizinho := range indice.consultar(ponto) {
		distancia := distanciaMinima(tipo, vizinho.Tipo)
		if distancia > 0 && distanciaHorizontal2(posicao, vizinho.Posicao) < distancia*distancia {
			return false
		}
	}
	return true
}

func registrarOcupacao(indice *IndiceOcupacao, id IDObjeto, tipo TipoObjeto, posicao rl.Vector3) {
	registro := RegistroOcupacao{
		ID:      id,
		Tipo:    tipo,
		Posicao: posicao,
		Area:    areaInfluencia(tipo, posicao),
	}
	if !indice.adicionar(registro) {
		panic("ID de objeto duplicado ou área de ocupação inválida")
	}
}
