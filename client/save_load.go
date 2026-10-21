package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Aplica ao mundo recém-gerado os recursos que não existem mais no save.
// Retorna quantos cubos foram coletados para restaurar o contador da tela.
func aplicarRecursosRemovidos(mundo *Mundo, recursos []RecursoRemovidoSalvo) (int, error) {
	vistos := make(map[IDObjeto]struct{}, len(recursos))
	cubosColetados := 0

	for _, salvo := range recursos {
		id := IDObjeto{
			RegiaoX: salvo.RegiaoX,
			RegiaoZ: salvo.RegiaoZ,
			Ordem:   salvo.Ordem,
		}

		switch salvo.Tipo {
		case "cubo":
			id.Tipo = TipoCubo
		case "arvore":
			id.Tipo = TipoArvore
		case "rocha":
			id.Tipo = TipoRocha
		default:
			return 0, fmt.Errorf("tipo de recurso desconhecido: %q", salvo.Tipo)
		}

		if _, repetido := vistos[id]; repetido {
			return 0, fmt.Errorf("recurso repetido no save: %+v", id)
		}
		vistos[id] = struct{}{}

		encontrado := false

		switch id.Tipo {
		case TipoCubo:
			for i := range mundo.cubos {
				if mundo.cubos[i].ID == id {
					mundo.cubos[i].Existe = false
					cubosColetados++
					encontrado = true
					break
				}
			}

		case TipoArvore:
			for i := range mundo.arvores {
				if mundo.arvores[i].ID == id {
					mundo.arvores[i].Existe = false
					encontrado = true
					break
				}
			}

		case TipoRocha:
			for i := range mundo.rochas {
				if mundo.rochas[i].ID == id {
					mundo.rochas[i].Existe = false
					encontrado = true
					break
				}
			}
		}

		if !encontrado {
			return 0, fmt.Errorf("recurso do save não encontrado no mundo: %+v", id)
		}

		mundo.ocupacao.remover(id)
	}

	return cubosColetados, nil
}

func aplicarConstrucoes(mundo *Mundo, catalogo *Catalogo, construcoes []PecaSalva) error {
	for indice, salva := range construcoes {
		var tipo TipoObjeto

		switch salva.Tipo {
		case "piso":
			tipo = TipoPiso
		case "parede":
			tipo = TipoParede
		case "telhado":
			tipo = TipoTelhado
		default:
			return fmt.Errorf("construção %d tem tipo desconhecido: %q", indice, salva.Tipo)
		}

		modelo, existe := catalogo.modelo(tipo)
		if !existe {
			return fmt.Errorf("modelo da construção %d não encontrado", indice)
		}
		if salva.Escala != catalogo.escala(tipo) {
			return fmt.Errorf("construção %d tem escala incompatível", indice)
		}

		const limitePosicao float32 = 16
		p := salva.Posicao

		if p.X < -limitePosicao || p.X > limitePosicao ||
			p.Y < -limitePosicao || p.Y > limitePosicao ||
			p.Z < -limitePosicao || p.Z > limitePosicao {
			return fmt.Errorf("construção %d está fora do mundo atual", indice)
		}

		posicao := rl.NewVector3(
			salva.Posicao.X,
			salva.Posicao.Y,
			salva.Posicao.Z,
		)

		mundo.adicionarPeca(tipo, posicao, salva.Escala, salva.Rotacao, modelo)
	}

	return nil
}

