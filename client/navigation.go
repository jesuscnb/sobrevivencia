package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type CelulaNavegacao struct {
	X, Z int
}

type GradeNavegacao struct {
	MinX, MinZ float32
	Passo      float32
	Largura    int
	Altura     int
}

func novaGradeNavegacao() GradeNavegacao {
	return GradeNavegacao{
		MinX:    -8,
		MinZ:    -8,
		Passo:   0.5,
		Largura: 32,
		Altura:  32,
	}
}

func (grade GradeNavegacao) celulaDaPosicao(posicao rl.Vector3) (CelulaNavegacao, bool) {
	// O limite maximo fica fora da grade: de -8 ate 8 existem 32 celulas
	// de 0,5 unidade, numeradas de 0 a 31 em cada eixo.
	maxX := grade.MinX + float32(grade.Largura)*grade.Passo
	maxZ := grade.MinZ + float32(grade.Altura)*grade.Passo

	if posicao.X < grade.MinX || posicao.X >= maxX ||
		posicao.Z < grade.MinZ || posicao.Z >= maxZ {
		return CelulaNavegacao{}, false
	}

	// Subtrair o minimo transforma o inicio da grade em zero.
	// Dividir pelo passo converte a distancia em indice de celula.
	return CelulaNavegacao{
		X: int((posicao.X - grade.MinX) / grade.Passo),
		Z: int((posicao.Z - grade.MinZ) / grade.Passo),
	}, true
}

func (grade GradeNavegacao) centroDaCelula(celula CelulaNavegacao) (float32, float32) {
	// O indice indica o inicio da celula. Somar 0,5 leva ao centro dela.
	x := grade.MinX + (float32(celula.X)+0.5)*grade.Passo
	z := grade.MinZ + (float32(celula.Z)+0.5)*grade.Passo
	return x, z
}

func (grade GradeNavegacao) passoParaCelula(mundo *Mundo, atual rl.Vector3, destino CelulaNavegacao) (rl.Vector3, bool) {

	if destino.X < 0 || destino.X >= grade.Largura ||
		destino.Z < 0 || destino.Z >= grade.Altura {
		return rl.Vector3{}, false
	}

	x, z := grade.centroDaCelula(destino)
	tentativa := rl.NewVector3(x, atual.Y, z)

	// O centro pode estar livre mesmo que o caminho até ele cruze uma pedra.
	if !mundo.segmentoLivreInimigo(atual, tentativa) {
		return rl.Vector3{}, false
	}

	return mundo.passoLivreInimigo(atual, tentativa)
}

func (grade GradeNavegacao) vizinhosDaCelula(celula CelulaNavegacao) []CelulaNavegacao {

	direcoes := [...]CelulaNavegacao{
		{X: 1, Z: 0},  // direita
		{X: -1, Z: 0}, // esquerda
		{X: 0, Z: 1},  // frente
		{X: 0, Z: -1}, // trás
	}

	vizinhos := make([]CelulaNavegacao, 0, 4)

	for _, direcao := range direcoes {

		candidato := CelulaNavegacao{
			X: celula.X + direcao.X,
			Z: celula.Z + direcao.Z,
		}

		if candidato.X < 0 || candidato.X >= grade.Largura ||
			candidato.Z < 0 || candidato.Z >= grade.Altura {
			continue
		}

		vizinhos = append(vizinhos, candidato)
	}

	return vizinhos
}

type NoBusca struct {
	Celula     CelulaNavegacao
	Posicao    rl.Vector3
	Custo      int // g: passos percorridos desde o início
	Estimativa int // h: passos que faltariam sem obstáculos
}

func (no NoBusca) prioridade() int {
	return no.Custo + no.Estimativa // f = g + h
}

func distanciaManhattan(a, b CelulaNavegacao) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}

	dz := a.Z - b.Z
	if dz < 0 {
		dz = -dz
	}

	return dx + dz
}

