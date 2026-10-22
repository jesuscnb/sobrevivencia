package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	espacamentoLotes float32 = 18
	afastamentoLotes float32 = 4
)

type IDLote struct {
	Plano CoordPlanoTerritorial
	Ordem uint32
}

type CandidatoLote struct {
	ID     IDLote
	Centro PontoTerritorial // Local proposto para a construção.
	Acesso PontoTerritorial // Ponto correspondente na estrada.
}

type SituacaoLote uint8

const (
	LotePendente SituacaoLote = iota
	LoteInvalido
	LoteLivre
	LoteOcupado
)

type AvaliacaoLote struct {
	Situacao SituacaoLote
	Motivo   string
	Posicao  rl.Vector3
	Area     AreaXZ
}

func candidatosLotesEstrada(
	plano CoordPlanoTerritorial,
	caminho CaminhoEstrada,
) []CandidatoLote {
	var comprimentoTotal float32

	for i := 0; i+1 < len(caminho.Pontos); i++ {
		a, b := caminho.Pontos[i], caminho.Pontos[i+1]
		comprimentoTotal += float32(math.Hypot(
			float64(b.X-a.X), float64(b.Z-a.Z),
		))
	}

	proximo := espacamentoLotes / 2
	var percorrido float32
	var lotes []CandidatoLote

	for i := 0; i+1 < len(caminho.Pontos); i++ {
		a, b := caminho.Pontos[i], caminho.Pontos[i+1]
		dx, dz := b.X-a.X, b.Z-a.Z
		comprimento := float32(math.Hypot(float64(dx), float64(dz)))
		if comprimento < 0.0001 {
			continue
		}

		for proximo <= percorrido+comprimento &&
			proximo < comprimentoTotal-espacamentoLotes/2 {
			t := (proximo - percorrido) / comprimento
			acesso := PontoTerritorial{
				X: a.X + dx*t,
				Z: a.Z + dz*t,
			}

			lado := float32(1)
			if len(lotes)%2 == 1 {
				lado = -1
			}

			// (-dz, dx) aponta para o lado do trecho da estrada.
			centro := PontoTerritorial{
				X: acesso.X + lado*(-dz/comprimento)*afastamentoLotes,
				Z: acesso.Z + lado*(dx/comprimento)*afastamentoLotes,
			}

			lotes = append(lotes, CandidatoLote{
				ID: IDLote{
					Plano: plano,
					Ordem: uint32(len(lotes)),
				},
				Centro: centro,
				Acesso: acesso,
			})
			proximo += espacamentoLotes
		}

		percorrido += comprimento
	}

	return lotes
}

// Escolhe uma sequência estável ao redor do candidato mais próximo do centro
// planejado. A escolha não consulta terreno, objetos ou regiões carregadas.
func selecionarLotesAssentamento(plano PlanoTerritorial, lotes []CandidatoLote) []IDLote {
	if len(lotes) == 0 {
		return nil
	}

	maisProximo := 0
	menorDistancia := float32(math.Inf(1))
	for i, lote := range lotes {
		dx := lote.Centro.X - plano.CentroX
		dz := lote.Centro.Z - plano.CentroZ
		distancia := dx*dx + dz*dz
		if distancia < menorDistancia {
			menorDistancia = distancia
			maisProximo = i
		}
	}

	quantidade := min(5, len(lotes))
	escolhidos := make([]IDLote, 0, quantidade)
	for i := 0; i < quantidade; i++ {
		indice := (maisProximo - quantidade/2 + i + len(lotes)) % len(lotes)
		escolhidos = append(escolhidos, lotes[indice].ID)
	}
	return escolhidos
}

// O lote é ocupado pela construção efetiva, registrada no índice espacial.
// Assim, saves antigos continuam válidos e não aparecem clareiras vazias.
func loteComConstrucao(mundo *Mundo, area AreaXZ) bool {
	for _, registro := range mundo.ocupacao.consultar(area) {
		if registro.Tipo == TipoPiso && areasXZSobrepostas(area, registro.Area) {
			return true
		}
	}
	return false
}

