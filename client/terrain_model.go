package main

import (
	"fmt"
	"math"
	"time"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector/vector3"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type DadosMalhaTerreno struct {
	Vertices   []float32
	Normais    []float32
	Triangulos int
}

type amostradorTerreno func(x, y, z int) float32

func prepararDadosMalhaTerreno(
	campo *[17][9][17]float32,
	coord CoordRegiaoTerreno,
	amostrarGlobal amostradorTerreno,
) DadosMalhaTerreno {
	var amostrarDensidade sample.Vec3ToFloat = func(posicao vector3.Float64) float64 {
		return float64(densidadePonto(
			campo,
			int(posicao.X()),
			int(posicao.Y()),
			int(posicao.Z()),
		))
	}

	dominio := geometry.NewAABB(
		vector3.Zero[float64](),
		vector3.New[float64](16, 8, 16),
	)
	malha := marching.March(amostrarDensidade, dominio, 1, 0)
	vertices := malha.Float3Attribute(modeling.PositionAttribute)
	indices := malha.Indices()
	cacheAmostras := make(map[CoordAmostraTerreno]float32)
	amostrar := func(x, y, z int) float32 {
		chave := CoordAmostraTerreno{X: x, Y: y, Z: z}
		if valor, existe := cacheAmostras[chave]; existe {
			return valor
		}
		valor := amostrarGlobal(x, y, z)
		cacheAmostras[chave] = valor
		return valor
	}
	// O mesmo vértice de divisa consulta os mesmos pontos globais dos dois lados.
	normaisVertices := make([][3]float32, vertices.Len())
	for i := range normaisVertices {
		p := vertices.At(i)
		x := float64(coord.X)*tamanhoRegiaoXZ + p.X()
		y := float64(coord.Y)*tamanhoRegiaoY + p.Y()
		z := float64(coord.Z)*tamanhoRegiaoXZ + p.Z()
		dx := densidadeInterpolada(amostrar, x+0.5, y, z) - densidadeInterpolada(amostrar, x-0.5, y, z)
		dy := densidadeInterpolada(amostrar, x, y+0.5, z) - densidadeInterpolada(amostrar, x, y-0.5, z)
		dz := densidadeInterpolada(amostrar, x, y, z+0.5) - densidadeInterpolada(amostrar, x, y, z-0.5)
		comprimento := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if comprimento < 0.000001 {
			normaisVertices[i] = [3]float32{0, 1, 0}
		} else {
			normaisVertices[i] = [3]float32{-dx / comprimento, -dy / comprimento, -dz / comprimento}
		}
	}

	dados := make([]float32, 0, indices.Len()*3)
	normais := make([]float32, 0, indices.Len()*3)
	for t := 0; t < malha.PrimitiveCount(); t++ {
		ordem := [3]int{
			indices.At(t * 3),
			indices.At(t*3 + 2),
			indices.At(t*3 + 1),
		}

		for _, indice := range ordem {
			p := vertices.At(indice)
			dados = append(dados, float32(p.X()), float32(p.Y()), float32(p.Z()))
			normais = append(normais, normaisVertices[indice][:]...)
		}
	}

	return DadosMalhaTerreno{
		Vertices:   dados,
		Normais:    normais,
		Triangulos: malha.PrimitiveCount(),
	}
}

func densidadeInterpolada(amostrar amostradorTerreno, x, y, z float64) float32 {
	x0, y0, z0 := int(math.Floor(x)), int(math.Floor(y)), int(math.Floor(z))
	tx, ty, tz := float32(x-float64(x0)), float32(y-float64(y0)), float32(z-float64(z0))
	misturar := func(a, b, peso float32) float32 { return a + (b-a)*peso }
	a := misturar(amostrar(x0, y0, z0), amostrar(x0+1, y0, z0), tx)
	b := misturar(amostrar(x0, y0, z0+1), amostrar(x0+1, y0, z0+1), tx)
	c := misturar(amostrar(x0, y0+1, z0), amostrar(x0+1, y0+1, z0), tx)
	d := misturar(amostrar(x0, y0+1, z0+1), amostrar(x0+1, y0+1, z0+1), tx)
	return misturar(misturar(a, b, tz), misturar(c, d, tz), ty)
}

func criarModeloTerreno(campo *[17][9][17]float32, coord CoordRegiaoTerreno, amostrarGlobal amostradorTerreno) rl.Model {
	inicio := time.Now()
	dadosMalha := prepararDadosMalhaTerreno(campo, coord, amostrarGlobal)
	tempoPreparacao := time.Since(inicio)
	inicioCarga := time.Now()
	modeloTerreno := carregarModeloTerreno(dadosMalha)
	fmt.Printf("Terreno: preparo=%s | carga=%s | triângulos=%d\n",
		tempoPreparacao, time.Since(inicioCarga), dadosMalha.Triangulos)
	return modeloTerreno
}
