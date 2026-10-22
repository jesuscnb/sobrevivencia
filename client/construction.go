package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	custoPisoMadeira    = 1
	custoParedeMadeira  = 1
	custoTelhadoMadeira = 1
)

type PreviaConstrucao struct {
	Visivel     bool
	Posicao     rl.Vector3
	PodeColocar bool
	Tipo        TipoObjeto
	Rotacao     float32
}

func calcularPreviaPiso(jogo *Jogo, mira ResultadoMira) PreviaConstrucao {
	if !jogo.modoConstrucao || !mira.terrenoAoAlcance {
		return PreviaConstrucao{}
	}

	const modulo float32 = 2
	ponto := mira.colisaoTerreno.Point
	x := float32(math.Round(float64(ponto.X/modulo))) * modulo
	z := float32(math.Round(float64(ponto.Z/modulo))) * modulo

	modelo := jogo.catalogo.modeloObrigatorio(TipoPiso)
	escala := jogo.catalogo.escala(TipoPiso)
	posicao, ok := posicaoNoTerreno(
		&jogo.mundo, modelo, x, z, escala, -1,
	)

	if !ok {
		return PreviaConstrucao{}
	}

	areaNova := areaPecaXZ(modelo, posicao, escala, 0)
	podeColocar := true

	for _, existente := range jogo.mundo.ocupacao.consultar(areaNova) {
		var areaOcupada AreaXZ

		switch existente.Tipo {
		case TipoPiso, TipoParede, TipoTelhado:
			areaOcupada = existente.Area
		case TipoArvore:
			areaOcupada = areaDaCaixaXZ(caixaDoTronco(existente.Posicao))
		case TipoRocha:
			caixa := caixaDaRocha(existente.Posicao, jogo.mundo.limitesRocha)
			areaOcupada = areaDaCaixaXZ(caixa)
		case TipoCubo:
			areaOcupada = areaDaCaixaXZ(caixaDoCubo(existente.Posicao))
		default:
			continue
		}

		if areasXZSobrepostas(areaNova, areaOcupada) {
			podeColocar = false
			break
		}
	}

	const raioJogador float32 = 0.35
	areaJogador := areaAoRedor(jogo.camera.Position, raioJogador)

	if areasXZSobrepostas(areaNova, areaJogador) {
		podeColocar = false
	}

	if podeColocar {
		baseDoPiso := posicao.Y + rl.GetModelBoundingBox(modelo).Min.Y*escala

		if !pisoTemApoio(&jogo.mundo, areaNova, baseDoPiso) {
			podeColocar = false
		}
	}

	if jogo.inventario.Madeira < custoPisoMadeira {
		podeColocar = false
	}

	return PreviaConstrucao{
		Visivel:     true,
		Posicao:     posicao,
		Tipo:        TipoPiso,
		PodeColocar: podeColocar,
	}

}

func pisoSobMira(jogo *Jogo, mira ResultadoMira, alturaExtra float32) (RegistroOcupacao, rl.RayCollision, float32, bool) {
	modeloPiso := jogo.catalogo.modeloObrigatorio(TipoPiso)
	escalaPiso := jogo.catalogo.escala(TipoPiso)
	limitesPiso := rl.GetModelBoundingBox(modeloPiso)

	centroTela := rl.NewVector2(
		float32(rl.GetScreenWidth())/2,
		float32(rl.GetScreenHeight())/2,
	)
	raio := rl.GetScreenToWorldRay(centroTela, jogo.camera)

	var piso RegistroOcupacao
	var contato rl.RayCollision
	encontrou := false
	menorDistancia := float32(4)

	for _, candidato := range jogo.mundo.ocupacao.consultar(
		areaAoRedor(jogo.camera.Position, 4),
	) {
		if candidato.Tipo != TipoPiso {
			continue
		}

		caixa := rl.NewBoundingBox(
			rl.NewVector3(
				candidato.Posicao.X+limitesPiso.Min.X*escalaPiso,
				candidato.Posicao.Y+limitesPiso.Min.Y*escalaPiso,
				candidato.Posicao.Z+limitesPiso.Min.Z*escalaPiso,
			),
			rl.NewVector3(
				candidato.Posicao.X+limitesPiso.Max.X*escalaPiso,
				candidato.Posicao.Y+limitesPiso.Max.Y*escalaPiso+alturaExtra,
				candidato.Posicao.Z+limitesPiso.Max.Z*escalaPiso,
			),
		)

		colisao := rl.GetRayCollisionBox(raio, caixa)
		if !colisao.Hit || colisao.Distance > menorDistancia {
			continue
		}
		if mira.colisaoTerreno.Hit &&
			colisao.Distance > mira.colisaoTerreno.Distance+0.01 {
			continue
		}

		piso = candidato
		contato = colisao
		menorDistancia = colisao.Distance
		encontrou = true
	}

	if !encontrou {
		return RegistroOcupacao{}, rl.RayCollision{}, 0, false
	}

	topoPiso := piso.Posicao.Y + limitesPiso.Max.Y*escalaPiso
	return piso, contato, topoPiso, true
}

