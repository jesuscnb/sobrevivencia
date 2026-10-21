package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EstadoInimigo uint8

const (
	EstadoPatrulha EstadoInimigo = iota
	EstadoPerseguicao
	EstadoAtaque
)

const (
	danoAtaqueInimigo      = 10
	intervaloAtaqueInimigo = float32(1.5)
	alcanceAtaqueInimigo   = float32(1.5)
)

type Inimigo struct {
	ID              IDObjeto
	Posicao         rl.Vector3
	Origem          rl.Vector3
	Estado          EstadoInimigo
	Vida            int
	DirecaoPatrulha uint8
	RecargaAtaque   float32
	Caminho         []rl.Vector3
	IndiceCaminho   int
	AlvoCaminho     CelulaNavegacao
	EsperaRecalculo float32
	TempoAnimacao   float64
	PosicaoAnterior rl.Vector3
	AnimacaoVisual  string
	RotacaoY        float32
	Visual          TipoVisualInimigo
}

func gerarInimigo(terreno rl.Model, ocupacao *IndiceOcupacao, semente int, posicaoJogador rl.Vector3, visual TipoVisualInimigo, existentes []Inimigo) (Inimigo, bool) {
	malhas := terreno.GetMeshes()
	if len(malhas) == 0 {
		return Inimigo{}, false
	}

	regiao := regiaoInicial()
	candidatos := gerarCandidatos(TipoInimigo, int64(semente), regiao, 100)

	for _, candidato := range candidatos {

		raio := rl.NewRay(
			rl.NewVector3(candidato.X, 5, candidato.Z),
			rl.NewVector3(0, -1, 0),
		)

		chao := rl.GetRayCollisionMesh(raio, malhas[0], terreno.Transform)

		if !chao.Hit || chao.Normal.Y < 0.85 {
			continue
		}

		posicao := chao.Point
		if !podePosicionar(ocupacao, TipoInimigo, posicao) {
			continue
		}

		if distanciaHorizontal2(posicao, posicaoJogador) < 3*3 {
			continue
		}

		pertoDeOutroInimigo := false
		for _, existente := range existentes {
			if distanciaHorizontal2(posicao, existente.Posicao) < 3*3 {
				pertoDeOutroInimigo = true
				break
			}
		}
		if pertoDeOutroInimigo {
			continue
		}

		return Inimigo{
			ID: IDObjeto{
				Tipo:    TipoInimigo,
				RegiaoX: regiao.X,
				RegiaoZ: regiao.Z,
				Ordem:   candidato.Ordem,
			},
			Posicao: posicao,
			Origem:  posicao,
			Estado:  EstadoPatrulha,
			Vida:    30,
			Visual:  visual,
		}, true
	}

	return Inimigo{}, false
}

func (mundo *Mundo) iniciarInimigo(posicaoJogador rl.Vector3) bool {
	mundo.inimigos = nil

	for _, visual := range []TipoVisualInimigo{VisualOrc, VisualBlob, VisualGhost} {
		inimigo, ok := gerarInimigo(
			mundo.modeloTerreno,
			mundo.ocupacao,
			mundo.semente,
			posicaoJogador,
			visual,
			mundo.inimigos,
		)
		if !ok {
			fmt.Printf("Nenhum ponto livre para o inimigo visual %d\n", visual)
			continue
		}

		mundo.inimigos = append(mundo.inimigos, inimigo)
	}

	return len(mundo.inimigos) > 0
}

func (mundo *Mundo) passoLivreInimigo(atual, tentativa rl.Vector3) (rl.Vector3, bool) {
	malhas := mundo.modeloTerreno.GetMeshes()
	if len(malhas) == 0 {
		return rl.Vector3{}, false
	}

	raio := rl.NewRay(
		rl.NewVector3(tentativa.X, 5, tentativa.Z),
		rl.NewVector3(0, -1, 0),
	)

	chao := rl.GetRayCollisionMesh(
		raio, malhas[0], mundo.modeloTerreno.Transform,
	)

	if !chao.Hit || chao.Normal.Y < 0.85 {
		return rl.Vector3{}, false
	}

	desnivel := chao.Point.Y - atual.Y
	if desnivel < -0.35 || desnivel > 0.35 {
		return rl.Vector3{}, false
	}

	proxima := chao.Point
	areaInimigo := areaAoRedor(proxima, 0.4)

	for _, objeto := range mundo.ocupacao.consultar(areaInimigo) {
		var bloqueio AreaXZ

		switch objeto.Tipo {
		case TipoCubo:
			bloqueio = areaDaCaixaXZ(caixaDoCubo(objeto.Posicao))
		case TipoArvore:
			bloqueio = areaDaCaixaXZ(caixaDoTronco(objeto.Posicao))
		case TipoRocha:
			bloqueio = areaDaCaixaXZ(
				caixaDaRocha(objeto.Posicao, mundo.limitesRocha),
			)
		case TipoParede:
			bloqueio = objeto.Area
		default:
			continue
		}

		if areasXZSobrepostas(areaInimigo, bloqueio) {
			return rl.Vector3{}, false
		}
	}

	return proxima, true
}

