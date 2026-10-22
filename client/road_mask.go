package main

import (
	"math"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	amostrasPorUnidade = 4
	resolucaoMascara   = tamanhoRegiaoXZ * amostrasPorUnidade // 64
	tamanhoMascara     = resolucaoMascara + 2                 // 66, com bordas extras

	meiaLarguraEstrada float32 = 1.2
	transicaoBorda     float32 = 0.6
	alcanceEstrada     float32 = meiaLarguraEstrada + transicaoBorda
)

func distanciaQuadradaPontoTrecho(x, z float32, a, b PontoTerritorial) float32 {
	dx := b.X - a.X
	dz := b.Z - a.Z
	comprimentoQuadrado := dx*dx + dz*dz

	if comprimentoQuadrado == 0 {
		dx = x - a.X
		dz = z - a.Z
		return dx*dx + dz*dz
	}

	t := ((x-a.X)*dx + (z-a.Z)*dz) / comprimentoQuadrado
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	pontoX := a.X + t*dx
	pontoZ := a.Z + t*dz
	dx = x - pontoX
	dz = z - pontoZ
	return dx*dx + dz*dz
}

func gerarDadosMascaraEstrada(c CaminhoEstrada, coord CoordRegiaoTerreno) []byte {
	if coord.Y != 0 {
		return nil
	}

	area := coord.areaXZ()
	const metadePixel = float32(0.5) / amostrasPorUnidade
	trechos := c.trechosNaArea(area, alcanceEstrada+metadePixel)
	if len(trechos) == 0 {
		return nil
	}

	pixels := make([]byte, tamanhoMascara*tamanhoMascara)
	alcanceQuadrado := alcanceEstrada * alcanceEstrada

	for iz := 0; iz < tamanhoMascara; iz++ {
		z := area.MinZ + (float32(iz)-0.5)/amostrasPorUnidade

		for ix := 0; ix < tamanhoMascara; ix++ {
			x := area.MinX + (float32(ix)-0.5)/amostrasPorUnidade
			menorDistanciaQuadrada := alcanceQuadrado

			for _, trecho := range trechos {
				distancia := distanciaQuadradaPontoTrecho(x, z, trecho[0], trecho[1])
				if distancia < menorDistanciaQuadrada {
					menorDistanciaQuadrada = distancia
				}
			}

			if menorDistanciaQuadrada >= alcanceQuadrado {
				continue
			}

			distancia := float32(math.Sqrt(float64(menorDistanciaQuadrada)))
			intensidade := (alcanceEstrada - distancia) / transicaoBorda
			if intensidade > 1 {
				intensidade = 1
			}

			// Suaviza a passagem entre estrada e grama.
			intensidade = intensidade * intensidade * (3 - 2*intensidade)
			pixels[iz*tamanhoMascara+ix] = byte(intensidade*255 + 0.5)
		}
	}

	return pixels
}

func carregarMascaraEstrada(c CaminhoEstrada, coord CoordRegiaoTerreno) rl.Texture2D {
	dados := gerarDadosMascaraEstrada(c, coord)
	if len(dados) == 0 {
		return rl.Texture2D{}
	}

	imagem := rl.NewImage(
		dados,
		int32(tamanhoMascara),
		int32(tamanhoMascara),
		1,
		rl.UncompressedGrayscale,
	)

	textura := rl.LoadTextureFromImage(imagem)
	runtime.KeepAlive(dados)

	if textura.ID == 0 {
		panic("não foi possível carregar a máscara da estrada")
	}

	rl.SetTextureFilter(textura, rl.FilterBilinear)
	rl.SetTextureWrap(textura, rl.WrapClamp)
	return textura
}