func restaurarPartida(jogo *Jogo, estado PartidaSalva) error {
	if estado.Jogador.VidaMaxima <= 0 ||
		estado.Jogador.VidaAtual < 0 ||
		estado.Jogador.VidaAtual > estado.Jogador.VidaMaxima {
		return fmt.Errorf("vida inválida no save")
	}

	if estado.Inventario.Madeira < 0 || estado.Inventario.Pedra < 0 {
		return fmt.Errorf("inventário inválido no save")
	}

	mundoNovo := novoMundo(
		novoRuido(estado.Geracao.Semente),
		estado.Geracao.Semente,
		&jogo.catalogo,
	)

	err := aplicarAlteracoesTerreno(&mundoNovo, estado.AlteracoesTerreno)
	cubosColetados := 0

	if err == nil {
		cubosColetados, err = aplicarRecursosRemovidos(
			&mundoNovo, estado.RecursosRemovidos,
		)
	}

	if err == nil {
		err = aplicarConstrucoes(&mundoNovo, &jogo.catalogo, estado.Construcoes)
	}

	if err == nil && estado.VersaoFormato >= 3 {
		err = aplicarInimigosSalvos(&mundoNovo, estado.Inimigos)
	}

	if err != nil {
		rl.UnloadModel(mundoNovo.modeloTerreno)
		return err
	}

	if estado.VersaoFormato < 3 {
		posicaoJogador := rl.NewVector3(
			estado.Jogador.Posicao.X,
			estado.Jogador.Posicao.Y,
			estado.Jogador.Posicao.Z,
		)
		if !mundoNovo.iniciarInimigo(posicaoJogador) {
			fmt.Println("Nenhum ponto livre para o inimigo")
		}
	}

	modeloAntigo := jogo.mundo.modeloTerreno
	jogo.mundo = mundoNovo
	jogo.cubosColetados = cubosColetados
	jogo.inventario = Inventario{
		Madeira: estado.Inventario.Madeira,
		Pedra:   estado.Inventario.Pedra,
	}
	jogo.vida = Vida{
		Atual:  estado.Jogador.VidaAtual,
		Maxima: estado.Jogador.VidaMaxima,
	}
	jogo.camera.Position = rl.NewVector3(
		estado.Jogador.Posicao.X,
		estado.Jogador.Posicao.Y,
		estado.Jogador.Posicao.Z,
	)
	jogo.camera.Target = rl.NewVector3(
		estado.Jogador.Alvo.X,
		estado.Jogador.Alvo.Y,
		estado.Jogador.Alvo.Z,
	)
	jogo.modoConstrucao = false
	jogo.previaConstrucao = PreviaConstrucao{}
	jogo.tipoConstrucao = TipoPiso

	rl.UnloadModel(modeloAntigo)
	return nil
}

func aplicarAlteracoesTerreno(mundo *Mundo, alteracoes []AmostraTerrenoSalva) error {
	if len(alteracoes) == 0 {
		return nil
	}

	vistos := make(map[[3]int]struct{}, len(alteracoes))

	// Validamos a lista inteira antes de modificar o campo.
	for _, amostra := range alteracoes {
		if amostra.X < -8 || amostra.X > 8 ||
			amostra.Y < -4 || amostra.Y > 4 ||
			amostra.Z < -8 || amostra.Z > 8 {
			return fmt.Errorf("amostra do terreno fora dos limites: %+v", amostra)
		}

		chave := [3]int{amostra.X, amostra.Y, amostra.Z}
		if _, repetida := vistos[chave]; repetida {
			return fmt.Errorf("amostra do terreno repetida: %v", chave)
		}
		vistos[chave] = struct{}{}

		densidade := float64(amostra.Densidade)
		if math.IsNaN(densidade) || math.IsInf(densidade, 0) {
			return fmt.Errorf("densidade inválida na amostra %v", chave)
		}
	}

	for _, amostra := range alteracoes {
		mundo.campo[amostra.X+8][amostra.Y+4][amostra.Z+8] = amostra.Densidade
	}

	// Reconstruímos a malha uma vez, depois de aplicar todas as mudanças.
	modeloNovo := criarModeloTerreno(&mundo.campo)
	rl.UnloadModel(mundo.modeloTerreno)
	mundo.modeloTerreno = modeloNovo
	return nil
}

func aplicarInimigosSalvos(mundo *Mundo, salvos []InimigoSalvo) error {
	vistos := make(map[IDObjeto]struct{}, len(salvos))

	for _, salvo := range salvos {

		if salvo.Estado > EstadoAtaque ||
			salvo.Visual > VisualGhost ||
			salvo.Vida < 0 || salvo.Vida > 30 ||
			salvo.DirecaoPatrulha > 3 {
			return fmt.Errorf("estado inválido do inimigo no save")
		}

		recarga := float64(salvo.RecargaAtaque)
		if math.IsNaN(recarga) || math.IsInf(recarga, 0) ||
			recarga < 0 || recarga > float64(intervaloAtaqueInimigo) {
			return fmt.Errorf("recarga de ataque inválida no save")
		}

		id := IDObjeto{
			Tipo:    TipoInimigo,
			RegiaoX: salvo.RegiaoX,
			RegiaoZ: salvo.RegiaoZ,
			Ordem:   salvo.Ordem,
		}
		if _, existe := vistos[id]; existe {
			return fmt.Errorf("inimigo repetido no save: %+v", id)
		}
		vistos[id] = struct{}{}

		mundo.inimigos = append(mundo.inimigos, Inimigo{
			ID: id,
			Posicao: rl.NewVector3(
				salvo.Posicao.X, salvo.Posicao.Y, salvo.Posicao.Z,
			),
			Origem: rl.NewVector3(
				salvo.Origem.X, salvo.Origem.Y, salvo.Origem.Z,
			),
			Estado:          salvo.Estado,
			Vida:            salvo.Vida,
			DirecaoPatrulha: salvo.DirecaoPatrulha,
			RecargaAtaque:   salvo.RecargaAtaque,
			Visual:          salvo.Visual,
		})
	}

	return nil
}
