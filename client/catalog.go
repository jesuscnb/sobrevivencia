package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DefinicaoObjeto guarda dados compartilhados por todas as instâncias do tipo.
type DefinicaoObjeto struct {
	Caminho string
	Escala  float32
	Modelo  rl.Model
}

// Catalogo é o único dono dos modelos de árvores, pedras, grama e construção.
type Catalogo struct {
	tipos map[TipoObjeto]DefinicaoObjeto
	ordem []TipoObjeto
}

func carregarCatalogo() (Catalogo, error) {
	definicoes := []struct {
		tipo    TipoObjeto
		caminho string
		escala  float32
	}{
		{TipoArvore, "assets/nature-kit/models/tree_simple_fixed.glb", escalaArvore},
		{TipoRocha, "assets/nature-kit/models/rock_largeA_fixed.glb", escalaRocha},
		{TipoGrama, "assets/nature-kit/models/grass_leafs_fixed.glb", 2.4},
		{TipoPiso, "assets/build-kit/models/wood-floor.glb", escalaConstrucaoTeste},
		{TipoParede, "assets/build-kit/models/wall-pane-wood.glb", escalaConstrucaoTeste},
		{TipoTelhado, "assets/build-kit/models/roof.glb", escalaConstrucaoTeste},
	}

	catalogo := Catalogo{tipos: make(map[TipoObjeto]DefinicaoObjeto, len(definicoes))}

	for _, definicao := range definicoes {
		modelo := rl.LoadModel(definicao.caminho)

		if !rl.IsModelValid(modelo) {
			modelos := []rl.Model{modelo}
			for i := len(catalogo.ordem) - 1; i >= 0; i-- {
				modelos = append(modelos, catalogo.tipos[catalogo.ordem[i]].Modelo)
			}
			descarregarModelos(modelos)
			return Catalogo{}, fmt.Errorf("não foi possível carregar o modelo %q do tipo %d", definicao.caminho, definicao.tipo)
		}

		catalogo.tipos[definicao.tipo] = DefinicaoObjeto{
			Caminho: definicao.caminho,
			Escala:  definicao.escala,
			Modelo:  modelo,
		}

		catalogo.ordem = append(catalogo.ordem, definicao.tipo)
	}
	return catalogo, nil
}

func (catalogo *Catalogo) modelo(tipo TipoObjeto) (rl.Model, bool) {
	definicao, existe := catalogo.tipos[tipo]
	return definicao.Modelo, existe
}

func (catalogo *Catalogo) modeloObrigatorio(tipo TipoObjeto) rl.Model {
	modelo, existe := catalogo.modelo(tipo)
	if !existe {
		panic(fmt.Sprintf("tipo %d não está no catálogo de modelos", tipo))
	}
	return modelo
}

func (catalogo *Catalogo) escala(tipo TipoObjeto) float32 {
	return catalogo.tipos[tipo].Escala
}

func (catalogo *Catalogo) descarregar() {
	modelos := make([]rl.Model, 0, len(catalogo.ordem))
	for i := len(catalogo.ordem) - 1; i >= 0; i-- {
		tipo := catalogo.ordem[i]
		modelos = append(modelos, catalogo.tipos[tipo].Modelo)
	}
	descarregarModelos(modelos)
	catalogo.ordem = nil
	catalogo.tipos = nil
}

// UnloadModel deixa texturas sob responsabilidade de quem carregou o modelo.
// Reunimos os IDs antes de liberar os materiais, que guardam essas referências.
func descarregarModelos(modelos []rl.Model) {
	if len(modelos) == 0 {
		return
	}
	materialPadrao := rl.LoadMaterialDefault()
	idTexturaPadrao := materialPadrao.GetMap(rl.MapAlbedo).Texture.ID
	rl.UnloadMaterial(materialPadrao)

	vistos := make(map[uint32]struct{})
	var texturas []rl.Texture2D
	for _, modelo := range modelos {
		for _, material := range modelo.GetMaterials() {
			if material.Maps == nil {
				continue
			}
			for mapa := int32(0); mapa < rl.MaxMaterialMaps; mapa++ {
				textura := material.GetMap(mapa).Texture
				if textura.ID == 0 || textura.ID == idTexturaPadrao {
					continue
				}
				if _, repetida := vistos[textura.ID]; !repetida {
					vistos[textura.ID] = struct{}{}
					texturas = append(texturas, textura)
				}
			}
		}
	}
	for _, modelo := range modelos {
		rl.UnloadModel(modelo)
	}
	for _, textura := range texturas {
		rl.UnloadTexture(textura)
	}
}
