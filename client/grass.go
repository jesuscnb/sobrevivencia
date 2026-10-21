package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Grama struct {
	ID      IDObjeto
	Posicao rl.Vector3
	Escala  float32
	Rotacao float32
}

func gerarGrama(terreno, modeloGrama rl.Model, indice *IndiceOcupacao, semente int) []Grama {
	const quantidade = 80
	const maxTentativas = 400

	gramas := make([]Grama, 0, quantidade)
	regiao := regiaoInicial()
	candidatos := gerarCandidatos(TipoGrama, int64(semente), regiao, maxTentativas)

	for _, candidato := range candidatos {
		if len(gramas) >= quantidade {
			break
		}
		posicao, ok := posicaoNoTerreno(terreno, modeloGrama, candidato.X, candidato.Z, candidato.Escala, 0.9)
		if !ok {
			continue
		}
		if !podePosicionar(indice, TipoGrama, posicao) {
			continue
		}

		gramas = append(gramas, Grama{
			ID:      IDObjeto{Tipo: TipoGrama, RegiaoX: regiao.X, RegiaoZ: regiao.Z, Ordem: candidato.Ordem},
			Posicao: posicao,
			Escala:  candidato.Escala,
			Rotacao: candidato.Rotacao,
		})
		registrarOcupacao(indice, gramas[len(gramas)-1].ID, TipoGrama, posicao)

	}

	return gramas
}

func (mundo *Mundo) removerGramaNaArea(areaPiso AreaXZ) {
	const margem float32 = 0.4

	areaLimpeza := AreaXZ{
		MinX: areaPiso.MinX - margem,
		MaxX: areaPiso.MaxX + margem,
		MinZ: areaPiso.MinZ - margem,
		MaxZ: areaPiso.MaxZ + margem,
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
