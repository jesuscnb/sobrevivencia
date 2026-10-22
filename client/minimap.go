package main

import (
	"fmt"
	"image/color"
	"math"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	tamanhoCelulaExplorada = 4
	tamanhoMinimapa        = 192
	escalaMinimapa         = 4 // pixels de tela por unidade do mundo
	raioDescoberta         = 8 // unidades ao redor do jogador
)

// A exploração usa coordenadas globais. Ela continua existindo após o descarregamento
// de uma região, enquanto as texturas são apenas um cache dos tiles próximos.
type CelulaExplorada struct{ X, Z int32 }

type Minimapa struct {
	exploradas map[CelulaExplorada]struct{}
	tiles      map[CoordRegiaoTerreno]rl.Texture2D
}

func novoMinimapa() Minimapa {
	return Minimapa{
		exploradas: make(map[CelulaExplorada]struct{}),
		tiles:      make(map[CoordRegiaoTerreno]rl.Texture2D),
	}
}

func coordenadaExploracao(x, z float32) CelulaExplorada {
	return CelulaExplorada{
		X: int32(math.Floor(float64(x) / tamanhoCelulaExplorada)),
		Z: int32(math.Floor(float64(z) / tamanhoCelulaExplorada)),
	}
}

func (mapa *Minimapa) descobrir(x, z float32) {
	centro := coordenadaExploracao(x, z)
	for dz := int32(-2); dz <= 2; dz++ {
		for dx := int32(-2); dx <= 2; dx++ {
			coord := CelulaExplorada{X: centro.X + dx, Z: centro.Z + dz}
			cx := (float32(coord.X) + 0.5) * tamanhoCelulaExplorada
			cz := (float32(coord.Z) + 0.5) * tamanhoCelulaExplorada
			dxMundo, dzMundo := cx-x, cz-z
			if dxMundo*dxMundo+dzMundo*dzMundo <= raioDescoberta*raioDescoberta {
				mapa.exploradas[coord] = struct{}{}
			}
		}
	}
	mapa.exploradas[centro] = struct{}{}
}

func (mapa *Minimapa) descarregar() {
	for coord, tile := range mapa.tiles {
		rl.UnloadTexture(tile)
		delete(mapa.tiles, coord)
	}
}

func (mapa *Minimapa) invalidarArea(x, z, raio float32) {
	min := coordenadaRegiao(x-raio, 0, z-raio)
	max := coordenadaRegiao(x+raio, 0, z+raio)
	for rz := min.Z; rz <= max.Z; rz++ {
		for rx := min.X; rx <= max.X; rx++ {
			coord := CoordRegiaoTerreno{X: rx, Z: rz}
			if tile, ok := mapa.tiles[coord]; ok {
				rl.UnloadTexture(tile)
				delete(mapa.tiles, coord)
			}
		}
	}
}

