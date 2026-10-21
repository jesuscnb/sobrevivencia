package main

import (
	"math"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type TipoObjeto uint8

const (
	TipoCubo TipoObjeto = iota
	TipoArvore
	TipoRocha
	TipoGrama
	TipoPiso
	TipoParede
	TipoTelhado
	TipoInicioJogador
	TipoInimigo
)

// IDObjeto identifica uma instância sem depender da posição dela em uma lista.
type IDObjeto struct {
	Tipo             TipoObjeto
	RegiaoX, RegiaoZ int32
	Ordem            uint32
}

type AreaXZ struct {
	MinX, MaxX float32
	MinZ, MaxZ float32
}

type RegistroOcupacao struct {
	ID      IDObjeto
	Tipo    TipoObjeto
	Posicao rl.Vector3
	Area    AreaXZ
}

type celulaOcupacao struct {
	X, Z int
}

// IndiceOcupacao guarda somente a localização dos objetos. As regras que
// decidem se dois objetos podem ficar próximos estão em placement_rules.go.
type IndiceOcupacao struct {
	tamanhoCelula float32
	celulas       map[celulaOcupacao]map[IDObjeto]struct{}
	registros     map[IDObjeto]RegistroOcupacao
}

func novoIndiceOcupacao(tamanhoCelula float32) *IndiceOcupacao {
	if tamanhoCelula <= 0 {
		panic("o tamanho da célula deve ser positivo")
	}
	return &IndiceOcupacao{
		tamanhoCelula: tamanhoCelula,
		celulas:       make(map[celulaOcupacao]map[IDObjeto]struct{}),
		registros:     make(map[IDObjeto]RegistroOcupacao),
	}
}

func (indice *IndiceOcupacao) intervaloCelulas(area AreaXZ) (int, int, int, int) {
	tamanho := float64(indice.tamanhoCelula)
	return int(math.Floor(float64(area.MinX) / tamanho)),
		int(math.Floor(float64(area.MaxX) / tamanho)),
		int(math.Floor(float64(area.MinZ) / tamanho)),
		int(math.Floor(float64(area.MaxZ) / tamanho))
}

func (indice *IndiceOcupacao) adicionar(registro RegistroOcupacao) bool {
	if registro.Area.MinX > registro.Area.MaxX || registro.Area.MinZ > registro.Area.MaxZ {
		return false
	}
	if _, existe := indice.registros[registro.ID]; existe {
		return false
	}
	indice.registros[registro.ID] = registro
	minX, maxX, minZ, maxZ := indice.intervaloCelulas(registro.Area)
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			celula := celulaOcupacao{X: x, Z: z}
			if indice.celulas[celula] == nil {
				indice.celulas[celula] = make(map[IDObjeto]struct{})
			}
			indice.celulas[celula][registro.ID] = struct{}{}
		}
	}
	return true
}

func (indice *IndiceOcupacao) remover(id IDObjeto) {
	registro, existe := indice.registros[id]
	if !existe {
		return
	}
	minX, maxX, minZ, maxZ := indice.intervaloCelulas(registro.Area)
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			celula := celulaOcupacao{X: x, Z: z}
			delete(indice.celulas[celula], id)
			if len(indice.celulas[celula]) == 0 {
				delete(indice.celulas, celula)
			}
		}
	}
	delete(indice.registros, id)
}

func (indice *IndiceOcupacao) consultar(area AreaXZ) []RegistroOcupacao {
	if area.MinX > area.MaxX || area.MinZ > area.MaxZ {
		return nil
	}
	minX, maxX, minZ, maxZ := indice.intervaloCelulas(area)
	vistos := make(map[IDObjeto]struct{})
	var encontrados []RegistroOcupacao
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			for id := range indice.celulas[celulaOcupacao{X: x, Z: z}] {
				if _, repetido := vistos[id]; repetido {
					continue
				}
				vistos[id] = struct{}{}
				registro := indice.registros[id]
				if registro.Area.MinX <= area.MaxX && registro.Area.MaxX >= area.MinX &&
					registro.Area.MinZ <= area.MaxZ && registro.Area.MaxZ >= area.MinZ {
					encontrados = append(encontrados, registro)
				}
			}
		}
	}
	sort.Slice(encontrados, func(a, b int) bool {
		idA, idB := encontrados[a].ID, encontrados[b].ID
		if idA.Tipo != idB.Tipo {
			return idA.Tipo < idB.Tipo
		}
		if idA.RegiaoX != idB.RegiaoX {
			return idA.RegiaoX < idB.RegiaoX
		}
		if idA.RegiaoZ != idB.RegiaoZ {
			return idA.RegiaoZ < idB.RegiaoZ
		}
		return idA.Ordem < idB.Ordem
	})
	return encontrados
}
