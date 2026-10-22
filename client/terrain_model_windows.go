//go:build windows && !cgo

package main

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jupiterrider/ffi"
)

var memoriaTerreno struct {
	umaVez     sync.Once
	biblioteca ffi.Lib
	alocar     ffi.Fun
	liberar    ffi.Fun
	erro       error
}

// A raylib libera os arrays de uma Mesh no UnloadModel. Por isso os vértices
// precisam vir do alocador nativo dela, e nunca de uma fatia gerenciada pelo Go.
func alocarFloatsTerreno(quantidade int) *float32 {
	memoriaTerreno.umaVez.Do(func() {
		memoriaTerreno.biblioteca, memoriaTerreno.erro = ffi.Load("raylib.dll")
		if memoriaTerreno.erro != nil {
			return
		}
		memoriaTerreno.alocar, memoriaTerreno.erro = memoriaTerreno.biblioteca.Prep("MemAlloc", &ffi.TypePointer, &ffi.TypeUint32)
		if memoriaTerreno.erro != nil {
			return
		}
		memoriaTerreno.liberar, memoriaTerreno.erro = memoriaTerreno.biblioteca.Prep("MemFree", &ffi.TypeVoid, &ffi.TypePointer)
	})
	if memoriaTerreno.erro != nil {
		panic(fmt.Errorf("alocador da raylib indisponível: %w", memoriaTerreno.erro))
	}
	if quantidade <= 0 || quantidade > math.MaxUint32/4 {
		panic("tamanho inválido para os vértices do terreno")
	}
	tamanho := uint32(quantidade * 4)
	var ponteiro unsafe.Pointer
	memoriaTerreno.alocar.Call(&ponteiro, &tamanho)
	if ponteiro == nil {
		panic("sem memória para os vértices do terreno")
	}
	return (*float32)(ponteiro)
}

func liberarFloatsTerreno(ponteiro *float32) {
	p := unsafe.Pointer(ponteiro)
	memoriaTerreno.liberar.Call(nil, &p)
}

// Chamado na thread do jogo, que possui o contexto gráfico da raylib.
func carregarModeloTerreno(dados DadosMalhaTerreno) rl.Model {
	if dados.Triangulos <= 0 || len(dados.Vertices) != dados.Triangulos*9 ||
		len(dados.Normais) != len(dados.Vertices) || len(dados.Vertices)/3 > math.MaxInt32 {
		panic("dados inválidos para a malha do terreno")
	}

	posicoes := alocarFloatsTerreno(len(dados.Vertices))
	var normais *float32
	defer func() {
		if posicoes != nil {
			liberarFloatsTerreno(posicoes)
		}
		if normais != nil {
			liberarFloatsTerreno(normais)
		}
	}()
	copy(unsafe.Slice(posicoes, len(dados.Vertices)), dados.Vertices)
	normais = alocarFloatsTerreno(len(dados.Vertices))
	copy(unsafe.Slice(normais, len(dados.Normais)), dados.Normais)

	malha := rl.Mesh{
		VertexCount:   int32(len(dados.Vertices) / 3),
		TriangleCount: int32(dados.Triangulos),
		Vertices:      posicoes,
		Normals:       normais,
	}
	rl.UploadMesh(&malha, false)
	if malha.VboID == nil || *malha.VboID == 0 {
		rl.UnloadMesh(&malha)
		posicoes, normais = nil, nil
		panic("não foi possível enviar a malha do terreno para a GPU")
	}
	modelo := rl.LoadModelFromMesh(malha)
	// O modelo passa a ser dono dos dois buffers nativos.
	posicoes, normais = nil, nil
	if !rl.IsModelValid(modelo) {
		rl.UnloadModel(modelo)
		panic("não foi possível criar o modelo do terreno")
	}
	return modelo
}
