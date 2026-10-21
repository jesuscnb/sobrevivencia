package main

import (
	"fmt"
	"os"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector/vector3"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func criarModeloTerreno(campo *[17][9][17]float32) rl.Model {
	var amostrarDensidade sample.Vec3ToFloat = func(posicao vector3.Float64) float64 {
		return float64(densidadePonto(campo, int(posicao.X()), int(posicao.Y()), int(posicao.Z())))
	}

	fmt.Printf("Densidades em Y=-4, Y=0 e Y=4: %g, %g, %g\n",
		amostrarDensidade(vector3.New[float64](0, -4, 0)),
		amostrarDensidade(vector3.New[float64](0, 0, 0)),
		amostrarDensidade(vector3.New[float64](0, 4, 0)))

	dominio := geometry.NewAABB(vector3.Zero[float64](), vector3.New[float64](16, 8, 16))
	malhaTerreno := marching.March(amostrarDensidade, dominio, 1, 0)
	vertices := malhaTerreno.Float3Attribute(modeling.PositionAttribute)
	indices := malhaTerreno.Indices()

	dadosVertices := make([]float32, 0, indices.Len()*3)
	for t := 0; t < malhaTerreno.PrimitiveCount(); t++ {
		// Mantemos a ordem que fez a face aparecer voltada para cima.
		ordem := [3]int{
			indices.At(t * 3),
			indices.At(t*3 + 2),
			indices.At(t*3 + 1),
		}

		for _, indice := range ordem {
			p := vertices.At(indice)
			dadosVertices = append(dadosVertices,
				float32(p.X()),
				float32(p.Y()),
				float32(p.Z()),
			)
		}
	}

	arquivo, err := os.CreateTemp("", "terreno-*.obj")
	if err != nil {
		panic(err)
	}
	defer os.Remove(arquivo.Name())

	for i := 0; i < len(dadosVertices); i += 3 {
		fmt.Fprintf(arquivo, "v %f %f %f\n",
			dadosVertices[i],
			dadosVertices[i+1],
			dadosVertices[i+2],
		)
	}

	for t := 0; t < len(dadosVertices)/9; t++ {
		inicio := t*3 + 1
		fmt.Fprintf(arquivo, "f %d %d %d\n", inicio, inicio+1, inicio+2)
	}

	if err := arquivo.Close(); err != nil {
		panic(err)
	}

	modeloTerreno := rl.LoadModel(arquivo.Name())
	if !rl.IsModelValid(modeloTerreno) {
		panic("não foi possível carregar o modelo do terreno")
	}

	fmt.Printf("Vértices preparados: %d\n", len(dadosVertices)/3)
	fmt.Printf("Triângulos gerados para o terreno: %d\n", malhaTerreno.PrimitiveCount())
	return modeloTerreno
}
