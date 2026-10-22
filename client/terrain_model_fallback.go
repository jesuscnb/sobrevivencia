//go:build !windows || cgo

package main

import (
	"bufio"
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mantém o carregamento por OBJ nas plataformas ainda não validadas com a malha nativa.
func carregarModeloTerreno(dados DadosMalhaTerreno) rl.Model {
	arquivo, err := os.CreateTemp("", "terreno-*.obj")
	if err != nil {
		panic(err)
	}
	defer os.Remove(arquivo.Name())
	escritor := bufio.NewWriterSize(arquivo, 64*1024)

	for i := 0; i < len(dados.Vertices); i += 3 {
		if _, err := fmt.Fprintf(escritor, "v %f %f %f\n",
			dados.Vertices[i], dados.Vertices[i+1], dados.Vertices[i+2]); err != nil {
			arquivo.Close()
			panic(err)
		}
	}
	for i := 0; i < len(dados.Normais); i += 3 {
		if _, err := fmt.Fprintf(escritor, "vn %f %f %f\n",
			dados.Normais[i], dados.Normais[i+1], dados.Normais[i+2]); err != nil {
			arquivo.Close()
			panic(err)
		}
	}
	for t := 0; t < dados.Triangulos; t++ {
		inicio := t*3 + 1
		if _, err := fmt.Fprintf(escritor, "f %d//%d %d//%d %d//%d\n",
			inicio, inicio, inicio+1, inicio+1, inicio+2, inicio+2); err != nil {
			arquivo.Close()
			panic(err)
		}
	}
	if err := escritor.Flush(); err != nil {
		arquivo.Close()
		panic(err)
	}
	if err := arquivo.Close(); err != nil {
		panic(err)
	}

	modelo := rl.LoadModel(arquivo.Name())
	if !rl.IsModelValid(modelo) {
		panic("não foi possível carregar o modelo do terreno")
	}
	return modelo
}
