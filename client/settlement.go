package main

import rl "github.com/gen2brain/raylib-go/raylib"

const versaoAssentamentoAtual = 1

// Pecas procedurais têm identidade própria e não entram no save de construções
// feitas pelo jogador. A decisão por lote é persistida separadamente.
type PecaGerada struct {
	Lote   IDLote
	Regiao CoordRegiaoTerreno
	Peca   PecaConstrucao
}

func (mundo *Mundo) adicionarPecaGerada(lote IDLote, regiao CoordRegiaoTerreno,
	parte uint32, tipo TipoObjeto, posicao rl.Vector3, rotacao float32,
	catalogo *Catalogo) {
	modelo := catalogo.modeloObrigatorio(tipo)
	escala := catalogo.escala(tipo)
	id := IDObjeto{Tipo: tipo, RegiaoX: regiao.X, RegiaoZ: regiao.Z,
		Ordem: lote.Ordem*5 + parte, Gerado: true}
	area := areaPecaXZ(modelo, posicao, escala, rotacao)
	if !mundo.ocupacao.adicionar(RegistroOcupacao{
		ID: id, Tipo: tipo, Posicao: posicao, Area: area,
	}) {
		panic("ID duplicado em peça gerada")
	}
	mundo.pecasGeradas = append(mundo.pecasGeradas, PecaGerada{
		Lote: lote, Regiao: regiao,
		Peca: PecaConstrucao{ID: id, Tipo: tipo, Posicao: posicao,
			Escala: escala, Rotacao: rotacao},
	})
	if tipo == TipoPiso {
		mundo.removerGramaNaArea(area)
	}
}

func (mundo *Mundo) montarCabana(lote CandidatoLote, posicao rl.Vector3,
	catalogo *Catalogo) {
	regiao := coordenadaRegiao(lote.Centro.X, 0, lote.Centro.Z)
	modeloPiso := catalogo.modeloObrigatorio(TipoPiso)
	escalaPiso := catalogo.escala(TipoPiso)
	area := areaPecaXZ(modeloPiso, posicao, escalaPiso, 0)
	topoPiso := posicao.Y + rl.GetModelBoundingBox(modeloPiso).Max.Y*escalaPiso
	mundo.adicionarPecaGerada(lote.ID, regiao, 0, TipoPiso, posicao, 0, catalogo)

	escalaParede := catalogo.escala(TipoParede)
	limitesParede := rl.GetModelBoundingBox(catalogo.modeloObrigatorio(TipoParede))
	baseParede := topoPiso - limitesParede.Min.Y*escalaParede
	centroX := (area.MinX + area.MaxX) / 2
	centroZ := (area.MinZ + area.MaxZ) / 2
	bordas := [4]rl.Vector3{
		rl.NewVector3(centroX, baseParede, area.MinZ),
		rl.NewVector3(area.MaxX, baseParede, centroZ),
		rl.NewVector3(centroX, baseParede, area.MaxZ),
		rl.NewVector3(area.MinX, baseParede, centroZ),
	}
	rotacoes := [4]float32{0, 90, 0, 90}
	dx := lote.Acesso.X - lote.Centro.X
	dz := lote.Acesso.Z - lote.Centro.Z
	aberta := 0
	if dx > 0 && dx >= abs32(dz) {
		aberta = 1
	} else if dz > 0 && dz >= abs32(dx) {
		aberta = 2
	} else if dx < 0 && -dx >= abs32(dz) {
		aberta = 3
	}
	parte := uint32(1)
	for i := range bordas {
		if i == aberta {
			continue
		}
		mundo.adicionarPecaGerada(lote.ID, regiao, parte, TipoParede,
			bordas[i], rotacoes[i], catalogo)
		parte++
	}

	alturaParede := (limitesParede.Max.Y - limitesParede.Min.Y) * escalaParede
	escalaTelhado := catalogo.escala(TipoTelhado)
	limitesTelhado := rl.GetModelBoundingBox(catalogo.modeloObrigatorio(TipoTelhado))
	posicaoTelhado := rl.NewVector3(centroX,
		topoPiso+alturaParede-limitesTelhado.Min.Y*escalaTelhado, centroZ)
	mundo.adicionarPecaGerada(lote.ID, regiao, 4, TipoTelhado,
		posicaoTelhado, 0, catalogo)
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// Cada lote recebe uma decisão uma única vez, após todas as regiões da área e
// do acesso estarem carregadas. Uma decisão inválida não muda se o jogador
// remover um recurso mais tarde.
func (mundo *Mundo) atualizarAssentamento(catalogo *Catalogo) {
	if mundo.versaoAssentamento == 0 {
		return
	}
	for _, id := range mundo.lotesAssentamentoInicial {
		if int(id.Ordem) >= len(mundo.lotesIniciais) {
			continue
		}
		lote := mundo.lotesIniciais[id.Ordem]
		if !mundo.superficieCarregada(lote.Centro.X, lote.Centro.Z) {
			continue
		}
		carregada := false
		for _, peca := range mundo.pecasGeradas {
			if peca.Lote == id {
				carregada = true
				break
			}
		}
		if carregada {
			continue
		}
		aceito, decidido := mundo.lotesDecididos[id]
		if !decidido {
			avaliacao := avaliarLote(mundo, catalogo, lote)
			if avaliacao.Situacao == LotePendente {
				continue
			}
			aceito = avaliacao.Situacao == LoteLivre
			mundo.lotesDecididos[id] = aceito
		}
		if !aceito {
			continue
		}
		modelo := catalogo.modeloObrigatorio(TipoPiso)
		escala := catalogo.escala(TipoPiso)
		raio := raioHorizontalLimites(rl.GetModelBoundingBox(modelo), escala) + 1
		if !mundo.areaTerrenoCarregada(lote.Centro.X, lote.Centro.Z, raio) {
			continue
		}
		posicao, ok := posicaoNoTerreno(mundo, modelo,
			lote.Centro.X, lote.Centro.Z, escala, -1)
		if !ok {
			continue
		}
		mundo.montarCabana(lote, posicao, catalogo)
	}
}

func (mundo *Mundo) descarregarAssentamentoRegiao(regiao CoordRegiaoTerreno) {
	restantes := mundo.pecasGeradas[:0]
	for _, gerada := range mundo.pecasGeradas {
		if gerada.Regiao == regiao {
			mundo.ocupacao.remover(gerada.Peca.ID)
			continue
		}
		restantes = append(restantes, gerada)
	}
	mundo.pecasGeradas = restantes
}