func (mapa *Minimapa) criarTile(mundo *Mundo, coord CoordRegiaoTerreno) rl.Texture2D {
	area := coord.areaXZ()
	pixels := make([]byte, tamanhoRegiaoXZ*tamanhoRegiaoXZ*4)
	trechos := mundo.estradaInicial.trechosNaArea(area, alcanceEstrada+1)
	limiteEstrada := meiaLarguraEstrada * meiaLarguraEstrada
	for iz := 0; iz < tamanhoRegiaoXZ; iz++ {
		for ix := 0; ix < tamanhoRegiaoXZ; ix++ {
			x := area.MinX + float32(ix) + 0.5
			z := area.MinZ + float32(iz) + 0.5
			altura := densidadeNoPonto(mundo.ruido, x, 0, z)
			cor := color.RGBA{R: 66, G: 129, B: 75, A: 255}
			if altura > 0.7 {
				cor = color.RGBA{R: 92, G: 148, B: 91, A: 255}
			} else if altura < -0.7 {
				cor = color.RGBA{R: 49, G: 103, B: 64, A: 255}
			}
			// Um corte que retirou a superfície em Y=0 aparece como solo exposto.
			if mundo.densidadeGlobal(int(math.Floor(float64(x))), 0,
				int(math.Floor(float64(z)))) < -0.3 && altura > -0.3 {
				cor = color.RGBA{R: 111, G: 91, B: 67, A: 255}
			}
			for _, trecho := range trechos {
				if distanciaQuadradaPontoTrecho(x, z, trecho[0], trecho[1]) <= limiteEstrada {
					cor = color.RGBA{R: 169, G: 108, B: 65, A: 255}
					break
				}
			}
			indice := (iz*tamanhoRegiaoXZ + ix) * 4
			pixels[indice] = cor.R
			pixels[indice+1] = cor.G
			pixels[indice+2] = cor.B
			pixels[indice+3] = cor.A
		}
	}
	imagem := rl.NewImage(pixels, tamanhoRegiaoXZ, tamanhoRegiaoXZ, 1,
		rl.UncompressedR8g8b8a8)
	textura := rl.LoadTextureFromImage(imagem)
	runtime.KeepAlive(pixels)
	if textura.ID != 0 {
		rl.SetTextureFilter(textura, rl.FilterPoint)
	}
	return textura
}

// Atualiza recursos da GPU antes de BeginDrawing, somente para a janela visível.
func (mapa *Minimapa) atualizarTiles(mundo *Mundo, centroX, centroZ float32) {
	const metadeMundo = tamanhoMinimapa / (2 * escalaMinimapa)
	minRegiao := coordenadaRegiao(centroX-metadeMundo, 0, centroZ-metadeMundo)
	maxRegiao := coordenadaRegiao(centroX+metadeMundo, 0, centroZ+metadeMundo)
	visiveis := make(map[CoordRegiaoTerreno]struct{}, 16)
	for rz := minRegiao.Z; rz <= maxRegiao.Z; rz++ {
		for rx := minRegiao.X; rx <= maxRegiao.X; rx++ {
			coord := CoordRegiaoTerreno{X: rx, Z: rz}
			visiveis[coord] = struct{}{}
			if _, ok := mapa.tiles[coord]; ok {
				continue
			}
			if tile := mapa.criarTile(mundo, coord); tile.ID != 0 {
				mapa.tiles[coord] = tile
			}
		}
	}
	for coord, tile := range mapa.tiles {
		if _, ok := visiveis[coord]; !ok {
			rl.UnloadTexture(tile)
			delete(mapa.tiles, coord)
		}
	}
}

