package main

func coletarAlvo(jogo *Jogo, mira ResultadoMira) {
	switch mira.tipoAlvo {
	case AlvoCubo:
		cubo := &jogo.mundo.cubos[mira.indiceAlvo]
		cubo.Existe = false
		jogo.mundo.ocupacao.remover(cubo.ID)
		jogo.mundo.recursosRemovidos[cubo.ID] = struct{}{}
		jogo.cubosColetados++
		jogo.som.tocar(SomColetaCubo)

	case AlvoArvore:
		arvore := &jogo.mundo.arvores[mira.indiceAlvo]
		arvore.Existe = false
		jogo.mundo.ocupacao.remover(arvore.ID)
		jogo.mundo.recursosRemovidos[arvore.ID] = struct{}{}
		jogo.inventario.Madeira++
		jogo.som.tocar(SomColetaMadeira)

	case AlvoRocha:
		rocha := &jogo.mundo.rochas[mira.indiceAlvo]
		rocha.Existe = false
		jogo.mundo.ocupacao.remover(rocha.ID)
		jogo.mundo.recursosRemovidos[rocha.ID] = struct{}{}
		jogo.inventario.Pedra++
		jogo.som.tocar(SomColetaPedra)
	}
}
