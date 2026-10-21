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

	// O terreno vai de -8 a 8. Deixamos 0,5 de margem para o corpo.
	const limiteJogador float32 = 7.5

	xAnterior := jogo.camera.Position.X
	zAnterior := jogo.camera.Position.Z

	if jogo.camera.Position.X < -limiteJogador {
		jogo.camera.Position.X = -limiteJogador
	} else if jogo.camera.Position.X > limiteJogador {
		jogo.camera.Position.X = limiteJogador
	}

	if jogo.camera.Position.Z < -limiteJogador {
		jogo.camera.Position.Z = -limiteJogador
	} else if jogo.camera.Position.Z > limiteJogador {
		jogo.camera.Position.Z = limiteJogador
	}

	// Desloca o alvo junto com a câmera para preservar a direção da visão.
	jogo.camera.Target.X += jogo.camera.Position.X - xAnterior
	jogo.camera.Target.Z += jogo.camera.Position.Z - zAnterior

	malhas := jogo.mundo.modeloTerreno.GetMeshes()
	if len(malhas) > 0 {
		raioParaBaixo := rl.NewRay(
			rl.NewVector3(jogo.camera.Position.X, 5, jogo.camera.Position.Z),
			rl.NewVector3(0, -1, 0),
		)
		chao := rl.GetRayCollisionMesh(raioParaBaixo, malhas[0], jogo.mundo.modeloTerreno.Transform)
		if chao.Hit {
			alturaDosOlhos := chao.Point.Y + 1.7
			diferenca := alturaDosOlhos - jogo.camera.Position.Y
			jogo.camera.Position.Y += diferenca
			jogo.camera.Target.Y += diferenca
		}
	}
}
