package main

func coletarAlvo(jogo *Jogo, mira ResultadoMira) {
	switch mira.tipoAlvo {
	case AlvoCubo:
		jogo.mundo.cubos[mira.indiceAlvo].Existe = false
		jogo.mundo.ocupacao.remover(jogo.mundo.cubos[mira.indiceAlvo].ID)
		jogo.cubosColetados++

	case AlvoArvore:
		jogo.mundo.arvores[mira.indiceAlvo].Existe = false
		jogo.mundo.ocupacao.remover(jogo.mundo.arvores[mira.indiceAlvo].ID)
		jogo.inventario.Madeira++

	case AlvoRocha:
		jogo.mundo.rochas[mira.indiceAlvo].Existe = false
		jogo.mundo.ocupacao.remover(jogo.mundo.rochas[mira.indiceAlvo].ID)
		jogo.inventario.Pedra++
	}
}