func calcularPreviaParede(jogo *Jogo, mira ResultadoMira) PreviaConstrucao {
	if !jogo.modoConstrucao {
		return PreviaConstrucao{}
	}

	piso, contato, topoPiso, encontrou := pisoSobMira(jogo, mira, 0)
	if !encontrou {
		return PreviaConstrucao{}
	}

	modeloParede := jogo.catalogo.modeloObrigatorio(TipoParede)
	escalaParede := jogo.catalogo.escala(TipoParede)
	limitesParede := rl.GetModelBoundingBox(modeloParede)

	centroX := (piso.Area.MinX + piso.Area.MaxX) / 2
	centroZ := (piso.Area.MinZ + piso.Area.MaxZ) / 2

	posicao := rl.NewVector3(
		centroX,
		topoPiso-limitesParede.Min.Y*escalaParede,
		centroZ,
	)
	var rotacao float32

	dx := contato.Point.X - centroX
	dz := contato.Point.Z - centroZ

	if math.Abs(float64(dx)) > math.Abs(float64(dz)) {
		rotacao = 90
		if dx > 0 {
			posicao.X = piso.Area.MaxX
		} else {
			posicao.X = piso.Area.MinX
		}
	} else {
		if dz > 0 {
			posicao.Z = piso.Area.MaxZ
		} else {
			posicao.Z = piso.Area.MinZ
		}
	}

	podeColocar := jogo.inventario.Madeira >= custoParedeMadeira

	alturaParede := (limitesParede.Max.Y - limitesParede.Min.Y) * escalaParede
	for _, existente := range jogo.mundo.ocupacao.consultar(
		areaAoRedor(posicao, 0.1),
	) {
		if existente.Tipo != TipoParede {
			continue
		}

		mesmoLugar := math.Abs(float64(existente.Posicao.X-posicao.X)) < 0.05 &&
			math.Abs(float64(existente.Posicao.Z-posicao.Z)) < 0.05
		alturasCruzam := math.Abs(float64(existente.Posicao.Y-posicao.Y)) <
			float64(alturaParede-0.05)

		if mesmoLugar && alturasCruzam {
			podeColocar = false
			break
		}
	}

	return PreviaConstrucao{
		Visivel:     true,
		Tipo:        TipoParede,
		Posicao:     posicao,
		Rotacao:     rotacao,
		PodeColocar: podeColocar,
	}
}

func tentarConstruirPiso(jogo *Jogo) bool {

	previa := jogo.previaConstrucao

	if !jogo.modoConstrucao || !previa.Visivel || !previa.PodeColocar || previa.Tipo != TipoPiso {
		return false
	}

	if jogo.inventario.Madeira < custoPisoMadeira {
		return false
	}

	modelo := jogo.catalogo.modeloObrigatorio(TipoPiso)
	escala := jogo.catalogo.escala(TipoPiso)

	jogo.mundo.adicionarPeca(TipoPiso, previa.Posicao, escala, 0, modelo)
	jogo.inventario.Madeira -= custoPisoMadeira

	return true
}