func (mundo *Mundo) segmentoLivreInimigo(origem, destino rl.Vector3) bool {
	dx := destino.X - origem.X
	dz := destino.Z - origem.Z
	distancia := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if distancia == 0 {
		return true
	}

	// Nenhuma amostra fica a mais de 0,2 unidade da anterior.
	passos := int(math.Ceil(float64(distancia / 0.2)))
	atual := origem

	for i := 1; i <= passos; i++ {
		parte := float32(i) / float32(passos)
		tentativa := rl.NewVector3(
			origem.X+dx*parte,
			atual.Y,
			origem.Z+dz*parte,
		)

		proxima, livre := mundo.passoLivreInimigo(atual, tentativa)
		if !livre {
			return false
		}
		atual = proxima
	}

	return true
}

func (mundo *Mundo) linhaAtaqueLivre(origem, destino rl.Vector3) bool {
	dx := destino.X - origem.X
	dz := destino.Z - origem.Z
	distancia := float32(math.Sqrt(float64(dx*dx + dz*dz)))

	if distancia == 0 {
		return true
	}

	// O raio tem direção de comprimento 1. Assim, colisao.Distance
	// pode ser comparada com a distância até o jogador.
	raio := rl.NewRay(
		rl.NewVector3(origem.X, 0, origem.Z),
		rl.NewVector3(dx/distancia, 0, dz/distancia),
	)

	areaTrecho := AreaXZ{
		MinX: min(origem.X, destino.X),
		MaxX: max(origem.X, destino.X),
		MinZ: min(origem.Z, destino.Z),
		MaxZ: max(origem.Z, destino.Z),
	}

	for _, objeto := range mundo.ocupacao.consultar(areaTrecho) {
		var obstaculo AreaXZ

		switch objeto.Tipo {
		case TipoCubo:
			obstaculo = areaDaCaixaXZ(caixaDoCubo(objeto.Posicao))
		case TipoArvore:
			obstaculo = areaDaCaixaXZ(caixaDoTronco(objeto.Posicao))
		case TipoRocha:
			obstaculo = areaDaCaixaXZ(
				caixaDaRocha(objeto.Posicao, mundo.limitesRocha),
			)
		case TipoParede:
			obstaculo = objeto.Area
		default:
			continue
		}

		// Usamos a caixa em Y=0 apenas para testar a linha no plano XZ.
		caixa := rl.NewBoundingBox(
			rl.NewVector3(obstaculo.MinX, -1, obstaculo.MinZ),
			rl.NewVector3(obstaculo.MaxX, 1, obstaculo.MaxZ),
		)

		colisao := rl.GetRayCollisionBox(raio, caixa)
		if colisao.Hit && colisao.Distance < distancia-0.01 {
			return false
		}
	}

	return true
}

func (mundo *Mundo) pontoPodeAtacarInimigo(posicao, posicaoJogador rl.Vector3) bool {
	if distanciaHorizontal2(posicao, posicaoJogador) >
		alcanceAtaqueInimigo*alcanceAtaqueInimigo {
		return false
	}

	pesJogadorY := posicaoJogador.Y - 1.7
	if math.Abs(float64(pesJogadorY-posicao.Y)) > 0.6 {
		return false
	}

	return mundo.linhaAtaqueLivre(posicao, posicaoJogador)
}

func (mundo *Mundo) inimigoPodeAtacar(inimigo *Inimigo, posicaoJogador rl.Vector3) bool {
	return mundo.pontoPodeAtacarInimigo(inimigo.Posicao, posicaoJogador)
}

