package main

import (
	"fmt"
	"time"

	fastnoise "github.com/Auburn/FastNoiseLite/Go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mundo guarda o terreno e as instâncias presentes na partida.
type Mundo struct {
	semente                  int
	versaoGerador            int
	ruido                    *fastnoise.State[float32]
	regioes                  map[CoordRegiaoTerreno]*RegiaoTerreno
	cubos                    []CuboColetavel
	arvores                  []Arvore
	rochas                   []Rocha
	gramas                   []Grama
	construcoes              []PecaConstrucao
	ocupacao                 *IndiceOcupacao
	limitesRocha             rl.BoundingBox
	inimigos                 []Inimigo
	alteracoesTerreno        map[CoordAmostraTerreno]float32
	recursosRemovidos        map[IDObjeto]struct{}
	ultimoCarregamento       time.Duration
	ultimaDescarga           time.Duration
	estradaInicial           CaminhoEstrada
	lotesIniciais            []CandidatoLote
	lotesAssentamentoInicial []IDLote
	versaoAssentamento       int
	lotesDecididos           map[IDLote]bool
	pecasGeradas             []PecaGerada
}

type PecaConstrucao struct {
	ID      IDObjeto
	Tipo    TipoObjeto
	Posicao rl.Vector3
	Escala  float32
	Rotacao float32
}

func novoMundo(ruido *fastnoise.State[float32], semente, versao int, catalogo *Catalogo) Mundo {
	mundo := Mundo{
		semente:           semente,
		ruido:             ruido,
		versaoGerador:     versao,
		ocupacao:          novoIndiceOcupacao(2),
		regioes:           make(map[CoordRegiaoTerreno]*RegiaoTerreno),
		alteracoesTerreno: make(map[CoordAmostraTerreno]float32),
		recursosRemovidos: make(map[IDObjeto]struct{}),
		lotesDecididos:    make(map[IDLote]bool),
		estradaInicial:    gerarEstradaInicial(int64(semente)),
	}

	mundo.lotesIniciais = candidatosLotesEstrada(
		CoordPlanoTerritorial{},
		mundo.estradaInicial,
	)
	planoInicial := gerarPlanoTerritorial(int64(semente), CoordPlanoTerritorial{})
	mundo.lotesAssentamentoInicial = selecionarLotesAssentamento(
		planoInicial, mundo.lotesIniciais,
	)

	mundo.carregarRegiaoTerreno(CoordRegiaoTerreno{})
	mundo.carregarRegiaoTerreno(CoordRegiaoTerreno{X: 1})

	mundo.cubos = gerarCubos(ruido, semente)

	registrarOcupacao(mundo.ocupacao, IDObjeto{Tipo: TipoInicioJogador}, TipoInicioJogador, rl.NewVector3(4, 0, 4))
	for _, cubo := range mundo.cubos {
		registrarOcupacao(mundo.ocupacao, cubo.ID, TipoCubo, cubo.Posicao)
	}

	modeloArvore := catalogo.modeloObrigatorio(TipoArvore)
	modeloRocha := catalogo.modeloObrigatorio(TipoRocha)
	modeloGrama := catalogo.modeloObrigatorio(TipoGrama)

	mundo.carregarArvoresRegiao(CoordRegiaoTerreno{}, modeloArvore)
	mundo.carregarArvoresRegiao(CoordRegiaoTerreno{X: 1}, modeloArvore)

	mundo.limitesRocha = rl.GetModelBoundingBox(modeloRocha)

	mundo.carregarRochasRegiao(CoordRegiaoTerreno{}, modeloRocha)
	mundo.carregarRochasRegiao(CoordRegiaoTerreno{X: 1}, modeloRocha)

	mundo.carregarGramaRegiao(CoordRegiaoTerreno{}, modeloGrama)
	mundo.carregarGramaRegiao(CoordRegiaoTerreno{X: 1}, modeloGrama)

	return mundo
}

func (mundo *Mundo) adicionarPeca(tipo TipoObjeto, posicao rl.Vector3, escala float32, rotacao float32, modelo rl.Model) {
	id := IDObjeto{Tipo: tipo, Ordem: uint32(len(mundo.construcoes))}
	area := areaPecaXZ(modelo, posicao, escala, rotacao)

	registro := RegistroOcupacao{
		ID:      id,
		Tipo:    tipo,
		Posicao: posicao,
		Area:    area,
	}

	if !mundo.ocupacao.adicionar(registro) {
		panic("não foi possível registrar peça de construção")
	}
	mundo.construcoes = append(mundo.construcoes, PecaConstrucao{
		ID: id, Tipo: tipo, Posicao: posicao, Escala: escala, Rotacao: rotacao,
	})

	if tipo == TipoPiso {
		mundo.removerGramaNaArea(area)
	}

}

func (mundo *Mundo) colisaoTerreno(raio rl.Ray) rl.RayCollision {
	var maisProxima rl.RayCollision

	for _, regiao := range mundo.regioes {
		posicao := regiao.posicaoMundo()

		// A caixa serve apenas para descartar regiões distantes. Uma pequena
		// margem inclui raios paralelos que começam exatamente na divisa,
		// caso que GetRayCollisionBox rejeita na borda da caixa.
		const margemFiltro = float32(0.001)
		limites := rl.NewBoundingBox(
			rl.NewVector3(posicao.X-8-margemFiltro, posicao.Y-4-margemFiltro, posicao.Z-8-margemFiltro),
			rl.NewVector3(posicao.X+8+margemFiltro, posicao.Y+4+margemFiltro, posicao.Z+8+margemFiltro),
		)
		if !rl.GetRayCollisionBox(raio, limites).Hit {
			continue
		}

		transformacao := rl.MatrixMultiply(
			regiao.Modelo.Transform,
			rl.MatrixTranslate(posicao.X, posicao.Y, posicao.Z),
		)

		for _, malha := range regiao.Modelo.GetMeshes() {
			colisao := rl.GetRayCollisionMesh(raio, malha, transformacao)
			if colisao.Hit && (!maisProxima.Hit || colisao.Distance < maisProxima.Distance) {
				maisProxima = colisao
			}
		}
	}

	return maisProxima
}

func (mundo *Mundo) areaTerrenoCarregada(x, z, margem float32) bool {
	deslocamentos := [2]float32{-margem, margem}

	for _, dx := range deslocamentos {
		for _, dz := range deslocamentos {
			coord := coordenadaRegiao(x+dx, 0, z+dz)
			if _, existe := mundo.regioes[coord]; !existe {
				return false
			}
		}
	}

	return true
}

func (mundo *Mundo) chaoCarregado(x, z, margem float32) rl.RayCollision {
	if !mundo.areaTerrenoCarregada(x, z, margem) {
		return rl.RayCollision{}
	}

	raioParaBaixo := rl.NewRay(
		rl.NewVector3(x, 5, z),
		rl.NewVector3(0, -1, 0),
	)
	return mundo.colisaoTerreno(raioParaBaixo)
}

func (mundo *Mundo) reconstruirRegiaoTerreno(regiao *RegiaoTerreno) {
	novoModelo := criarModeloTerreno(&regiao.Campo, regiao.Coord, mundo.densidadeGlobal)
	rl.UnloadModel(regiao.Modelo)
	regiao.Modelo = novoModelo
}

func (mundo *Mundo) reconstruirTerreno() {
	// Ainda é usado ao carregar um save antigo da região inicial.
	mundo.reconstruirRegiaoTerreno(mundo.regiaoTerrenoInicial())
}

func (mundo *Mundo) cortarTerreno(ponto rl.Vector3, raio float32) bool {
	plano := mundo.planejarCorte(ponto, raio)
	if len(plano) == 0 {
		return false
	}

	malhasPropostas := mundo.prepararMalhasPropostas(plano)

	if motivo := mundo.motivoBloqueioCorte(malhasPropostas, ponto, raio); motivo != "" {
		fmt.Println("Corte recusado:", motivo)
		return false
	}

	gramasAfetadas := mundo.gramasAfetadasPorCorte(malhasPropostas, ponto, raio)

	amostrasAlteradas := make(map[CoordAmostraTerreno]struct{}, len(plano))

	for chave, densidade := range plano {
		mundo.alteracoesTerreno[chave] = densidade
		amostrasAlteradas[chave] = struct{}{}

		// Uma amostra na divisa pertence às duas regiões. Atualizamos ambas.
		for _, regiao := range mundo.regioes {
			ix, iy, iz, pertence := regiao.indicesDaAmostraGlobal(
				chave.X, chave.Y, chave.Z,
			)
			if pertence {
				regiao.Campo[ix][iy][iz] = densidade
			}
		}
	}

	// A normal de um vértice de borda usa amostras até uma unidade fora da
	// região. Uma edição próxima à borda também exige atualizar a vizinha.
	for _, regiao := range mundo.regioes {
		for amostra := range amostrasAlteradas {
			x := amostra.X - int(regiao.Coord.X)*tamanhoRegiaoXZ
			y := amostra.Y - int(regiao.Coord.Y)*tamanhoRegiaoY
			z := amostra.Z - int(regiao.Coord.Z)*tamanhoRegiaoXZ
			if x >= -9 && x <= 9 && y >= -5 && y <= 5 && z >= -9 && z <= 9 {
				mundo.reconstruirRegiaoTerreno(regiao)
				break
			}
		}
	}

	mundo.removerGramasEscavadas(gramasAfetadas)
	return true

}

func (mundo *Mundo) planejarCorte(ponto rl.Vector3, raio float32) map[CoordAmostraTerreno]float32 {
	amostras := make(map[CoordAmostraTerreno]float32)

	for _, regiao := range mundo.regioes {
		// cortarEsfera pode modificar o campo recebido. A cópia preserva
		// o terreno real enquanto calculamos o resultado proposto.
		campoProposto := regiao.Campo
		origem := regiao.posicaoMundo()

		cortarEsfera(
			&campoProposto,
			ponto.X-origem.X,
			ponto.Y-origem.Y,
			ponto.Z-origem.Z,
			raio,
			func(x, y, z int, densidade float32) {
				chave := CoordAmostraTerreno{
					X: int(regiao.Coord.X)*tamanhoRegiaoXZ + x,
					Y: int(regiao.Coord.Y)*tamanhoRegiaoY + y,
					Z: int(regiao.Coord.Z)*tamanhoRegiaoXZ + z,
				}
				amostras[chave] = densidade
			},
		)
	}

	return amostras
}

func (mundo *Mundo) regiaoTerrenoInicial() *RegiaoTerreno {
	return mundo.regioes[CoordRegiaoTerreno{}]
}

func (mundo *Mundo) descarregarTerreno() {
	for coord := range mundo.regioes {
		mundo.descarregarRegiaoTerreno(coord)
	}
}

func (mundo *Mundo) carregarRegiaoTerreno(coord CoordRegiaoTerreno) {

	if _, carregada := mundo.regioes[coord]; carregada {
		return
	}

	regiao := &RegiaoTerreno{Coord: coord}
	preencherCampoRegiao(&regiao.Campo, coord, mundo.ruido)

	// Cada amostra local consulta sua coordenada global.
	for ix := range regiao.Campo {
		for iy := range regiao.Campo[ix] {
			for iz := range regiao.Campo[ix][iy] {
				chave := CoordAmostraTerreno{
					X: int(coord.X)*tamanhoRegiaoXZ + ix - tamanhoRegiaoXZ/2,
					Y: int(coord.Y)*tamanhoRegiaoY + iy - tamanhoRegiaoY/2,
					Z: int(coord.Z)*tamanhoRegiaoXZ + iz - tamanhoRegiaoXZ/2,
				}

				if densidade, editada := mundo.alteracoesTerreno[chave]; editada {
					regiao.Campo[ix][iy][iz] = densidade
				}
			}
		}
	}

	regiao.Modelo = criarModeloTerreno(&regiao.Campo, regiao.Coord, mundo.densidadeGlobal)
	regiao.MascaraEstrada = carregarMascaraEstrada(mundo.estradaInicial, coord)
	mundo.regioes[coord] = regiao
}

func (mundo *Mundo) densidadeGlobal(x, y, z int) float32 {
	chave := CoordAmostraTerreno{X: x, Y: y, Z: z}
	if densidade, editada := mundo.alteracoesTerreno[chave]; editada {
		return densidade
	}
	return densidadeNoPonto(mundo.ruido, float32(x), float32(y), float32(z))
}

func (mundo *Mundo) descarregarRegiaoTerreno(coord CoordRegiaoTerreno) {
	regiao, carregada := mundo.regioes[coord]
	if !carregada {
		return
	}

	mundo.descarregarArvoresRegiao(coord)
	mundo.descarregarAssentamentoRegiao(coord)
	mundo.descarregarRochasRegiao(coord)
	mundo.descarregarGramaRegiao(coord)

	if regiao.MascaraEstrada.ID != 0 {
		rl.UnloadTexture(regiao.MascaraEstrada)
	}

	rl.UnloadModel(regiao.Modelo)
	delete(mundo.regioes, coord)
}

func (mundo *Mundo) superficieCarregada(x, z float32) bool {
	coord := coordenadaRegiao(x, 0, z)
	_, carregada := mundo.regioes[coord]
	return carregada
}

func (mundo *Mundo) carregarBordaVegetacaoRegiao(coord CoordRegiaoTerreno, catalogo *Catalogo) {
	regiao := mundo.regioes[coord]
	if mundo.versaoGerador > versaoGeradorAnterior || coord.Y != 0 || regiao == nil || regiao.BordaCarregada {
		return
	}

	arvores := gerarArvores(
		mundo, catalogo.modeloObrigatorio(TipoArvore), mundo.ocupacao,
		regiaoGeracaoParaTipo(coord, TipoArvore, versaoGerador),
		gerarCandidatosBorda(TipoArvore, int64(mundo.semente), coord, 100), 2,
	)

	for i := range arvores {
		if _, removida := mundo.recursosRemovidos[arvores[i].ID]; removida {
			arvores[i].Existe = false
			mundo.ocupacao.remover(arvores[i].ID)
		}
	}
	mundo.arvores = append(mundo.arvores, arvores...)

	rochas := gerarRochas(
		mundo, catalogo.modeloObrigatorio(TipoRocha), mundo.ocupacao,
		regiaoGeracaoParaTipo(coord, TipoRocha, versaoGerador),
		gerarCandidatosBorda(TipoRocha, int64(mundo.semente), coord, 200), 1,
	)

	for i := range rochas {
		if _, removida := mundo.recursosRemovidos[rochas[i].ID]; removida {
			rochas[i].Existe = false
			mundo.ocupacao.remover(rochas[i].ID)
		}
	}
	mundo.rochas = append(mundo.rochas, rochas...)

	gramas := gerarGrama(
		mundo, catalogo.modeloObrigatorio(TipoGrama), mundo.ocupacao,
		regiaoGeracaoParaTipo(coord, TipoGrama, versaoGerador),
		gerarCandidatosBorda(TipoGrama, int64(mundo.semente), coord, 400), 55,
	)

	for i := range gramas {
		if _, removida := mundo.recursosRemovidos[gramas[i].ID]; removida {
			gramas[i].Existe = false
			mundo.ocupacao.remover(gramas[i].ID)
		}
	}

	mundo.gramas = append(mundo.gramas, gramas...)
	mundo.removerGramaSobrePisos(regiaoGeracaoNaCoord(coord).Area)
	regiao.BordaCarregada = true
}

func (mundo *Mundo) planosNaArea(area AreaXZ) []PlanoTerritorial {
	return planosNaArea(int64(mundo.semente), area)
}
