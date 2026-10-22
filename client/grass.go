package main

import rl "github.com/gen2brain/raylib-go/raylib"

const margemRemocaoGrama float32 = 0.4

type Grama struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Escala  float32
	Rotacao float32
}

func gerarGrama(mundo *Mundo, modeloGrama rl.Model, indice *IndiceOcupacao, regiao RegiaoGeracao, candidatos []CandidatoObjeto, quantidade int) []Grama {
	gramas := make([]Grama, 0, quantidade)

	tentar := func(candidato CandidatoObjeto) {
		posicao, ok := posicaoNoTerreno(mundo, modeloGrama, candidato.X, candidato.Z, candidato.Escala, 0.9)
		if !ok {
			return
		}
		if !podePosicionar(indice, TipoGrama, posicao) {
			return
		}

		gramas = append(gramas, Grama{
			ID:      IDObjeto{Tipo: TipoGrama, RegiaoX: regiao.X, RegiaoZ: regiao.Z, Ordem: candidato.Ordem},
			Posicao: posicao,
			Escala:  candidato.Escala,
			Rotacao: candidato.Rotacao,
		})
		registrarOcupacao(indice, gramas[len(gramas)-1].ID, TipoGrama, posicao)
	}

	for _, candidato := range candidatos {
		if len(gramas) >= quantidade {
			break
		}
		tentar(candidato)
	}

	return gramas
}

func (mundo *Mundo) carregarGramaRegiao(coord CoordRegiaoTerreno, modelo rl.Model) {
	regiao := mundo.regioes[coord]
	if coord.Y != 0 || regiao == nil || regiao.GramaCarregada {
		return
	}

	area := regiaoGeracaoParaTipo(coord, TipoGrama, mundo.versaoGerador)
	candidatos := gerarCandidatos(TipoGrama, int64(mundo.semente), area, 400)
	novas := gerarGrama(mundo, modelo, mundo.ocupacao, area, candidatos, 80)
	mundo.gramas = append(mundo.gramas, novas...)
	regiao.GramaCarregada = true
	mundo.removerGramaSobrePisos(area.Area)
}

func (mundo *Mundo) removerGramaSobrePisos(area AreaXZ) {
	// Pisos salvos podem já existir quando esta região é carregada.
	buscaPisos := AreaXZ{
		MinX: area.MinX - margemRemocaoGrama,
		MaxX: area.MaxX + margemRemocaoGrama,
		MinZ: area.MinZ - margemRemocaoGrama,
		MaxZ: area.MaxZ + margemRemocaoGrama,
	}
	for _, registro := range mundo.ocupacao.consultar(buscaPisos) {
		if registro.Tipo == TipoPiso {
			mundo.removerGramaNaArea(registro.Area)
		}
	}
}

func (mundo *Mundo) descarregarGramaRegiao(coord CoordRegiaoTerreno) {
	if coord.Y != 0 {
		return
	}

	restantes := mundo.gramas[:0]
	for _, grama := range mundo.gramas {
		if grama.ID.RegiaoX == coord.X && grama.ID.RegiaoZ == coord.Z {
			mundo.ocupacao.remover(grama.ID)
			continue
		}
		restantes = append(restantes, grama)
	}
	mundo.gramas = restantes
}

func (mundo *Mundo) removerGramaNaArea(areaPiso AreaXZ) {

	areaLimpeza := AreaXZ{
		MinX: areaPiso.MinX - margemRemocaoGrama,
		MaxX: areaPiso.MaxX + margemRemocaoGrama,
		MinZ: areaPiso.MinZ - margemRemocaoGrama,
		MaxZ: areaPiso.MaxZ + margemRemocaoGrama,
	}

	idsRemovidos := make(map[IDObjeto]struct{})

	for _, registro := range mundo.ocupacao.consultar(areaLimpeza) {
		if registro.Tipo != TipoGrama {
			continue
		}

		p := registro.Posicao
		if p.X >= areaLimpeza.MinX && p.X <= areaLimpeza.MaxX &&
			p.Z >= areaLimpeza.MinZ && p.Z <= areaLimpeza.MaxZ {
			idsRemovidos[registro.ID] = struct{}{}
		}
	}

	if len(idsRemovidos) == 0 {
		return
	}

	restantes := make([]Grama, 0, len(mundo.gramas))
	for _, grama := range mundo.gramas {
		if _, remover := idsRemovidos[grama.ID]; remover {
			mundo.ocupacao.remover(grama.ID)
			continue
		}
		restantes = append(restantes, grama)
	}
	mundo.gramas = restantes
}