func reconstruirCaminho(
	origem, destino CelulaNavegacao,
	anterior map[CelulaNavegacao]CelulaNavegacao,
	posicoes map[CelulaNavegacao]rl.Vector3,
) []rl.Vector3 {
	caminho := []rl.Vector3{}

	// Voltamos do destino até a origem seguindo os registros de "anterior".
	for celula := destino; celula != origem; celula = anterior[celula] {
		caminho = append(caminho, posicoes[celula])
	}

	// O caminho foi encontrado de trás para frente; invertemos a ordem.
	for i, j := 0, len(caminho)-1; i < j; i, j = i+1, j-1 {
		caminho[i], caminho[j] = caminho[j], caminho[i]
	}

	return caminho
}

func estimativaPassosAteAtaque(
	posicao, jogador rl.Vector3,
	passo float32,
) int {
	distancia := float32(math.Sqrt(
		float64(distanciaHorizontal2(posicao, jogador)),
	))
	restante := distancia - alcanceAtaqueInimigo
	if restante <= 0 {
		return 0
	}

	return int(math.Ceil(float64(restante / passo)))
}

func (grade GradeNavegacao) encontrarCaminho(mundo *Mundo, origem, destino rl.Vector3) ([]rl.Vector3, bool) {
	inicio, inicioValido := grade.celulaDaPosicao(origem)
	_, fimValido := grade.celulaDaPosicao(destino)
	if !inicioValido || !fimValido {
		return nil, false
	}

	if mundo.pontoPodeAtacarInimigo(origem, destino) {
		return []rl.Vector3{}, true
	}

	abertos := []NoBusca{{
		Celula:     inicio,
		Posicao:    origem,
		Custo:      0,
		Estimativa: estimativaPassosAteAtaque(origem, destino, grade.Passo),
	}}

	melhorCusto := map[CelulaNavegacao]int{inicio: 0}
	anterior := make(map[CelulaNavegacao]CelulaNavegacao)
	posicoes := map[CelulaNavegacao]rl.Vector3{inicio: origem}
	fechados := make(map[CelulaNavegacao]bool)

	for len(abertos) > 0 {
		// Escolhe a célula com menor f = custo percorrido + estimativa.
		indice := 0
		for i := 1; i < len(abertos); i++ {
			if abertos[i].prioridade() < abertos[indice].prioridade() {
				indice = i
			}
		}

		atual := abertos[indice]
		abertos[indice] = abertos[len(abertos)-1]
		abertos = abertos[:len(abertos)-1]

		// Uma célula pode aparecer mais de uma vez em "abertos".
		// Ignoramos a versão antiga quando já encontramos uma rota melhor.
		if fechados[atual.Celula] || atual.Custo > melhorCusto[atual.Celula] {
			continue
		}

		if mundo.pontoPodeAtacarInimigo(atual.Posicao, destino) {
			caminho := reconstruirCaminho(
				inicio, atual.Celula, anterior, posicoes,
			)
			return mundo.suavizarCaminho(origem, caminho), true
		}

		fechados[atual.Celula] = true

		for _, vizinho := range grade.vizinhosDaCelula(atual.Celula) {
			if fechados[vizinho] {
				continue
			}

			ponto, livre := grade.passoParaCelula(mundo, atual.Posicao, vizinho)
			if !livre {
				continue
			}

			novoCusto := atual.Custo + 1
			custoAnterior, conhecido := melhorCusto[vizinho]
			if conhecido && novoCusto >= custoAnterior {
				continue
			}

			melhorCusto[vizinho] = novoCusto
			anterior[vizinho] = atual.Celula
			posicoes[vizinho] = ponto

			abertos = append(abertos, NoBusca{
				Celula:     vizinho,
				Posicao:    ponto,
				Custo:      novoCusto,
				Estimativa: estimativaPassosAteAtaque(ponto, destino, grade.Passo),
			})
		}
	}

	return nil, false // Nenhum caminho alcançou o destino.
}
