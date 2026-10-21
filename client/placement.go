package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func posicaoNoTerreno(terreno, modelo rl.Model, x, z, escala, inclinacaoMinima float32) (rl.Vector3, bool) {

	malhas := terreno.GetMeshes()
	if len(malhas) == 0 {
		return rl.Vector3{}, false
	}

	raio := rl.NewRay(
		rl.NewVector3(x, 5, z),
		rl.NewVector3(0, -1, 0),
	)

	chao := rl.GetRayCollisionMesh(raio, malhas[0], terreno.Transform)

	if !chao.Hit || chao.Normal.Y < inclinacaoMinima {
		return rl.Vector3{}, false
	}

	baseDoModelo := rl.GetModelBoundingBox(modelo).Min.Y * escala
	return rl.NewVector3(x, chao.Point.Y-baseDoModelo, z), true
}

func distanciaHorizontal2(a, b rl.Vector3) float32 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return dx*dx + dz*dz
}

func areaDaCaixaXZ(caixa rl.BoundingBox) AreaXZ {
	return AreaXZ{
		MinX: caixa.Min.X,
		MaxX: caixa.Max.X,
		MinZ: caixa.Min.Z,
		MaxZ: caixa.Max.Z,
	}
}

func areasXZSobrepostas(a, b AreaXZ) bool {
	return a.MinX < b.MaxX &&
		a.MaxX > b.MinX &&
		a.MinZ < b.MaxZ &&
		a.MaxZ > b.MinZ
}

func pisoTemApoio(terreno rl.Model, area AreaXZ, alturaBase float32) bool {
	malhas := terreno.GetMeshes()
	if len(malhas) == 0 {
		return false
	}

	const margem float32 = 0.1
	const desnivelMaximo float32 = 0.35
	const inclinacaoMinima float32 = 0.85

	pontos := [5][2]float32{
		{(area.MinX + area.MaxX) / 2, (area.MinZ + area.MaxZ) / 2},
		{area.MinX + margem, area.MinZ + margem},
		{area.MinX + margem, area.MaxZ - margem},
		{area.MaxX - margem, area.MinZ + margem},
		{area.MaxX - margem, area.MaxZ - margem},
	}

	for _, ponto := range pontos {
		raio := rl.NewRay(
			rl.NewVector3(ponto[0], 5, ponto[1]),
			rl.NewVector3(0, -1, 0),
		)
		chao := rl.GetRayCollisionMesh(raio, malhas[0], terreno.Transform)

		if !chao.Hit || chao.Normal.Y < inclinacaoMinima {
			return false
		}

		desnivel := chao.Point.Y - alturaBase
		if desnivel < -desnivelMaximo || desnivel > desnivelMaximo {
			return false
		}
	}

	return true
}

func areaPecaXZ(modelo rl.Model, posicao rl.Vector3, escala, rotacao float32) AreaXZ {
	limites := rl.GetModelBoundingBox(modelo)
	angulo := float64(rotacao) * math.Pi / 180
	cosseno := float32(math.Cos(angulo))
	seno := float32(math.Sin(angulo))

	cantos := [4][2]float32{
		{limites.Min.X * escala, limites.Min.Z * escala},
		{limites.Min.X * escala, limites.Max.Z * escala},
		{limites.Max.X * escala, limites.Min.Z * escala},
		{limites.Max.X * escala, limites.Max.Z * escala},
	}

	area := AreaXZ{
		MinX: float32(math.Inf(1)),
		MaxX: float32(math.Inf(-1)),
		MinZ: float32(math.Inf(1)),
		MaxZ: float32(math.Inf(-1)),
	}

	for _, canto := range cantos {
		x := posicao.X + canto[0]*cosseno + canto[1]*seno
		z := posicao.Z - canto[0]*seno + canto[1]*cosseno

		if x < area.MinX {
			area.MinX = x
		}
		if x > area.MaxX {
			area.MaxX = x
		}
		if z < area.MinZ {
			area.MinZ = z
		}
		if z > area.MaxZ {
			area.MaxZ = z
		}
	}

	// A malha da parede é quase um plano. Damos espessura à sua área
	// para que o índice espacial possa encontrá-la com segurança.
	const meiaEspessura float32 = 0.05
	if area.MaxX-area.MinX < 2*meiaEspessura {
		centro := (area.MinX + area.MaxX) / 2
		area.MinX, area.MaxX = centro-meiaEspessura, centro+meiaEspessura
	}
	if area.MaxZ-area.MinZ < 2*meiaEspessura {
		centro := (area.MinZ + area.MaxZ) / 2
		area.MinZ, area.MaxZ = centro-meiaEspessura, centro+meiaEspessura
	}

	return area
}