func (mapa *Minimapa) desenhar(jogo *Jogo) {
	const margem = float32(14)
	lado := float32(tamanhoMinimapa)
	xTela := float32(rl.GetScreenWidth()) - lado - margem
	zTela := margem
	if xTela < margem {
		return
	}
	centroX, centroZ := jogo.camera.Position.X, jogo.camera.Position.Z
	metadeMundo := lado / (2 * escalaMinimapa)
	minX, minZ := centroX-metadeMundo, centroZ-metadeMundo
	maxX, maxZ := centroX+metadeMundo, centroZ+metadeMundo
	rl.DrawRectangleRec(rl.NewRectangle(xTela-3, zTela-3, lado+6, lado+6),
		color.RGBA{R: 17, G: 24, B: 29, A: 255})
	rl.BeginScissorMode(int32(xTela), int32(zTela), tamanhoMinimapa, tamanhoMinimapa)
	rl.DrawRectangleRec(rl.NewRectangle(xTela, zTela, lado, lado),
		color.RGBA{R: 32, G: 49, B: 46, A: 255})
	minRegiao := coordenadaRegiao(minX, 0, minZ)
	maxRegiao := coordenadaRegiao(maxX, 0, maxZ)
	for rz := minRegiao.Z; rz <= maxRegiao.Z; rz++ {
		for rx := minRegiao.X; rx <= maxRegiao.X; rx++ {
			coord := CoordRegiaoTerreno{X: rx, Z: rz}
			tile := mapa.tiles[coord]
			if tile.ID == 0 {
				continue
			}
			area := coord.areaXZ()
			rl.DrawTexturePro(tile,
				rl.NewRectangle(0, 0, tamanhoRegiaoXZ, tamanhoRegiaoXZ),
				rl.NewRectangle(xTela+(area.MinX-minX)*escalaMinimapa,
					zTela+(area.MinZ-minZ)*escalaMinimapa,
					tamanhoRegiaoXZ*escalaMinimapa, tamanhoRegiaoXZ*escalaMinimapa),
				rl.Vector2{}, 0, rl.White)
		}
	}
	// Escurece somente células ainda não descobertas. O recorte limita o desenho.
	minCel := coordenadaExploracao(minX, minZ)
	maxCel := coordenadaExploracao(maxX, maxZ)
	for cz := minCel.Z; cz <= maxCel.Z; cz++ {
		for cx := minCel.X; cx <= maxCel.X; cx++ {
			if _, vista := mapa.exploradas[CelulaExplorada{X: cx, Z: cz}]; vista {
				continue
			}
			rl.DrawRectangleRec(rl.NewRectangle(
				xTela+(float32(cx*tamanhoCelulaExplorada)-minX)*escalaMinimapa,
				zTela+(float32(cz*tamanhoCelulaExplorada)-minZ)*escalaMinimapa,
				tamanhoCelulaExplorada*escalaMinimapa,
				tamanhoCelulaExplorada*escalaMinimapa),
				color.RGBA{R: 12, G: 18, B: 24, A: 245})
		}
	}
	// Pontos de interesse só aparecem após a área ter sido explorada.
	for _, lote := range jogo.mundo.lotesIniciais {
		if !jogo.mundo.lotesDecididos[lote.ID] {
			continue
		}
		if _, vista := mapa.exploradas[coordenadaExploracao(lote.Centro.X, lote.Centro.Z)]; !vista {
			continue
		}
		mapa.marcador(minX, minZ, xTela, zTela, lote.Centro.X, lote.Centro.Z,
			color.RGBA{R: 237, G: 211, B: 129, A: 255})
	}
	for _, peca := range jogo.mundo.construcoes {
		if peca.Tipo != TipoPiso {
			continue
		}
		if _, vista := mapa.exploradas[coordenadaExploracao(peca.Posicao.X, peca.Posicao.Z)]; !vista {
			continue
		}
		mapa.marcador(minX, minZ, xTela, zTela, peca.Posicao.X, peca.Posicao.Z,
			color.RGBA{R: 239, G: 223, B: 182, A: 255})
	}
	// O ponto branco fica fixo no centro; a ponta aponta para onde a câmera olha.
	px, pz := xTela+lado/2, zTela+lado/2
	dx := jogo.camera.Target.X - centroX
	dz := jogo.camera.Target.Z - centroZ
	comprimento := float32(math.Hypot(float64(dx), float64(dz)))
	if comprimento > 0.001 {
		rl.DrawLineEx(rl.NewVector2(px, pz),
			rl.NewVector2(px+8*dx/comprimento, pz+8*dz/comprimento), 2, rl.White)
	}
	rl.DrawCircle(int32(px), int32(pz), 3, rl.White)
	rl.EndScissorMode()
	rl.DrawRectangleLinesEx(rl.NewRectangle(xTela, zTela, lado, lado), 2,
		color.RGBA{R: 220, G: 225, B: 204, A: 255})
	rl.DrawText(fmt.Sprintf("X %.0f  Z %.0f", centroX, centroZ),
		int32(xTela), int32(zTela+lado+7), 14, rl.White)
}

func (mapa *Minimapa) marcador(minX, minZ, xTela, zTela, x, z float32, cor color.RGBA) {
	px := xTela + (x-minX)*escalaMinimapa
	pz := zTela + (z-minZ)*escalaMinimapa
	rl.DrawRectangleRec(rl.NewRectangle(px-3, pz-3, 6, 6), cor)
}
