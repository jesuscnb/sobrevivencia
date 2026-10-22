package main

import (
	"fmt"
	"math"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EfeitoSonoro uint8

const (
	SomPassoGramaA EfeitoSonoro = iota
	SomPassoGramaB
	SomPassoEstradaA
	SomPassoEstradaB
	SomColetaMadeira
	SomColetaPedra
	SomConstrucao
	SomEscavacao
	SomDano
	SomColetaCubo
	quantidadeEfeitos
)

var arquivosEfeitos = [...]string{
	"step_grass_a.wav", "step_grass_b.wav",
	"step_road_a.wav", "step_road_b.wav",
	"collect_wood.wav", "collect_stone.wav", "build.wav",
	"dig.wav", "hurt.wav", "collect_cube.wav",
}

type SomJogo struct {
	efeitos         [quantidadeEfeitos]rl.Sound
	ambiente        rl.Music
	config          ConfigAudio
	ativo           bool
	silenciado      bool
	distanciaPassos float32
	proximoPassoB   bool
}

func carregarSom(config ConfigAudio) SomJogo {
	som := SomJogo{config: config, ativo: rl.IsAudioDeviceReady()}
	if !som.ativo {
		fmt.Println("Áudio indisponível: o jogo continuará sem som")
		return som
	}
	base := filepath.Join("assets", "audio")
	for i, nome := range arquivosEfeitos {
		som.efeitos[i] = rl.LoadSound(filepath.Join(base, nome))
		if !rl.IsSoundValid(som.efeitos[i]) {
			fmt.Println("Efeito sonoro não carregado:", nome)
		}
	}
	som.ambiente = rl.LoadMusicStream(filepath.Join(base, "wind.wav"))
	if rl.IsMusicValid(som.ambiente) {
		som.ambiente.Looping = true
		rl.PlayMusicStream(som.ambiente)
	} else {
		fmt.Println("Ambiente sonoro não carregado: wind.wav")
	}
	som.aplicarVolumes()
	return som
}

func (som *SomJogo) aplicarVolumes() {
	if !som.ativo {
		return
	}
	geral := som.config.Geral
	if som.silenciado {
		geral = 0
	}
	for i, efeito := range som.efeitos {
		if rl.IsSoundValid(efeito) {
			volume := geral * som.config.Efeitos
			if i < int(SomColetaMadeira) {
				volume *= 0.65 // passos abaixo das interações
			}
			rl.SetSoundVolume(efeito, volume)
		}
	}
	if rl.IsMusicValid(som.ambiente) {
		rl.SetMusicVolume(som.ambiente, geral*som.config.Ambiente)
	}
}

func (som *SomJogo) alternarSilencio() {
	som.silenciado = !som.silenciado
	som.aplicarVolumes()
	if som.silenciado {
		fmt.Println("Som desativado (F10 para ativar)")
	} else {
		fmt.Println("Som ativado (F10 para desativar)")
	}
}

func (som *SomJogo) atualizar() {
	if som.ativo && rl.IsMusicValid(som.ambiente) {
		rl.UpdateMusicStream(som.ambiente)
	}
}

func (som *SomJogo) tocar(efeito EfeitoSonoro) {
	if !som.ativo || som.silenciado || efeito >= quantidadeEfeitos {
		return
	}
	if rl.IsSoundValid(som.efeitos[efeito]) {
		rl.PlaySound(som.efeitos[efeito])
	}
}

func (som *SomJogo) atualizarPassos(mundo *Mundo, anterior, atual rl.Vector3) {
	dx, dz := atual.X-anterior.X, atual.Z-anterior.Z
	distancia := float32(math.Hypot(float64(dx), float64(dz)))
	if distancia < 0.001 || distancia > 2 {
		return
	}
	som.distanciaPassos += distancia
	const intervaloPassos float32 = 0.9
	if som.distanciaPassos < intervaloPassos {
		return
	}
	som.distanciaPassos -= intervaloPassos
	trechos := mundo.estradaInicial.trechosNaArea(areaAoRedor(atual, 2), alcanceEstrada)
	naEstrada := sobrepoeFaixaEstrada(trechos, atual.X, atual.Z, 0)
	efeito := SomPassoGramaA
	if naEstrada {
		efeito = SomPassoEstradaA
	}
	if som.proximoPassoB {
		efeito++
	}
	som.proximoPassoB = !som.proximoPassoB
	som.tocar(efeito)
}

func (som *SomJogo) descarregar() {
	if !som.ativo {
		return
	}
	if rl.IsMusicValid(som.ambiente) {
		rl.StopMusicStream(som.ambiente)
		rl.UnloadMusicStream(som.ambiente)
	}
	for _, efeito := range som.efeitos {
		if rl.IsSoundValid(efeito) {
			rl.UnloadSound(efeito)
		}
	}
}
