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
	MascaraVazia  rl.Texture2D
	localEstrada  int32
	localOrigem   int32
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
	material.localEstrada = rl.GetShaderLocation(material.Shader, "roadMask")
	material.localOrigem = rl.GetShaderLocation(material.Shader, "regionMinXZ")
	localSol := rl.GetShaderLocation(material.Shader, "sunDirection")
	if material.localGrama < 0 || material.localSolo < 0 ||
		material.localEstrada < 0 || material.localOrigem < 0 || localSol < 0 {
		rl.UnloadShader(material.Shader)
		panic("shader do terreno sem os controles de textura e luz")
	}
	material.Shader.UpdateLocation(rl.ShaderLocMapAlbedo, material.localGrama)
	material.Shader.UpdateLocation(rl.ShaderLocMapMetalness, material.localSolo)
	material.Shader.UpdateLocation(rl.ShaderLocMapOcclusion, material.localEstrada)
	rl.SetShaderValue(material.Shader, localSol, direcaoSol[:], rl.ShaderUniformVec3)
	material.Grama = rl.LoadTexture("assets/terrain/grass.png")
	material.Solo = rl.LoadTexture("assets/terrain/soil.png")

	if material.Grama.ID == 0 || material.Solo.ID == 0 {
		material.descarregar()
		panic("não foi possível carregar as texturas do terreno")
	}

	imagemVazia := rl.GenImageColor(1, 1, rl.Black)
	material.MascaraVazia = rl.LoadTextureFromImage(imagemVazia)
	rl.UnloadImage(imagemVazia)
	if material.MascaraVazia.ID == 0 {
		material.descarregar()
		panic("não foi possível carregar a máscara vazia da estrada")
	}

	rl.GenTextureMipmaps(&material.Grama)
	rl.GenTextureMipmaps(&material.Solo)
	rl.SetTextureFilter(material.Grama, rl.FilterTrilinear)
	rl.SetTextureFilter(material.Solo, rl.FilterTrilinear)
	rl.SetTextureWrap(material.Grama, rl.WrapRepeat)
	rl.SetTextureWrap(material.Solo, rl.WrapRepeat)

	fmt.Println("Material do terreno: shader e texturas carregados")
	return material
}

func (material *MaterialTerreno) usarNoModelo(regiao *RegiaoTerreno, texturizado bool) {
	shader := material.Padrao
	if texturizado {
		shader = material.Shader

		area := regiao.Coord.areaXZ()
		origem := []float32{area.MinX, area.MinZ}
		rl.SetShaderValue(
			material.Shader,
			material.localOrigem,
			origem,
			rl.ShaderUniformVec2,
		)
	}

	materiais := regiao.Modelo.GetMaterials()
	for i := range materiais {
		materiais[i].Shader = shader

		if texturizado {
			mascara := regiao.MascaraEstrada
			if mascara.ID == 0 {
				mascara = material.MascaraVazia
			}

			materiais[i].GetMap(rl.MapAlbedo).Texture = material.Grama
			materiais[i].GetMap(rl.MapMetalness).Texture = material.Solo
			materiais[i].GetMap(rl.MapOcclusion).Texture = mascara
		} else {
			materiais[i].GetMap(rl.MapAlbedo).Texture = material.TexturaPadrao
			materiais[i].GetMap(rl.MapOcclusion).Texture = rl.Texture2D{}
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
	if material.MascaraVazia.ID != 0 {
		rl.UnloadTexture(material.MascaraVazia)
	}
	if material.Shader.ID != 0 {
		rl.UnloadShader(material.Shader)
	}
}
