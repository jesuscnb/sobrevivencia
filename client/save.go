package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const versaoFormatoSave = 4

type VetorSalvo struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type GeracaoSalva struct {
	Semente          int     `json:"semente"`
	VersaoGerador    int     `json:"versao_gerador"`
	TipoRuido        string  `json:"tipo_ruido"`
	FrequenciaRuido  float32 `json:"frequencia_ruido"`
	AmplitudeTerreno float32 `json:"amplitude_terreno"`
}

type JogadorSalvo struct {
	Posicao    VetorSalvo `json:"posicao"`
	Alvo       VetorSalvo `json:"alvo"`
	VidaAtual  int        `json:"vida_atual"`
	VidaMaxima int        `json:"vida_maxima"`
}

type InimigoSalvo struct {
	RegiaoX         int32             `json:"regiao_x"`
	RegiaoZ         int32             `json:"regiao_z"`
	Ordem           uint32            `json:"ordem"`
	Posicao         VetorSalvo        `json:"posicao"`
	Origem          VetorSalvo        `json:"origem"`
	Estado          EstadoInimigo     `json:"estado"`
	Vida            int               `json:"vida"`
	DirecaoPatrulha uint8             `json:"direcao_patrulha"`
	RecargaAtaque   float32           `json:"recarga_ataque"`
	Visual          TipoVisualInimigo `json:"visual"`
}

type InventarioSalvo struct {
	Madeira int `json:"madeira"`
	Pedra   int `json:"pedra"`
}

type RecursoRemovidoSalvo struct {
	Tipo    string `json:"tipo"`
	RegiaoX int32  `json:"regiao_x"`
	RegiaoZ int32  `json:"regiao_z"`
	Ordem   uint32 `json:"ordem"`
}

type PecaSalva struct {
	Tipo    string     `json:"tipo"`
	Posicao VetorSalvo `json:"posicao"`
	Escala  float32    `json:"escala"`
	Rotacao float32    `json:"rotacao"`
}

type AmostraTerrenoSalva struct {
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Z         int     `json:"z"`
	Densidade float32 `json:"densidade"`
}

type PartidaSalva struct {
	VersaoFormato     int                    `json:"versao_formato"`
	Geracao           GeracaoSalva           `json:"geracao"`
	Jogador           JogadorSalvo           `json:"jogador"`
	Inventario        InventarioSalvo        `json:"inventario"`
	RecursosRemovidos []RecursoRemovidoSalvo `json:"recursos_removidos"`
	Construcoes       []PecaSalva            `json:"construcoes"`
	AlteracoesTerreno []AmostraTerrenoSalva  `json:"alteracoes_terreno"`
	Inimigos          []InimigoSalvo         `json:"inimigos"`
}

func vetorParaSave(v rl.Vector3) VetorSalvo {
	return VetorSalvo{X: v.X, Y: v.Y, Z: v.Z}
}

func recursoParaSave(tipo string, id IDObjeto) RecursoRemovidoSalvo {
	return RecursoRemovidoSalvo{
		Tipo:    tipo,
		RegiaoX: id.RegiaoX,
		RegiaoZ: id.RegiaoZ,
		Ordem:   id.Ordem,
	}
}

func capturarPartida(jogo *Jogo) (PartidaSalva, error) {
	estado := PartidaSalva{
		VersaoFormato: versaoFormatoSave,
		Geracao: GeracaoSalva{
			Semente:          jogo.mundo.semente,
			VersaoGerador:    jogo.mundo.versaoGerador,
			TipoRuido:        tipoRuidoAtual,
			FrequenciaRuido:  frequenciaRuidoAtual,
			AmplitudeTerreno: amplitudeTerrenoAtual,
		},
		Jogador: JogadorSalvo{
			Posicao:    vetorParaSave(jogo.camera.Position),
			Alvo:       vetorParaSave(jogo.camera.Target),
			VidaAtual:  jogo.vida.Atual,
			VidaMaxima: jogo.vida.Maxima,
		},
		Inventario: InventarioSalvo{
			Madeira: jogo.inventario.Madeira,
			Pedra:   jogo.inventario.Pedra,
		},
		RecursosRemovidos: make([]RecursoRemovidoSalvo, 0),
		Construcoes:       make([]PecaSalva, 0, len(jogo.mundo.construcoes)),
		AlteracoesTerreno: make([]AmostraTerrenoSalva, 0),
		Inimigos:          make([]InimigoSalvo, 0, len(jogo.mundo.inimigos)),
	}

	var campoOriginal [17][9][17]float32
	preencherCampo(&campoOriginal, novoRuido(jogo.mundo.semente))

	for ix := range jogo.mundo.campo {
		for iy := range jogo.mundo.campo[ix] {
			for iz := range jogo.mundo.campo[ix][iy] {
				densidadeAtual := jogo.mundo.campo[ix][iy][iz]
				if densidadeAtual == campoOriginal[ix][iy][iz] {
					continue
				}

				estado.AlteracoesTerreno = append(
					estado.AlteracoesTerreno,
					AmostraTerrenoSalva{
						X:         ix - 8,
						Y:         iy - 4,
						Z:         iz - 8,
						Densidade: densidadeAtual,
					},
				)
			}
		}
	}

	for _, cubo := range jogo.mundo.cubos {
		if !cubo.Existe {
			estado.RecursosRemovidos = append(
				estado.RecursosRemovidos,
				recursoParaSave("cubo", cubo.ID),
			)
		}
	}
	for _, arvore := range jogo.mundo.arvores {
		if !arvore.Existe {
			estado.RecursosRemovidos = append(
				estado.RecursosRemovidos,
				recursoParaSave("arvore", arvore.ID),
			)
		}
	}
	for _, rocha := range jogo.mundo.rochas {
		if !rocha.Existe {
			estado.RecursosRemovidos = append(
				estado.RecursosRemovidos,
				recursoParaSave("rocha", rocha.ID),
			)
		}
	}

	for _, peca := range jogo.mundo.construcoes {
		var tipo string
		switch peca.Tipo {
		case TipoPiso:
			tipo = "piso"
		case TipoParede:
			tipo = "parede"
		case TipoTelhado:
			tipo = "telhado"
		default:
			return PartidaSalva{}, fmt.Errorf(
				"tipo de construção desconhecido: %d", peca.Tipo,
			)
		}

		estado.Construcoes = append(estado.Construcoes, PecaSalva{
			Tipo:    tipo,
			Posicao: vetorParaSave(peca.Posicao),
			Escala:  peca.Escala,
			Rotacao: peca.Rotacao,
		})
	}

	for _, inimigo := range jogo.mundo.inimigos {
		estado.Inimigos = append(estado.Inimigos, InimigoSalvo{
			RegiaoX:         inimigo.ID.RegiaoX,
			RegiaoZ:         inimigo.ID.RegiaoZ,
			Ordem:           inimigo.ID.Ordem,
			Posicao:         vetorParaSave(inimigo.Posicao),
			Origem:          vetorParaSave(inimigo.Origem),
			Estado:          inimigo.Estado,
			Vida:            inimigo.Vida,
			DirecaoPatrulha: inimigo.DirecaoPatrulha,
			RecargaAtaque:   inimigo.RecargaAtaque,
			Visual:          inimigo.Visual,
		})
	}

	return estado, nil
}
