package main

import rl "github.com/gen2brain/raylib-go/raylib"

func alturaSuperiorNaMalha(
	dados DadosMalhaTerreno,
	coord CoordRegiaoTerreno,
	x, z float32,
) (float32, bool) {
	x -= float32(coord.X) * tamanhoRegiaoXZ
	z -= float32(coord.Z) * tamanhoRegiaoXZ

	vertices := dados.Vertices
	const tolerancia float32 = 0.0001

	encontrou := false
	var maiorY float32

	// Cada triângulo ocupa 9 números: X,Y,Z de seus 3 vértices.
	for i := 0; i+8 < len(vertices); i += 9 {
		x0, z0 := vertices[i], vertices[i+2]
		x1, z1 := vertices[i+3], vertices[i+5]
		x2, z2 := vertices[i+6], vertices[i+8]

		denominador := (z1-z2)*(x0-x2) + (x2-x1)*(z0-z2)
		if denominador > -0.000001 && denominador < 0.000001 {
			continue // Face vertical: não forma chão sob este ponto.
		}

		peso0 := ((z1-z2)*(x-x2) + (x2-x1)*(z-z2)) / denominador
		peso1 := ((z2-z0)*(x-x2) + (x0-x2)*(z-z2)) / denominador
		peso2 := 1 - peso0 - peso1

		if peso0 < -tolerancia ||
			peso1 < -tolerancia ||
			peso2 < -tolerancia {
			continue
		}

		yLocal := peso0*vertices[i+1] +
			peso1*vertices[i+4] +
			peso2*vertices[i+7]
		yGlobal := yLocal + float32(coord.Y)*tamanhoRegiaoY

		if !encontrou || yGlobal > maiorY {
			maiorY = yGlobal
			encontrou = true
		}
	}

	return maiorY, encontrou
}

func (mundo *Mundo) prepararMalhasPropostas(
	plano map[CoordAmostraTerreno]float32,
) map[CoordRegiaoTerreno]DadosMalhaTerreno {
	malhas := make(map[CoordRegiaoTerreno]DadosMalhaTerreno)

	// As normais da malha também devem consultar as densidades propostas.
	amostrarProposto := func(x, y, z int) float32 {
		chave := CoordAmostraTerreno{X: x, Y: y, Z: z}
		if densidade, alterada := plano[chave]; alterada {
			return densidade
		}
		return mundo.densidadeGlobal(x, y, z)
	}

	for coord, regiao := range mundo.regioes {
		campoProposto := regiao.Campo
		alterada := false

		for chave, densidade := range plano {
			ix, iy, iz, pertence := regiao.indicesDaAmostraGlobal(
				chave.X, chave.Y, chave.Z,
			)
			if pertence {
				campoProposto[ix][iy][iz] = densidade
				alterada = true
			}
		}

		if alterada {
			malhas[coord] = prepararDadosMalhaTerreno(
				&campoProposto, coord, amostrarProposto,
			)
		}
	}

	return malhas
}

func alturaNasMalhasPropostas(
	malhas map[CoordRegiaoTerreno]DadosMalhaTerreno,
	x, z float32,
) (float32, bool) {
	encontrou := false
	var maiorY float32

	for coord, dados := range malhas {
		area := coord.areaXZ()
		if x < area.MinX || x > area.MaxX ||
			z < area.MinZ || z > area.MaxZ {
			continue
		}

		y, ok := alturaSuperiorNaMalha(dados, coord, x, z)
		if ok && (!encontrou || y > maiorY) {
			maiorY = y
			encontrou = true
		}
	}

	return maiorY, encontrou
}

func malhaPropostaCobrePonto(
	malhas map[CoordRegiaoTerreno]DadosMalhaTerreno,
	x, z float32,
) bool {
	for coord := range malhas {
		area := coord.areaXZ()
		if x >= area.MinX && x <= area.MaxX &&
			z >= area.MinZ && z <= area.MaxZ {
			return true
		}
	}
	return false
}

func (mundo *Mundo) pontoPerdeApoio(
	malhas map[CoordRegiaoTerreno]DadosMalhaTerreno,
	x, z float32,
) bool {
	// Fora das regiões modificadas, o chão permanece igual.
	if !malhaPropostaCobrePonto(malhas, x, z) {
		return false
	}

	raio := rl.NewRay(
		rl.NewVector3(x, 5, z),
		rl.NewVector3(0, -1, 0),
	)
	antes := mundo.colisaoTerreno(raio)
	if !antes.Hit {
		// Um objeto já sem chão em um save antigo não deve bloquear
		// um corte que não criou esse problema.
		return false
	}

	depois, encontrou := alturaNasMalhasPropostas(malhas, x, z)
	if !encontrou {
		return true
	}

	const quedaMaxima float32 = 0.15
	return antes.Point.Y-depois > quedaMaxima
}

func (mundo *Mundo) motivoBloqueioCorte(
	malhas map[CoordRegiaoTerreno]DadosMalhaTerreno,
	ponto rl.Vector3,
	raio float32,
) string {
	areaBusca := areaAoRedor(ponto, raio+1)

	for _, objeto := range mundo.ocupacao.consultar(areaBusca) {
		switch objeto.Tipo {
		case TipoArvore:
			if mundo.pontoPerdeApoio(
				malhas, objeto.Posicao.X, objeto.Posicao.Z,
			) {
				return "uma árvore perderia o apoio"
			}

		case TipoRocha:
			if mundo.pontoPerdeApoio(
				malhas, objeto.Posicao.X, objeto.Posicao.Z,
			) {
				return "uma pedra perderia o apoio"
			}

		case TipoPiso:
			for _, apoio := range pontosApoioPiso(objeto.Area) {
				if mundo.pontoPerdeApoio(
					malhas, apoio[0], apoio[1],
				) {
					return "um piso perderia o apoio"
				}
			}

		case TipoCubo:
			if mundo.pontoPerdeApoio(
				malhas, objeto.Posicao.X, objeto.Posicao.Z,
			) {
				return "um cubo perderia o apoio"
			}
		}

	}

	return ""
}

func (mundo *Mundo) gramasAfetadasPorCorte(
	malhas map[CoordRegiaoTerreno]DadosMalhaTerreno,
	ponto rl.Vector3,
	raio float32,
) map[IDObjeto]struct{} {
	ids := make(map[IDObjeto]struct{})

	for _, objeto := range mundo.ocupacao.consultar(areaAoRedor(ponto, raio+1)) {
		if objeto.Tipo != TipoGrama {
			continue
		}

		if mundo.pontoPerdeApoio(malhas, objeto.Posicao.X, objeto.Posicao.Z) {
			ids[objeto.ID] = struct{}{}
		}
	}

	return ids
}