func tentarConstruirParede(jogo *Jogo) bool {
	previa := jogo.previaConstrucao

	if !jogo.modoConstrucao || !previa.Visivel ||
		!previa.PodeColocar || previa.Tipo != TipoParede {
		return false
	}

	if jogo.inventario.Madeira < custoParedeMadeira {
		return false
	}

	modelo := jogo.catalogo.modeloObrigatorio(TipoParede)
	escala := jogo.catalogo.escala(TipoParede)

	jogo.mundo.adicionarPeca(
		TipoParede,
		previa.Posicao,
		escala,
		previa.Rotacao,
		modelo,
	)
	jogo.inventario.Madeira -= custoParedeMadeira
	return true
}

func calcularPreviaTelhado(jogo *Jogo, mira ResultadoMira) PreviaConstrucao {
	if !jogo.modoConstrucao {
		return PreviaConstrucao{}
	}

	modeloParede := jogo.catalogo.modeloObrigatorio(TipoParede)
	escalaParede := jogo.catalogo.escala(TipoParede)
	limitesParede := rl.GetModelBoundingBox(modeloParede)
	alturaParede := (limitesParede.Max.Y - limitesParede.Min.Y) * escalaParede

	modeloTelhado := jogo.catalogo.modeloObrigatorio(TipoTelhado)
	escalaTelhado := jogo.catalogo.escala(TipoTelhado)
	limitesTelhado := rl.GetModelBoundingBox(modeloTelhado)
	alturaTelhado := (limitesTelhado.Max.Y - limitesTelhado.Min.Y) * escalaTelhado

	piso, _, topoPiso, encontrou := pisoSobMira(
		jogo, mira, alturaParede+alturaTelhado,
	)
	if !encontrou {
		return PreviaConstrucao{}
	}

	centroX := (piso.Area.MinX + piso.Area.MaxX) / 2
	centroZ := (piso.Area.MinZ + piso.Area.MaxZ) / 2
	posicao := rl.NewVector3(
		centroX,
		topoPiso+alturaParede-limitesTelhado.Min.Y*escalaTelhado,
		centroZ,
	)

	baseParede := topoPiso - limitesParede.Min.Y*escalaParede
	bordas := [4]rl.Vector3{
		rl.NewVector3(piso.Area.MinX, baseParede, centroZ),
		rl.NewVector3(piso.Area.MaxX, baseParede, centroZ),
		rl.NewVector3(centroX, baseParede, piso.Area.MinZ),
		rl.NewVector3(centroX, baseParede, piso.Area.MaxZ),
	}
	var temParede [4]bool
	jaTemTelhado := false

	for _, existente := range jogo.mundo.ocupacao.consultar(piso.Area) {
		switch existente.Tipo {
		case TipoParede:
			for i, borda := range bordas {
				if math.Abs(float64(existente.Posicao.X-borda.X)) < 0.05 &&
					math.Abs(float64(existente.Posicao.Y-borda.Y)) < 0.05 &&
					math.Abs(float64(existente.Posicao.Z-borda.Z)) < 0.05 {
					temParede[i] = true
				}
			}

		case TipoTelhado:
			if math.Abs(float64(existente.Posicao.X-posicao.X)) < 0.05 &&
				math.Abs(float64(existente.Posicao.Y-posicao.Y)) < 0.05 &&
				math.Abs(float64(existente.Posicao.Z-posicao.Z)) < 0.05 {
				jaTemTelhado = true
			}
		}
	}

	apoios := 0
	for _, presente := range temParede {
		if presente {
			apoios++
		}
	}

	return PreviaConstrucao{
		Visivel: true,
		Tipo:    TipoTelhado,
		Posicao: posicao,
		PodeColocar: apoios >= 3 &&
			!jaTemTelhado &&
			jogo.inventario.Madeira >= custoTelhadoMadeira,
	}
}

func tentarConstruirTelhado(jogo *Jogo) bool {
	previa := jogo.previaConstrucao

	if !jogo.modoConstrucao || !previa.Visivel ||
		!previa.PodeColocar || previa.Tipo != TipoTelhado {
		return false
	}
	if jogo.inventario.Madeira < custoTelhadoMadeira {
		return false
	}

	modelo := jogo.catalogo.modeloObrigatorio(TipoTelhado)
	escala := jogo.catalogo.escala(TipoTelhado)

	jogo.mundo.adicionarPeca(
		TipoTelhado, previa.Posicao, escala, previa.Rotacao, modelo,
	)
	jogo.inventario.Madeira -= custoTelhadoMadeira
	return true
}
