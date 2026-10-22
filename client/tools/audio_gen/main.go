// Gerador reproduzível dos primeiros efeitos originais do jogo.
// Execute na pasta client: go run ./tools/audio_gen
package main

import (
	"encoding/binary"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

const sampleRate = 22050

func main() {
	base := filepath.Join("assets", "audio")
	if err := os.MkdirAll(base, 0755); err != nil {
		panic(err)
	}
	write(base, "step_grass_a.wav", effect(0.17, 1, 92, 0.68, 0.04))
	write(base, "step_grass_b.wav", effect(0.16, 2, 103, 0.62, 0.05))
	write(base, "step_road_a.wav", effect(0.13, 3, 133, 0.42, 0.18))
	write(base, "step_road_b.wav", effect(0.14, 4, 151, 0.40, 0.20))
	write(base, "collect_wood.wav", effect(0.24, 5, 360, 0.22, 0.23))
	write(base, "collect_stone.wav", effect(0.29, 6, 590, 0.08, 0.34))
	write(base, "build.wav", build())
	write(base, "dig.wav", dig())
	write(base, "hurt.wav", effect(0.36, 8, 73, 0.58, 0.11))
	write(base, "collect_cube.wav", cube())
	write(base, "wind.wav", wind())
}

func effect(seconds float64, seed int64, frequency, body, grit float64) []float64 {
	rng := rand.New(rand.NewSource(seed))
	n := int(seconds * sampleRate)
	samples := make([]float64, n)
	filtered := 0.0
	for i := range samples {
		t := float64(i) / sampleRate
		u := float64(i) / float64(n)
		filtered = 0.72*filtered + 0.28*(rng.Float64()*2-1)
		envelope := math.Pow(1-u, 2.5) * math.Min(1, t/0.006)
		pitch := frequency * (1 - 0.38*u)
		wave := math.Sin(2 * math.Pi * pitch * t)
		samples[i] = envelope * (body*wave + grit*filtered)
	}
	return samples
}

func build() []float64 {
	samples := effect(0.34, 9, 210, 0.36, 0.36)
	second := effect(0.20, 10, 310, 0.26, 0.30)
	start := int(0.12 * sampleRate)
	for i, value := range second {
		if start+i < len(samples) {
			samples[start+i] += value * 0.65
		}
	}
	return samples
}

func dig() []float64 {
	rng := rand.New(rand.NewSource(11))
	n := int(0.42 * sampleRate)
	samples := make([]float64, n)
	filtered := 0.0
	for i := range samples {
		u := float64(i) / float64(n)
		filtered = filtered*0.83 + (rng.Float64()*2-1)*0.17
		envelope := math.Sin(math.Pi*u) * math.Pow(1-u, 0.5)
		samples[i] = envelope * (0.63*filtered + 0.13*math.Sin(2*math.Pi*(180-100*u)*float64(i)/sampleRate))
	}
	return samples
}

func cube() []float64 {
	n := int(0.36 * sampleRate)
	samples := make([]float64, n)
	for i := range samples {
		t := float64(i) / sampleRate
		u := float64(i) / float64(n)
		frequency := 660.0
		if u > 0.45 {
			frequency = 880
		}
		samples[i] = 0.27 * math.Sin(2*math.Pi*frequency*t) *
			math.Min(1, t/0.01) * math.Pow(1-u, 1.5)
	}
	return samples
}

// Vento periódico: a interpolação dá o mesmo valor no início e no fim do loop.
func wind() []float64 {
	const seconds = 8
	n := seconds * sampleRate
	samples := make([]float64, n)
	rng := rand.New(rand.NewSource(12))
	knots := make([]float64, 129)
	for i := 0; i < 128; i++ {
		knots[i] = rng.Float64()*2 - 1
	}
	knots[128] = knots[0]
	for i := range samples {
		p := float64(i) * 128 / float64(n)
		k := int(p)
		f := p - float64(k)
		f = f * f * (3 - 2*f)
		gust := knots[k]*(1-f) + knots[k+1]*f
		wave := math.Sin(2 * math.Pi * float64(i) / float64(n) * 3)
		samples[i] = 0.15*gust + 0.04*wave
	}
	return samples
}

func write(base, name string, samples []float64) {
	file, err := os.Create(filepath.Join(base, name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	dataSize := uint32(len(samples) * 2)
	file.Write([]byte("RIFF"))
	binary.Write(file, binary.LittleEndian, dataSize+36)
	file.Write([]byte("WAVEfmt "))
	binary.Write(file, binary.LittleEndian, uint32(16))
	binary.Write(file, binary.LittleEndian, uint16(1))
	binary.Write(file, binary.LittleEndian, uint16(1))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(file, binary.LittleEndian, uint16(2))
	binary.Write(file, binary.LittleEndian, uint16(16))
	file.Write([]byte("data"))
	binary.Write(file, binary.LittleEndian, dataSize)
	for _, value := range samples {
		value = math.Max(-1, math.Min(1, value))
		binary.Write(file, binary.LittleEndian, int16(value*32767))
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}
