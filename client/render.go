package main

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func desenharJogo(jogo *Jogo, mira ResultadoMira) {
	modeloArvore := jogo.catalogo.modeloObrigatorio(TipoArvore)
	modeloRocha := jogo.catalogo.modeloObrigatorio(TipoRocha)
	modeloGrama := jogo.catalogo.modeloObrigatorio(TipoGrama)
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(20, 27, 39, 255))
	jogo.ceu.desenhar(jogo.camera)
	rl.BeginMode3D(jogo.camera)

	for _, regiao := range jogo.mundo.regioes {

		posicao := regiao.posicaoMundo()
		jogo.materialTerreno.usarNoModelo(regiao.Modelo, jogo.terrenoTexturizado && !jogo.xRay)

		if jogo.xRay {
			rl.DrawModelWires(
				regiao.Modelo, posicao, 1,
				color.RGBA{R: 180, G: 180, B: 180, A: 255},
			)
		} else {
			cor := rl.White
			if !jogo.terrenoTexturizado {
				cor = color.RGBA{R: 70, G: 160, B: 90, A: 255}
			}
			rl.DrawModel(regiao.Modelo, posicao, 1, cor)
		}
	}

	if mira.terrenoAoAlcance {
		rl.DrawSphere(
			mira.colisaoTerreno.Point,
			0.12,
			color.RGBA{R: 255, G: 40, B: 40, A: 255},
		)
	}

	for _, peca := range jogo.mundo.construcoes {

		if !jogo.mundo.superficieCarregada(peca.Posicao.X, peca.Posicao.Z) {
			continue
		}

		modelo := jogo.catalogo.modeloObrigatorio(peca.Tipo)
		rl.DrawModelEx(
			modelo,
			peca.Posicao,
			rl.NewVector3(0, 1, 0),
			peca.Rotacao,
			rl.NewVector3(peca.Escala, peca.Escala, peca.Escala),
			rl.White,
		)
	}

	if jogo.previaConstrucao.Visivel {
		previa := jogo.previaConstrucao
		modelo := jogo.catalogo.modeloObrigatorio(previa.Tipo)
		escala := jogo.catalogo.escala(previa.Tipo)

		cor := rl.Red
		if previa.PodeColocar {
			cor = rl.Green
		}

		rl.DrawModelWiresEx(
			modelo,
			previa.Posicao,
			rl.NewVector3(0, 1, 0),
			previa.Rotacao,
			rl.NewVector3(escala, escala, escala),
			cor,
		)
	}

	for _, cubo := range jogo.mundo.cubos {
		if cubo.Existe && jogo.mundo.superficieCarregada(cubo.Posicao.X, cubo.Posicao.Z) {
			rl.DrawCube(
				cubo.Posicao,
				1, 1, 1,
				color.RGBA{R: 0, G: 0, B: 255, A: 255},
			)
		}
	}

	for _, arvore := range jogo.mundo.arvores {
		if arvore.Existe {
			rl.DrawModel(modeloArvore, arvore.Posicao, escalaArvore, rl.White)
		}
	}

	for _, rocha := range jogo.mundo.rochas {
		if rocha.Existe {
			rl.DrawModel(modeloRocha, rocha.Posicao, escalaRocha, rl.White)
		}
	}

	for _, grama := range jogo.mundo.gramas {
		rl.DrawModelEx(
			modeloGrama,
			grama.Posicao,
			rl.NewVector3(0, 1, 0),
			grama.Rotacao,
			rl.NewVector3(grama.Escala, grama.Escala, grama.Escala),
			rl.White,
		)
	}

	for i := range jogo.mundo.inimigos {

		inimigo := &jogo.mundo.inimigos[i]

		if inimigo.Vida <= 0 || !jogo.mundo.superficieCarregada(inimigo.Posicao.X, inimigo.Posicao.Z) {
			continue
		}

		visual := jogo.catalogoInimigos.obrigatorio(inimigo.Visual)

		posicao := rl.NewVector3(
			inimigo.Posicao.X,
			inimigo.Posicao.Y+visual.AjusteY,
			inimigo.Posicao.Z,
		)

		const quadrosPorSegundoAnimacao = 30

		nomeAnimacao := "Idle"
		animacao := visual.Idle
		tempoDoGolpe := float32(0)
		mostrandoGolpe := false

		moveu := distanciaHorizontal2(
			inimigo.Posicao,
			inimigo.PosicaoAnterior,
		) > 0.000001

		// A recarga começa em 1,5 quando um golpe acontece.
		// Subtraí-la de 1,5 informa há quanto tempo esse golpe começou.
		if inimigo.Estado == EstadoAtaque && inimigo.RecargaAtaque > 0 {
			tempoDoGolpe = intervaloAtaqueInimigo - inimigo.RecargaAtaque
			duracaoPunch := float32(visual.Punch.KeyframeCount) /
				quadrosPorSegundoAnimacao

			if tempoDoGolpe >= 0 && tempoDoGolpe < duracaoPunch {
				nomeAnimacao = "Punch"
				animacao = visual.Punch
				mostrandoGolpe = true
			}
		} else if moveu {
			if inimigo.Estado == EstadoPerseguicao {
				nomeAnimacao = "Run"
				animacao = visual.Run
			} else {
				nomeAnimacao = "Walk"
				animacao = visual.Walk
			}
		}

		if inimigo.AnimacaoVisual != nomeAnimacao {
			inimigo.AnimacaoVisual = nomeAnimacao
			inimigo.TempoAnimacao = 0
		}

		quadro := int(inimigo.TempoAnimacao*quadrosPorSegundoAnimacao) %
			int(animacao.KeyframeCount)

		if mostrandoGolpe {
			// Punch é executado uma vez por ataque, sem voltar ao quadro zero.
			quadro = int(tempoDoGolpe * quadrosPorSegundoAnimacao)
			if quadro >= int(animacao.KeyframeCount) {
				quadro = int(animacao.KeyframeCount) - 1
			}
		}

		rl.UpdateModelAnimation(visual.Modelo, animacao, float32(quadro))

		dx := inimigo.Posicao.X - inimigo.PosicaoAnterior.X
		dz := inimigo.Posicao.Z - inimigo.PosicaoAnterior.Z

		if inimigo.Estado == EstadoAtaque {
			dx = jogo.camera.Position.X - inimigo.Posicao.X
			dz = jogo.camera.Position.Z - inimigo.Posicao.Z
		}

		if dx*dx+dz*dz > 0.000001 {
			inimigo.RotacaoY = anguloDaDirecao(dx, dz)
		}

		rl.DrawModelEx(
			visual.Modelo,
			posicao,
			rl.NewVector3(0, 1, 0),
			inimigo.RotacaoY+visual.CorrecaoFrente,
			rl.NewVector3(visual.Escala, visual.Escala, visual.Escala),
			rl.White,
		)

	}

	rl.EndMode3D()
	rl.DrawFPS(10, 10)
	rl.DrawCircle(
		int32(rl.GetScreenWidth()/2),
		int32(rl.GetScreenHeight()/2),
		3,
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
	)

	rl.DrawText(
		fmt.Sprintf("Vida: %d/%d", jogo.vida.Atual, jogo.vida.Maxima),
		10, 130, 20, rl.White,
	)
	switch mira.tipoAlvo {
	case AlvoCubo:
		rl.DrawText("E - Coletar Cubo", 10, 40, 20, rl.White)
	case AlvoArvore:
		rl.DrawText("E - Coletar Madeira", 10, 40, 20, rl.White)
	case AlvoRocha:
		rl.DrawText("E - Coletar Rocha", 10, 40, 20, rl.White)
	}
	rl.DrawText(
		fmt.Sprintf("Madeira: %d  Pedra: %d", jogo.inventario.Madeira, jogo.inventario.Pedra),
		10, 100, 20, rl.White,
	)
	rl.DrawText(fmt.Sprintf("Cubos: %d", jogo.cubosColetados), 10, 70, 20, rl.White)

	if jogo.vida.Atual == 0 {
		rl.DrawText("Sem vida - R para continuar o teste", 10, 160, 20, rl.Red)
	}

	if jogo.vida.Atual > 0 && jogo.modoConstrucao {
		var tipo string
		switch jogo.tipoConstrucao {
		case TipoPiso:
			tipo = "Piso"
		case TipoParede:
			tipo = "Parede"
		case TipoTelhado:
			tipo = "Telhado"
			rl.DrawText("Telhado requer 3 paredes de apoio", 10, 220, 20, rl.White)
		}

		rl.DrawText(
			fmt.Sprintf("%s: 1 madeira | Clique para construir | B para sair", tipo),
			10, 190, 20, rl.White,
		)
	}
	jogo.diagnostico.desenhar()

	rl.EndDrawing()
}