func (mundo *Mundo) atualizarInimigos(delta float32, posicaoJogador rl.Vector3) int {
	if delta <= 0 {
		return 0
	}
	if delta > 0.05 {
		delta = 0.05
	}

	direcoes := [...]rl.Vector3{
		rl.NewVector3(1, 0, 0),
		rl.NewVector3(0, 0, 1),
		rl.NewVector3(-1, 0, 0),
		rl.NewVector3(0, 0, -1),
	}

	const velocidade float32 = 1.2
	const raioPatrulha float32 = 2
	danoTotal := 0

	for i := range mundo.inimigos {
		inimigo := &mundo.inimigos[i]

		if inimigo.Vida <= 0 {
			continue
		}

		inimigo.PosicaoAnterior = inimigo.Posicao
		inimigo.TempoAnimacao += float64(delta)

		if inimigo.RecargaAtaque > 0 {
			inimigo.RecargaAtaque -= delta
			if inimigo.RecargaAtaque < 0 {
				inimigo.RecargaAtaque = 0
			}
		}

		distancia2 := distanciaHorizontal2(inimigo.Posicao, posicaoJogador)

		switch inimigo.Estado {

		case EstadoPatrulha:
			if distancia2 <= 4*4 {
				inimigo.Estado = EstadoPerseguicao
			}

		case EstadoPerseguicao:
			if distancia2 >= 6*6 {
				inimigo.Estado = EstadoPatrulha
			} else if mundo.inimigoPodeAtacar(inimigo, posicaoJogador) {
				inimigo.Estado = EstadoAtaque
			}

		case EstadoAtaque:
			if distancia2 >= 6*6 {
				inimigo.Estado = EstadoPatrulha
			} else if !mundo.inimigoPodeAtacar(inimigo, posicaoJogador) {
				inimigo.Estado = EstadoPerseguicao
			}
		}

		if inimigo.Estado == EstadoAtaque {
			if inimigo.RecargaAtaque == 0 {
				danoTotal += danoAtaqueInimigo
				inimigo.RecargaAtaque = intervaloAtaqueInimigo
			}
			continue
		}

		if inimigo.Estado == EstadoPerseguicao {
			mundo.moverInimigoPeloCaminho(
				inimigo,
				posicaoJogador,
				1.8,
				delta,
			)
			continue
		}

		if inimigo.Estado != EstadoPatrulha {
			continue
		}

		if distanciaHorizontal2(inimigo.Posicao, inimigo.Origem) >
			raioPatrulha*raioPatrulha {
			mundo.moverInimigoEmDirecao(
				inimigo,
				inimigo.Origem,
				1.2,
				1.5,
				delta,
			)
			continue
		}

		for tentativa := 0; tentativa < len(direcoes); tentativa++ {
			direcao := direcoes[inimigo.DirecaoPatrulha]
			proxima := rl.NewVector3(
				inimigo.Posicao.X+direcao.X*velocidade*delta,
				inimigo.Posicao.Y,
				inimigo.Posicao.Z+direcao.Z*velocidade*delta,
			)

			if distanciaHorizontal2(proxima, inimigo.Origem) <= raioPatrulha*raioPatrulha {
				if posicao, livre := mundo.passoLivreInimigo(inimigo.Posicao, proxima); livre {
					inimigo.Posicao = posicao
					break
				}
			}

			inimigo.DirecaoPatrulha = (inimigo.DirecaoPatrulha + 1) % uint8(len(direcoes))
		}
	}
	return danoTotal
}

func (mundo *Mundo) moverInimigoEmDirecao(inimigo *Inimigo, alvo rl.Vector3, velocidade, distanciaParada, delta float32) {
	distancia2 := distanciaHorizontal2(inimigo.Posicao, alvo)
	if distancia2 <= distanciaParada*distanciaParada {
		return
	}

	distancia := float32(math.Sqrt(float64(distancia2)))
	passo := velocidade * delta
	if passo > distancia-distanciaParada {
		passo = distancia - distanciaParada
	}

	dx := alvo.X - inimigo.Posicao.X
	dz := alvo.Z - inimigo.Posicao.Z
	tentativa := rl.NewVector3(
		inimigo.Posicao.X+dx/distancia*passo,
		inimigo.Posicao.Y,
		inimigo.Posicao.Z+dz/distancia*passo,
	)

	if posicao, livre := mundo.passoLivreInimigo(
		inimigo.Posicao, tentativa,
	); livre {
		inimigo.Posicao = posicao
	}
}

