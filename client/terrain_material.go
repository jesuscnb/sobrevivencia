package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Um único material e duas texturas atendem todas as malhas de terreno.
// Os modelos emprestam o shader; Mundo continua responsável apenas pelas malhas.
type MaterialTerreno struct {
	Shader        rl.Shader
	Padrao        rl.Shader
	TexturaPadrao rl.Texture2D
	Grama         rl.Texture2D
	Solo          rl.Texture2D
	localGrama    int32
	localSolo     int32
}

func carregarMaterialTerreno(modeloInicial rl.Model) MaterialTerreno {
	materiais := modeloInicial.GetMaterials()
	if len(materiais) == 0 {
		panic("modelo do terreno sem material")
	}
	material := MaterialTerreno{
		Padrao:        materiais[0].Shader,
		TexturaPadrao: materiais[0].GetMap(rl.MapAlbedo).Texture,
	}
	material.Shader = rl.LoadShader("assets/terrain/terrain.vs", "assets/terrain/terrain.fs")
	if !rl.IsShaderValid(material.Shader) {
		panic("não foi possível carregar o shader do terreno")
	}
	material.localGrama = rl.GetShaderLocation(material.Shader, "grassTexture")
	material.localSolo = rl.GetShaderLocation(material.Shader, "soilTexture")
	if material.localGrama < 0 || material.localSolo < 0 {
		rl.UnloadShader(material.Shader)
		panic("shader do terreno sem os controles das texturas")
	}
	material.Shader.UpdateLocation(rl.ShaderLocMapAlbedo, material.localGrama)
	material.Shader.UpdateLocation(rl.ShaderLocMapMetalness, material.localSolo)
	material.Grama = rl.LoadTexture("assets/terrain/grass.png")
	material.Solo = rl.LoadTexture("assets/terrain/soil.png")
	if material.Grama.ID == 0 || material.Solo.ID == 0 {
		material.descarregar()
		panic("não foi possível carregar as texturas do terreno")
	}
	rl.GenTextureMipmaps(&material.Grama)
	rl.GenTextureMipmaps(&material.Solo)
	rl.SetTextureFilter(material.Grama, rl.FilterTrilinear)
	rl.SetTextureFilter(material.Solo, rl.FilterTrilinear)
	rl.SetTextureWrap(material.Grama, rl.WrapRepeat)
	rl.SetTextureWrap(material.Solo, rl.WrapRepeat)
	fmt.Println("Material do terreno: shader e duas texturas carregados")
	return material
}

func (material *MaterialTerreno) usarNoModelo(modelo rl.Model, texturizado bool) {
	shader := material.Padrao
	if texturizado {
		shader = material.Shader
	}
	materiais := modelo.GetMaterials()
	for i := range materiais {
		materiais[i].Shader = shader
		if texturizado {
			materiais[i].GetMap(rl.MapAlbedo).Texture = material.Grama
			materiais[i].GetMap(rl.MapMetalness).Texture = material.Solo
		} else {
			materiais[i].GetMap(rl.MapAlbedo).Texture = material.TexturaPadrao
		}
	}
}

func (material *MaterialTerreno) descarregar() {
	if material.Grama.ID != 0 {
		rl.UnloadTexture(material.Grama)
	}
	if material.Solo.ID != 0 {
		rl.UnloadTexture(material.Solo)
	}
	if material.Shader.ID != 0 {
		rl.UnloadShader(material.Shader)
	}
}
