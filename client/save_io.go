package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func caminhoSave() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("localizar pasta do usuário: %w", err)
	}
	return filepath.Join(base, "Sobrevivencia", "saves", "slot-1.json"), nil
}

func gravarPartida(estado PartidaSalva) (string, error) {
	caminho, err := caminhoSave()
	if err != nil {
		return "", err
	}

	dados, err := json.MarshalIndent(estado, "", "  ")
	if err != nil {
		return "", fmt.Errorf("converter partida para JSON: %w", err)
	}
	dados = append(dados, '\n')

	pasta := filepath.Dir(caminho)
	if err := os.MkdirAll(pasta, 0700); err != nil {
		return "", fmt.Errorf("criar pasta de saves: %w", err)
	}

	temporario, err := os.CreateTemp(pasta, "slot-1-*.tmp")
	if err != nil {
		return "", fmt.Errorf("criar arquivo temporário: %w", err)
	}
	defer os.Remove(temporario.Name())
	defer temporario.Close()

	if _, err := temporario.Write(dados); err != nil {
		return "", fmt.Errorf("escrever save: %w", err)
	}
	if err := temporario.Sync(); err != nil {
		return "", fmt.Errorf("sincronizar save: %w", err)
	}
	if err := temporario.Close(); err != nil {
		return "", fmt.Errorf("fechar save: %w", err)
	}
	if err := os.Rename(temporario.Name(), caminho); err != nil {
		return "", fmt.Errorf("substituir save anterior: %w", err)
	}

	return caminho, nil
}

func lerPartida() (PartidaSalva, error) {
	caminho, err := caminhoSave()
	if err != nil {
		return PartidaSalva{}, err
	}

	dados, err := os.ReadFile(caminho)
	if err != nil {
		return PartidaSalva{}, fmt.Errorf("ler partida: %w", err)
	}

	var estado PartidaSalva
	if err := json.Unmarshal(dados, &estado); err != nil {
		return PartidaSalva{}, fmt.Errorf("save inválido: %w", err)
	}

	if estado.VersaoFormato != 1 &&
		estado.VersaoFormato != 2 &&
		estado.VersaoFormato != 3 &&
		estado.VersaoFormato != 4 &&
		estado.VersaoFormato != 5 &&
		estado.VersaoFormato != 6 &&
		estado.VersaoFormato != 7 &&
		estado.VersaoFormato != versaoFormatoSave {
		return PartidaSalva{}, fmt.Errorf(
			"versão do save incompatível: %d",
			estado.VersaoFormato,
		)
	}
	if estado.Geracao.VersaoAssentamento < 0 ||
		estado.Geracao.VersaoAssentamento > versaoAssentamentoAtual {
		return PartidaSalva{}, fmt.Errorf("versão de assentamento incompatível: %d",
			estado.Geracao.VersaoAssentamento)
	}

	if estado.Geracao.VersaoGerador != versaoGeradorAnterior &&
		estado.Geracao.VersaoGerador != 3 &&
		estado.Geracao.VersaoGerador != versaoGerador {
		return PartidaSalva{}, fmt.Errorf(
			"versão do gerador incompatível: %d",
			estado.Geracao.VersaoGerador,
		)
	}

	if estado.Geracao.TipoRuido != tipoRuidoAtual ||
		estado.Geracao.FrequenciaRuido != frequenciaRuidoAtual ||
		estado.Geracao.AmplitudeTerreno != amplitudeTerrenoAtual {
		return PartidaSalva{}, fmt.Errorf("configuração de geração incompatível")
	}

	return estado, nil
}
