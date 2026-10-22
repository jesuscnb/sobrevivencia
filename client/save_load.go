package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Aplica ao mundo recém-gerado os recursos que não existem mais no save.
// Retorna quantos cubos foram coletados para restaurar o contador da tela.
func aplicarRecursosRemovidos(mundo *Mundo, versao int, recursos []RecursoRemovidoSalvo) (int, error) {
	vistos := make(map[IDObjeto]struct{}, len(recursos))
	cubosColetados := 0

	for _, salvo := range recursos {
		id := IDObjeto{
			RegiaoX: salvo.RegiaoX,
			RegiaoZ: salvo.RegiaoZ,
			Ordem:   salvo.Ordem,
		}

		switch salvo.Tipo {
		case "grama":
			id.Tipo = TipoGrama
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

		case TipoGrama:
			for i := range mundo.gramas {
				if mundo.gramas[i].ID == id {
					mundo.gramas[i].Existe = false
					encontrado = true
					break
				}
			}

		}

		if !encontrado {
			// Cubos e recursos da região inicial já devem existir neste momento.
			// Em saves antigos, todos os recursos também precisam estar carregados.
			bordaPendente := mundo.versaoGerador <= versaoGeradorAnterior &&
				((id.Tipo == TipoArvore && id.Ordem >= 100) ||
					(id.Tipo == TipoGrama && id.Ordem >= 400) ||
					(id.Tipo == TipoRocha && id.Ordem >= 200))
			if versao < 5 || id.Tipo == TipoCubo ||
				(id.RegiaoX == 0 && id.RegiaoZ == 0 && !bordaPendente) {
				return 0, fmt.Errorf("recurso do save não encontrado no mundo: %+v", id)
			}
		} else {
			mundo.ocupacao.remover(id)
		}

		mundo.recursosRemovidos[id] = struct{}{}
	}

	return cubosColetados, nil
}

func aplicarConstrucoes(mundo *Mundo, catalogo *Catalogo, versao int, construcoes []PecaSalva) error {
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

		p := salva.Posicao

		for _, valor := range [3]float32{p.X, p.Y, p.Z} {
			if math.IsNaN(float64(valor)) || math.IsInf(float64(valor), 0) {
				return fmt.Errorf("construção %d tem posição inválida", indice)
			}
		}

		const limitePosicao float32 = 16
		if p.Y < -limitePosicao || p.Y > limitePosicao {
			return fmt.Errorf("construção %d tem altura inválida", indice)
		}

		if versao <= 4 {
			// Preserva a validação dos saves antigos.
			if p.X < -limitePosicao || p.X > limitePosicao ||
				p.Z < -limitePosicao || p.Z > limitePosicao {
				return fmt.Errorf("construção %d está fora do mundo antigo", indice)
			}
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
	minimapaNovo := novoMinimapa()
	for _, salva := range estado.CelulasExploradas {
		coord := CelulaExplorada{X: salva.X, Z: salva.Z}
		if _, repetida := minimapaNovo.exploradas[coord]; repetida {
			return fmt.Errorf("célula explorada repetida no save: %+v", coord)
		}
		minimapaNovo.exploradas[coord] = struct{}{}
	}
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
		estado.Geracao.VersaoGerador,
		&jogo.catalogo,
	)

	err := aplicarAlteracoesTerreno(
		&mundoNovo,
		estado.VersaoFormato,
		estado.AlteracoesTerreno,
	)

	cubosColetados := 0

	if err == nil {
		cubosColetados, err = aplicarRecursosRemovidos(
			&mundoNovo,
			estado.VersaoFormato,
			estado.RecursosRemovidos,
		)
	}

	if err == nil {
		err = aplicarConstrucoes(&mundoNovo, &jogo.catalogo, estado.VersaoFormato, estado.Construcoes)
	}
	if err == nil {
		if estado.Geracao.VersaoAssentamento == 0 && len(estado.LotesDecididos) != 0 {
			err = fmt.Errorf("save sem assentamento contém decisões de lotes")
		}
		for _, salvo := range estado.LotesDecididos {
			id := IDLote{Plano: CoordPlanoTerritorial{X: salvo.PlanoX, Z: salvo.PlanoZ}, Ordem: salvo.Ordem}
			if id.Plano != (CoordPlanoTerritorial{}) || int(id.Ordem) >= len(mundoNovo.lotesIniciais) {
				err = fmt.Errorf("lote desconhecido no save: %+v", id)
				break
			}
			if _, repetido := mundoNovo.lotesDecididos[id]; repetido {
				err = fmt.Errorf("lote repetido no save: %+v", id)
				break
			}
			mundoNovo.lotesDecididos[id] = salvo.Aceito
		}
		if err == nil {
			mundoNovo.versaoAssentamento = estado.Geracao.VersaoAssentamento
		}
	}

	if err == nil {
		mundoNovo.carregarBordaVegetacaoRegiao(CoordRegiaoTerreno{}, &jogo.catalogo)
		mundoNovo.carregarBordaVegetacaoRegiao(CoordRegiaoTerreno{X: 1}, &jogo.catalogo)
		mundoNovo.atualizarAssentamento(&jogo.catalogo)
	}

	if err == nil && estado.VersaoFormato >= 3 {
		err = aplicarInimigosSalvos(&mundoNovo, estado.Inimigos)
	}

	if err != nil {
		mundoNovo.descarregarTerreno()
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

	mundoAntigo := jogo.mundo
	jogo.minimapa.descarregar()
	jogo.minimapa = minimapaNovo
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

	mundoAntigo.descarregarTerreno()
	return nil
}

func aplicarAlteracoesTerreno(
	mundo *Mundo,
	versao int,
	alteracoes []AmostraTerrenoSalva,
) error {
	vistos := make(map[CoordAmostraTerreno]struct{}, len(alteracoes))

	// Valida tudo antes de modificar qualquer região.
	for _, amostra := range alteracoes {
		if versao <= 4 &&
			(amostra.X < -8 || amostra.X > 8 ||
				amostra.Y < -4 || amostra.Y > 4 ||
				amostra.Z < -8 || amostra.Z > 8) {
			return fmt.Errorf("amostra do terreno fora dos limites do save antigo: %+v", amostra)
		}

		chave := CoordAmostraTerreno{X: amostra.X, Y: amostra.Y, Z: amostra.Z}
		if _, repetida := vistos[chave]; repetida {
			return fmt.Errorf("amostra do terreno repetida: %v", chave)
		}
		vistos[chave] = struct{}{}

		densidade := float64(amostra.Densidade)
		if math.IsNaN(densidade) || math.IsInf(densidade, 0) {
			return fmt.Errorf("densidade inválida na amostra %v", chave)
		}

	}

	alteradas := make(map[CoordRegiaoTerreno]struct{})
	for _, amostra := range alteracoes {

		chave := CoordAmostraTerreno{X: amostra.X, Y: amostra.Y, Z: amostra.Z}
		mundo.alteracoesTerreno[chave] = amostra.Densidade

		for coord, regiao := range mundo.regioes {
			ix, iy, iz, pertence := regiao.indicesDaAmostraGlobal(
				amostra.X, amostra.Y, amostra.Z,
			)
			if !pertence {
				continue
			}

			regiao.Campo[ix][iy][iz] = amostra.Densidade
			alteradas[coord] = struct{}{}
		}
	}

	for coord := range alteradas {
		mundo.reconstruirRegiaoTerreno(mundo.regioes[coord])
	}

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
