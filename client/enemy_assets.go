package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type TipoVisualInimigo uint8

const (
	VisualOrc TipoVisualInimigo = iota
	VisualBlob
	VisualGhost
)

type VisualInimigo struct {
	Modelo         rl.Model
	Escala         float32
	AjusteY        float32
	Animacoes      []rl.ModelAnimation
	Idle           rl.ModelAnimation
	Walk           rl.ModelAnimation
	Run            rl.ModelAnimation
	Punch          rl.ModelAnimation
	CorrecaoFrente float32
}

type CatalogoInimigos struct {
	visuais map[TipoVisualInimigo]VisualInimigo
}

type definicaoVisualInimigo struct {
	Tipo           TipoVisualInimigo
	Caminho        string
	Altura         float32
	NomeIdle       string
	NomeWalk       string
	NomeRun        string
	NomePunch      string
	CorrecaoFrente float32
	ElevacaoY      float32
}

func carregarVisualInimigo(def definicaoVisualInimigo) (VisualInimigo, error) {
	modelo := rl.LoadModel(def.Caminho)
	// Modelos animados em CPU podem ter ossos sem VBOs de skinning na GPU.
	// IsModelValid exige esses VBOs, embora a geometria e as animações funcionem.
	if modelo.MeshCount == 0 || modelo.MaterialCount == 0 ||
		modelo.Meshes == nil || modelo.Materials == nil {
		return VisualInimigo{}, fmt.Errorf(
			"não foi possível carregar %s", def.Caminho,
		)
	}
	for _, malha := range modelo.GetMeshes() {
		if malha.VertexCount == 0 || malha.Vertices == nil || malha.VaoID == 0 {
			descarregarModelos([]rl.Model{modelo})
			return VisualInimigo{}, fmt.Errorf("malha inválida em %s", def.Caminho)
		}
	}

	limites := rl.GetModelBoundingBox(modelo)
	alturaOriginal := limites.Max.Y - limites.Min.Y
	if alturaOriginal <= 0 {
		descarregarModelos([]rl.Model{modelo})
		return VisualInimigo{}, fmt.Errorf(
			"altura inválida no modelo %s", def.Caminho,
		)
	}

	animacoes := rl.LoadModelAnimations(def.Caminho)

	procurar := func(nome string) (rl.ModelAnimation, bool) {
		for _, animacao := range animacoes {
			if animacao.GetName() == nome &&
				animacao.KeyframeCount > 0 &&
				rl.IsModelAnimationValid(modelo, animacao) {
				return animacao, true
			}
		}
		return rl.ModelAnimation{}, false
	}

	idle, temIdle := procurar(def.NomeIdle)
	walk, temWalk := procurar(def.NomeWalk)
	run, temRun := procurar(def.NomeRun)
	punch, temPunch := procurar(def.NomePunch)

	if !temIdle || !temWalk || !temRun || !temPunch {
		if len(animacoes) > 0 {
			rl.UnloadModelAnimations(animacoes)
		}
		descarregarModelos([]rl.Model{modelo})
		return VisualInimigo{}, fmt.Errorf(
			"animações necessárias ausentes ou incompatíveis em %s",
			def.Caminho,
		)
	}

	escala := def.Altura / alturaOriginal

	return VisualInimigo{
		Modelo:         modelo,
		Escala:         escala,
		AjusteY:        -limites.Min.Y*escala + def.ElevacaoY,
		Animacoes:      animacoes,
		Idle:           idle,
		Walk:           walk,
		Run:            run,
		Punch:          punch,
		CorrecaoFrente: def.CorrecaoFrente,
	}, nil
}

func carregarCatalogoInimigos() (CatalogoInimigos, error) {
	definicoes := []definicaoVisualInimigo{
		{
			Tipo:           VisualOrc,
			Caminho:        "assets/enemies/Big/Orc.gltf",
			Altura:         1.6,
			NomeIdle:       "Idle",
			NomeWalk:       "Walk",
			NomeRun:        "Run",
			NomePunch:      "Punch",
			CorrecaoFrente: 0,
		},
		{
			Tipo:           VisualBlob,
			Caminho:        "assets/enemies/Blob/GreenBlob.gltf",
			Altura:         1.0,
			NomeIdle:       "Idle",
			NomeWalk:       "Walk",
			NomeRun:        "Walk",
			NomePunch:      "Bite_Front",
			CorrecaoFrente: 0,
		},
		{
			Tipo:           VisualGhost,
			Caminho:        "assets/enemies/Flying/Ghost.gltf",
			Altura:         1.0,
			ElevacaoY:      0.35,
			NomeIdle:       "Flying_Idle",
			NomeWalk:       "Flying_Idle",
			NomeRun:        "Fast_Flying",
			NomePunch:      "Headbutt",
			CorrecaoFrente: 0,
		},
	}

	catalogo := CatalogoInimigos{
		visuais: make(map[TipoVisualInimigo]VisualInimigo, len(definicoes)),
	}

	for _, def := range definicoes {
		visual, err := carregarVisualInimigo(def)
		if err != nil {
			catalogo.descarregar()
			return CatalogoInimigos{}, err
		}
		catalogo.visuais[def.Tipo] = visual
	}

	return catalogo, nil
}

func (catalogo CatalogoInimigos) obrigatorio(tipo TipoVisualInimigo) VisualInimigo {
	visual, existe := catalogo.visuais[tipo]
	if !existe {
		panic(fmt.Sprintf("visual do inimigo %d não encontrado", tipo))
	}
	return visual
}

func (catalogo *CatalogoInimigos) descarregar() {
	modelos := make([]rl.Model, 0, len(catalogo.visuais))
	for _, visual := range catalogo.visuais {
		if len(visual.Animacoes) > 0 {
			rl.UnloadModelAnimations(visual.Animacoes)
		}
		modelos = append(modelos, visual.Modelo)
	}
	descarregarModelos(modelos)
	catalogo.visuais = nil
}
