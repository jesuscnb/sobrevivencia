package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const escalaConstrucaoTeste float32 = 2

// Jogo reúne o estado que precisa sobreviver de um quadro para o próximo.
type Jogo struct {
	camera           rl.Camera3D
	mundo            Mundo
	catalogo         Catalogo
	cubosColetados   int
	xRay             bool
	inventario       Inventario
	vida             Vida
	modoConstrucao   bool
	previaConstrucao PreviaConstrucao
	tipoConstrucao   TipoObjeto
	catalogoInimigos CatalogoInimigos
}

func novoJogo(semente int) *Jogo {
	ruido := novoRuido(semente)

	jogo := &Jogo{}
	jogo.vida = novaVida(100)
	jogo.tipoConstrucao = TipoPiso

	jogo.camera = rl.NewCamera3D(
		rl.NewVector3(4, 3, 4),
		rl.NewVector3(3, 3, 3),
		rl.NewVector3(0, 1, 0),
		55,
		rl.CameraPerspective,
	)

	var err error
	jogo.catalogo, err = carregarCatalogo()
	if err != nil {
		panic(err)
	}

	jogo.catalogoInimigos, err = carregarCatalogoInimigos()
	if err != nil {
		jogo.catalogo.descarregar()
		panic(err)
	}

	jogo.mundo = novoMundo(ruido, semente, &jogo.catalogo)

	if !jogo.mundo.iniciarInimigo(jogo.camera.Position) {
		fmt.Println("Nenhum ponto livre para o inimigo")
	}

	return jogo
}

func gameLoop() {
	rl.SetTargetFPS(60)
	jogo := novoJogo(54321)

	for !rl.WindowShouldClose() {

		if rl.IsKeyPressed(rl.KeyF9) {
			estado, err := lerPartida()
			if err == nil {
				err = restaurarPartida(jogo, estado)
			}
			if err != nil {
				fmt.Println("Erro ao carregar partida:", err)
			} else {
				fmt.Println("Partida carregada")
			}
		}

		if rl.IsKeyPressed(rl.KeyF1) {
			jogo.vida.ReceberDano(10)
		}

		if jogo.vida.Atual == 0 {

			if rl.IsKeyPressed(rl.KeyR) {
				jogo.vida = novaVida(jogo.vida.Maxima)
			}
			jogo.previaConstrucao = PreviaConstrucao{}
			desenharJogo(jogo, ResultadoMira{})
			continue
		}

		atualizarJogador(jogo)

		if rl.IsKeyPressed(rl.KeyF2) && len(jogo.mundo.inimigos) > 0 {
			grade := novaGradeNavegacao()
			caminho, encontrado := grade.encontrarCaminho(
				&jogo.mundo,
				jogo.mundo.inimigos[0].Posicao,
				jogo.camera.Position,
			)
			origem := jogo.mundo.inimigos[0].Posicao
			destino := jogo.camera.Position

			inicio, inicioOK := grade.celulaDaPosicao(origem)
			fim, fimOK := grade.celulaDaPosicao(destino)

			fmt.Printf(
				"Inimigo=(%.2f, %.2f) célula=%+v válida=%t | Jogador=(%.2f, %.2f) célula=%+v válida=%t\n",
				origem.X, origem.Z, inicio, inicioOK,
				destino.X, destino.Z, fim, fimOK,
			)

			if inicioOK {
				for _, vizinho := range grade.vizinhosDaCelula(inicio) {
					_, livre := grade.passoParaCelula(&jogo.mundo, origem, vizinho)
					fmt.Printf("Vizinho %+v: livre=%t\n", vizinho, livre)
				}
			}

			fmt.Printf("Caminho encontrado: %t | Células: %d\n", encontrado, len(caminho))
			anterior := jogo.mundo.inimigos[0].Posicao
			for i, ponto := range caminho {
				livre := jogo.mundo.segmentoLivreInimigo(anterior, ponto)
				fmt.Printf(
					"Trecho %d: (%.2f, %.2f) -> (%.2f, %.2f) | livre=%t\n",
					i,
					anterior.X, anterior.Z,
					ponto.X, ponto.Z,
					livre,
				)
				anterior = ponto
			}
		}

		dano := jogo.mundo.atualizarInimigos(
			rl.GetFrameTime(),
			jogo.camera.Position,
		)
		jogo.vida.ReceberDano(dano)

		mira := calcularMira(jogo)

		if rl.IsKeyPressed(rl.KeyB) {
			jogo.modoConstrucao = !jogo.modoConstrucao
		}

		if jogo.modoConstrucao {
			if rl.IsKeyPressed(rl.KeyOne) {
				jogo.tipoConstrucao = TipoPiso
			}
			if rl.IsKeyPressed(rl.KeyTwo) {
				jogo.tipoConstrucao = TipoParede
			}
			if rl.IsKeyPressed(rl.KeyThree) {
				jogo.tipoConstrucao = TipoTelhado
			}
		}

		switch jogo.tipoConstrucao {
		case TipoPiso:
			jogo.previaConstrucao = calcularPreviaPiso(jogo, mira)
		case TipoParede:
			jogo.previaConstrucao = calcularPreviaParede(jogo, mira)
		case TipoTelhado:
			jogo.previaConstrucao = calcularPreviaTelhado(jogo, mira)
		}

		if jogo.modoConstrucao && rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			switch jogo.tipoConstrucao {
			case TipoPiso:
				if tentarConstruirPiso(jogo) {
					jogo.previaConstrucao = calcularPreviaPiso(jogo, mira)
				}
			case TipoParede:
				if tentarConstruirParede(jogo) {
					jogo.previaConstrucao = calcularPreviaParede(jogo, mira)
				}
			case TipoTelhado:
				if tentarConstruirTelhado(jogo) {
					jogo.previaConstrucao = calcularPreviaTelhado(jogo, mira)
				}
			}
		}

		if rl.IsKeyPressed(rl.KeyE) {
			coletarAlvo(jogo, mira)
		}

		if mira.terrenoAoAlcance && rl.IsKeyPressed(rl.KeyC) {
			ponto := mira.colisaoTerreno.Point
			cortarEsfera(&jogo.mundo.campo, ponto.X, ponto.Y, ponto.Z, 2)
			novoModelo := criarModeloTerreno(&jogo.mundo.campo)
			rl.UnloadModel(jogo.mundo.modeloTerreno)
			jogo.mundo.modeloTerreno = novoModelo
		}

		if rl.IsKeyPressed(rl.KeyM) {
			jogo.xRay = !jogo.xRay
		}

		if rl.IsKeyPressed(rl.KeyF5) {
			estado, err := capturarPartida(jogo)
			if err == nil {
				var caminho string
				caminho, err = gravarPartida(estado)
				if err == nil {
					fmt.Println("Partida salva em:", caminho)
				}
			}
			if err != nil {
				fmt.Println("Erro ao salvar partida:", err)
			}
		}

		desenharJogo(jogo, mira)
	}

	rl.UnloadModel(jogo.mundo.modeloTerreno)
	jogo.catalogo.descarregar()
	jogo.catalogoInimigos.descarregar()

}
