package main

import rl "github.com/gen2brain/raylib-go/raylib"

func atualizarJogador(jogo *Jogo) {
	posicaoAnterior := jogo.camera.Position
	rl.UpdateCamera(&jogo.camera, rl.CameraFirstPerson)
	jogo.camera.Up = rl.NewVector3(0, 1, 0)

	centroCorpo := rl.NewVector3(
		jogo.camera.Position.X,
		jogo.camera.Position.Y-1.1,
		jogo.camera.Position.Z,
	)

	const raioJogador = float32(0.35)
	const raioTronco = float32(0.12) * escalaArvore

	bloqueado := false

	for i := range jogo.mundo.cubos {
		if !jogo.mundo.cubos[i].Existe {
			continue
		}

		if rl.CheckCollisionBoxSphere(
			caixaDoCubo(jogo.mundo.cubos[i].Posicao),
			centroCorpo,
			raioJogador,
		) {
			bloqueado = true
			break
		}
	}

	if !bloqueado {
		raioColisao := raioJogador + raioTronco

		for _, arvore := range jogo.mundo.arvores {
			if !arvore.Existe {
				continue
			}

			if distanciaHorizontal2(centroCorpo, arvore.Posicao) < raioColisao*raioColisao {
				bloqueado = true
				break
			}
		}
	}

	if !bloqueado {
		for _, rocha := range jogo.mundo.rochas {
			if !rocha.Existe {
				continue
			}

			if rl.CheckCollisionBoxSphere(
				caixaDaRocha(rocha.Posicao, jogo.mundo.limitesRocha),
				centroCorpo,
				raioJogador,
			) {
				bloqueado = true
				break
			}
		}
	}

	if bloqueado {
		diferencaX := posicaoAnterior.X - jogo.camera.Position.X
		diferencaZ := posicaoAnterior.Z - jogo.camera.Position.Z

		jogo.camera.Position.X = posicaoAnterior.X
		jogo.camera.Position.Z = posicaoAnterior.Z
		jogo.camera.Target.X += diferencaX
		jogo.camera.Target.Z += diferencaZ
	}

	chao := jogo.mundo.chaoCarregado(
		jogo.camera.Position.X,
		jogo.camera.Position.Z,
		raioJogador,
	)

	if !chao.Hit {
		// A posição tentada não tem terreno pronto ou não tem chão.
		diferencaX := posicaoAnterior.X - jogo.camera.Position.X
		diferencaZ := posicaoAnterior.Z - jogo.camera.Position.Z

		jogo.camera.Position.X = posicaoAnterior.X
		jogo.camera.Position.Z = posicaoAnterior.Z
		jogo.camera.Target.X += diferencaX
		jogo.camera.Target.Z += diferencaZ

		chao = jogo.mundo.chaoCarregado(
			posicaoAnterior.X,
			posicaoAnterior.Z,
			raioJogador,
		)
	}

	if chao.Hit {
		alturaDosOlhos := chao.Point.Y + 1.7
		diferenca := alturaDosOlhos - jogo.camera.Position.Y
		jogo.camera.Position.Y += diferenca
		jogo.camera.Target.Y += diferenca
	}
	jogo.som.atualizarPassos(&jogo.mundo, posicaoAnterior, jogo.camera.Position)

}