func (inimigo *Inimigo) limparCaminho() {
	inimigo.Caminho = nil
	inimigo.IndiceCaminho = 0
	inimigo.EsperaRecalculo = 0
}

func (mundo *Mundo) moverInimigoPeloCaminho(inimigo *Inimigo, alvo rl.Vector3, velocidade, delta float32) {
	grade := novaGradeNavegacao()
	celulaAlvo, dentroDaGrade := grade.celulaDaPosicao(alvo)
	if !dentroDaGrade {
		inimigo.limparCaminho()
		return
	}

	if inimigo.EsperaRecalculo > 0 {
		inimigo.EsperaRecalculo -= delta
		if inimigo.EsperaRecalculo < 0 {
			inimigo.EsperaRecalculo = 0
		}
	}

	rotaTerminou := inimigo.IndiceCaminho >= len(inimigo.Caminho)
	alvoMudou := inimigo.AlvoCaminho != celulaAlvo

	if (rotaTerminou || alvoMudou) && inimigo.EsperaRecalculo == 0 {
		caminho, encontrado := grade.encontrarCaminho(
			mundo,
			inimigo.Posicao,
			alvo,
		)

		// Mesmo quando não há caminho, aguardamos antes de tentar de novo.
		inimigo.EsperaRecalculo = 0.5
		inimigo.AlvoCaminho = celulaAlvo
		inimigo.Caminho = nil
		inimigo.IndiceCaminho = 0

		if encontrado {
			inimigo.Caminho = caminho
		}
	}

	const raioChegada float32 = 0.12
	for inimigo.IndiceCaminho < len(inimigo.Caminho) {
		ponto := inimigo.Caminho[inimigo.IndiceCaminho]
		if distanciaHorizontal2(inimigo.Posicao, ponto) > raioChegada*raioChegada {
			break
		}
		inimigo.IndiceCaminho++
	}

	if inimigo.IndiceCaminho >= len(inimigo.Caminho) {
		return
	}

	ponto := inimigo.Caminho[inimigo.IndiceCaminho]
	dx := ponto.X - inimigo.Posicao.X
	dz := ponto.Z - inimigo.Posicao.Z
	distancia := float32(math.Sqrt(float64(dx*dx + dz*dz)))

	passo := velocidade * delta
	if passo > distancia {
		passo = distancia
	}

	tentativa := rl.NewVector3(
		inimigo.Posicao.X+dx/distancia*passo,
		inimigo.Posicao.Y,
		inimigo.Posicao.Z+dz/distancia*passo,
	)

	proxima, livre := mundo.passoLivreInimigo(inimigo.Posicao, tentativa)
	if rl.IsKeyDown(rl.KeyF2) {
		fmt.Printf(
			"Passo: ponto=%d/%d pos=(%.2f, %.2f) tentativa=(%.2f, %.2f) livre=%t\n",
			inimigo.IndiceCaminho+1, len(inimigo.Caminho),
			inimigo.Posicao.X, inimigo.Posicao.Z,
			tentativa.X, tentativa.Z,
			livre,
		)
	}
	if !livre {
		// Algo mudou no mundo ou esta rota deixou de ser válida.
		// No próximo quadro, procuraremos outra.
		inimigo.limparCaminho()
		return
	}

	inimigo.Posicao = proxima
}

func (mundo *Mundo) suavizarCaminho(origem rl.Vector3, caminho []rl.Vector3) []rl.Vector3 {
	if len(caminho) < 2 {
		return caminho
	}

	suavizado := make([]rl.Vector3, 0, len(caminho))
	pontoAtual := origem

	for primeiro := 0; primeiro < len(caminho); {
		escolhido := primeiro

		// Tenta alcançar o ponto mais distante sem atravessar obstáculos.
		for i := len(caminho) - 1; i > primeiro; i-- {
			if mundo.segmentoLivreInimigo(pontoAtual, caminho[i]) {
				escolhido = i
				break
			}
		}

		suavizado = append(suavizado, caminho[escolhido])
		pontoAtual = caminho[escolhido]
		primeiro = escolhido + 1
	}

	return suavizado
}

func anguloDaDirecao(dx, dz float32) float32 {
	return float32(math.Atan2(float64(dx), float64(dz)) * 180 / math.Pi)
}
