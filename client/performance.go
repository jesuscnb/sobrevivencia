package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const capacidadeHistoricoQuadros = 120

type DiagnosticoDesempenho struct {
	Visivel bool

	quadros        [capacidadeHistoricoQuadros]float64
	proximaAmostra int
	quantidade     int
	ultimaLeitura  time.Time

	MedianaMS        float64
	Percentil95MS    float64
	PiorMS           float64
	HeapMB           float64
	MalhaGPUmb       float64
	Triangulos       int
	Regioes          int
	UltimaCargaMS    float64
	UltimaDescargaMS float64
}

func percentisQuadro(amostras []float64) (mediana, p95, pior float64) {
	if len(amostras) == 0 {
		return 0, 0, 0
	}
	ordenadas := append([]float64(nil), amostras...)
	sort.Float64s(ordenadas)
	meio := len(ordenadas) / 2
	mediana = ordenadas[meio]
	if len(ordenadas)%2 == 0 {
		mediana = (ordenadas[meio-1] + ordenadas[meio]) / 2
	}
	indice95 := int(math.Ceil(float64(len(ordenadas))*0.95)) - 1
	return mediana, ordenadas[indice95], ordenadas[len(ordenadas)-1]
}

func (d *DiagnosticoDesempenho) registrarQuadro(segundos float32, mundo *Mundo) {
	if segundos > 0 && segundos < 5 {
		d.quadros[d.proximaAmostra] = float64(segundos) * 1000
		d.proximaAmostra = (d.proximaAmostra + 1) % capacidadeHistoricoQuadros
		if d.quantidade < capacidadeHistoricoQuadros {
			d.quantidade++
		}
	}
	if !d.ultimaLeitura.IsZero() && time.Since(d.ultimaLeitura) < time.Second {
		return
	}
	d.ultimaLeitura = time.Now()
	d.MedianaMS, d.Percentil95MS, d.PiorMS = percentisQuadro(d.quadros[:d.quantidade])

	var memoria runtime.MemStats
	runtime.ReadMemStats(&memoria)
	d.HeapMB = float64(memoria.HeapAlloc) / (1024 * 1024)
	d.Regioes = len(mundo.regioes)
	d.Triangulos = 0
	var bytesMalha int64
	for _, regiao := range mundo.regioes {
		for _, malha := range regiao.Modelo.GetMeshes() {
			d.Triangulos += int(malha.TriangleCount)
			// Esta malha usa apenas posição e normal: 6 float32 por vértice.
			bytesMalha += int64(malha.VertexCount) * 6 * 4
		}
	}
	d.MalhaGPUmb = float64(bytesMalha) / (1024 * 1024)
	d.UltimaCargaMS = float64(mundo.ultimoCarregamento) / float64(time.Millisecond)
	d.UltimaDescargaMS = float64(mundo.ultimaDescarga) / float64(time.Millisecond)
}

func (d *DiagnosticoDesempenho) desenhar() {
	if !d.Visivel {
		return
	}
	x := int32(rl.GetScreenWidth() - 370)
	if x < 10 {
		x = 10
	}
	rl.DrawRectangle(x, 10, 360, 145, rl.NewColor(9, 15, 20, 220))
	rl.DrawText("F4 | Desempenho", x+10, 17, 17, rl.White)
	rl.DrawText(fmt.Sprintf("Quadro: med %.1f | p95 %.1f | max %.1f ms",
		d.MedianaMS, d.Percentil95MS, d.PiorMS), x+10, 42, 15, rl.White)
	rl.DrawText(fmt.Sprintf("Regioes: %d | triangulos: %d", d.Regioes, d.Triangulos),
		x+10, 64, 15, rl.White)
	rl.DrawText(fmt.Sprintf("Heap: %.1f MB | malhas GPU est.: %.1f MB",
		d.HeapMB, d.MalhaGPUmb), x+10, 86, 15, rl.White)
	rl.DrawText(fmt.Sprintf("Ultima carga: %.1f ms | descarga: %.1f ms",
		d.UltimaCargaMS, d.UltimaDescargaMS), x+10, 108, 15, rl.White)
}
