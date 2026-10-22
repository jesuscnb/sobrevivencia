package main

const (
	tipoRuidoAtual                = "perlin"
	frequenciaRuidoAtual  float32 = 0.12
	amplitudeTerrenoAtual float32 = 0.45
)

// Ponto central para a futura tela de opções. Valores entre 0 e 1.
type ConfigAudio struct {
	Geral    float32
	Efeitos  float32
	Ambiente float32
}

func configAudioPadrao() ConfigAudio {
	return ConfigAudio{Geral: 0.7, Efeitos: 0.85, Ambiente: 0.35}
}
