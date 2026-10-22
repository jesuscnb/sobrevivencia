package main

type CaminhoEstrada struct {
	Pontos []PontoTerritorial
}

func valorCatmullRom(p0, p1, p2, p3, t float32) float32 {
	t2 := t * t
	t3 := t2 * t

	return 0.5 * (2*p1 +
		(-p0+p2)*t +
		(2*p0-5*p1+4*p2-p3)*t2 +
		(-p0+3*p1-3*p2+p3)*t3)
}

func gerarEstradaInicial(semente int64) CaminhoEstrada {
	marcos := marcosIniciais(semente)
	const amostrasPorTrecho = 128

	pontos := make([]PontoTerritorial, 0, 4*amostrasPorTrecho+1)

	for trecho := 0; trecho < 4; trecho++ {
		p0 := marcos[(trecho+3)%4]
		p1 := marcos[trecho]
		p2 := marcos[(trecho+1)%4]
		p3 := marcos[(trecho+2)%4]

		for passo := 0; passo < amostrasPorTrecho; passo++ {
			t := float32(passo) / amostrasPorTrecho
			pontos = append(pontos, PontoTerritorial{
				X: valorCatmullRom(p0.X, p1.X, p2.X, p3.X, t),
				Z: valorCatmullRom(p0.Z, p1.Z, p2.Z, p3.Z, t),
			})
		}
	}

	pontos = append(pontos, pontos[0])
	return CaminhoEstrada{Pontos: pontos}
}

func (c CaminhoEstrada) trechosNaArea(area AreaXZ, margem float32) [][2]PontoTerritorial {
	minX := area.MinX - margem
	maxX := area.MaxX + margem
	minZ := area.MinZ - margem
	maxZ := area.MaxZ + margem

	var trechos [][2]PontoTerritorial

	for i := 0; i+1 < len(c.Pontos); i++ {
		a := c.Pontos[i]
		b := c.Pontos[i+1]

		// Descarte quando a caixa do trecho não toca a área ampliada.
		if max(a.X, b.X) < minX || min(a.X, b.X) > maxX ||
			max(a.Z, b.Z) < minZ || min(a.Z, b.Z) > maxZ {
			continue
		}

		trechos = append(trechos, [2]PontoTerritorial{a, b})
	}

	return trechos
}

func sobrepoeFaixaEstrada(
	trechos [][2]PontoTerritorial,
	x, z, raioObjeto float32,
) bool {
	limite := alcanceEstrada + raioObjeto
	limiteQuadrado := limite * limite

	for _, trecho := range trechos {
		distanciaQuadrada := distanciaQuadradaPontoTrecho(
			x, z, trecho[0], trecho[1],
		)
		if distanciaQuadrada < limiteQuadrado {
			return true
		}
	}
	return false
}