// Verifica uma passagem de 0,8 unidade entre o lote e seu ponto de acesso.
// Cada amostra cobre o intervalo até a seguinte (passos de no máximo 0,4).
func estadoAcessoLote(mundo *Mundo, lote CandidatoLote) (SituacaoLote, string) {
	dx := lote.Centro.X - lote.Acesso.X
	dz := lote.Centro.Z - lote.Acesso.Z
	comprimento := float32(math.Hypot(float64(dx), float64(dz)))
	passos := max(1, int(math.Ceil(float64(comprimento/0.4))))

	for i := 0; i <= passos; i++ {
		t := float32(i) / float32(passos)
		x := lote.Acesso.X + dx*t
		z := lote.Acesso.Z + dz*t
		// Objetos da região vizinha também podem alcançar a passagem.
		if !mundo.areaTerrenoCarregada(x, z, 0.85) {
			return LotePendente, "pendente: acesso ainda não carregado"
		}
	}

	var alturaAnterior float32
	for i := 0; i <= passos; i++ {
		t := float32(i) / float32(passos)
		x := lote.Acesso.X + dx*t
		z := lote.Acesso.Z + dz*t
		chao := mundo.chaoCarregado(x, z, 0)
		if !chao.Hit || chao.Normal.Y < 0.7 ||
			(i > 0 && math.Abs(float64(chao.Point.Y-alturaAnterior)) > 0.35) {
			return LoteInvalido, "inválido: acesso sem apoio caminhável"
		}
		alturaAnterior = chao.Point.Y
		area := areaAoRedor(rl.NewVector3(x, 0, z), 0.4)
		if areaTemObstaculo(mundo, area) {
			return LoteInvalido, "inválido: acesso à estrada bloqueado"
		}
	}
	return LoteLivre, ""
}

// Avaliação local. O estado pendente nunca deve ser tratado como rejeição.
// A escolha de um assentamento deve usar IDs estáveis, sem depender da ordem
// em que as regiões são carregadas.
func avaliarLote(mundo *Mundo, catalogo *Catalogo, lote CandidatoLote) AvaliacaoLote {
	modelo := catalogo.modeloObrigatorio(TipoPiso)
	escala := catalogo.escala(TipoPiso)

	centro := rl.NewVector3(lote.Centro.X, 0, lote.Centro.Z)
	area := areaPecaXZ(modelo, centro, escala, 0)
	resultado := AvaliacaoLote{Area: area}
	// Espera também os geradores das regiões vizinhas que podem colocar
	// árvores ou pedras com volume sobre a borda da área do piso.
	raioBusca := raioHorizontalLimites(rl.GetModelBoundingBox(modelo), escala) + 1
	if !mundo.areaTerrenoCarregada(lote.Centro.X, lote.Centro.Z, raioBusca) {
		resultado.Motivo = "pendente: vizinhança do lote ainda não carregada"
		return resultado
	}

	// O piso pode atravessar a divisa entre duas regiões.
	cantos := [4][2]float32{
		{area.MinX, area.MinZ},
		{area.MinX, area.MaxZ},
		{area.MaxX, area.MinZ},
		{area.MaxX, area.MaxZ},
	}
	for _, canto := range cantos {
		if !mundo.superficieCarregada(canto[0], canto[1]) {
			resultado.Motivo = "pendente: região do piso ainda não carregada"
			return resultado
		}
	}
	if loteComConstrucao(mundo, area) {
		resultado.Situacao = LoteOcupado
		resultado.Motivo = "ocupado: já existe um piso no lote"
		return resultado
	}
	resultado.Situacao = LoteInvalido

	// O raio inclui toda a peça, inclusive seus cantos. Isso evita aceitar
	// um piso que toque a faixa da estrada entre dois pontos amostrados.
	trechos := mundo.estradaInicial.trechosNaArea(area, alcanceEstrada)
	raioPiso := raioHorizontalLimites(rl.GetModelBoundingBox(modelo), escala)
	if sobrepoeFaixaEstrada(trechos, lote.Centro.X, lote.Centro.Z, raioPiso) {
		resultado.Motivo = "inválido: piso invade a estrada"
		return resultado
	}

	posicao, ok := posicaoNoTerreno(
		mundo, modelo, lote.Centro.X, lote.Centro.Z, escala, -1,
	)
	if !ok {
		resultado.Motivo = "inválido: sem superfície no centro"
		return resultado
	}
	resultado.Posicao = posicao

	base := posicao.Y + rl.GetModelBoundingBox(modelo).Min.Y*escala
	if !pisoTemApoio(mundo, area, base) {
		resultado.Motivo = "inválido: inclinação ou apoio insuficiente"
		return resultado
	}

	if areaTemObstaculo(mundo, area) {
		resultado.Motivo = "inválido: árvore, pedra ou construção ocupa o piso"
		return resultado
	}
	if situacao, motivo := estadoAcessoLote(mundo, lote); motivo != "" {
		resultado.Situacao = situacao
		resultado.Motivo = motivo
		return resultado
	}

	resultado.Situacao = LoteLivre
	resultado.Motivo = "válido: piso apoiado, área e acesso livres"
	return resultado
}

func estadoLote(jogo *Jogo, lote CandidatoLote) string {
	return avaliarLote(&jogo.mundo, &jogo.catalogo, lote).Motivo
}
