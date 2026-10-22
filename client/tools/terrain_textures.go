package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const tamanhoTextura = 128

func hash(x, y, s uint32) float64 {
	v := x*374761393 + y*668265263 + s*2246822519
	v = (v ^ (v >> 13)) * 1274126177
	return float64(v^(v>>16)) / float64(^uint32(0))
}

func ruidoRepetivel(x, y, passo int, s uint32) float64 {
	periodo := tamanhoTextura / passo
	gx, gy := x/passo, y/passo
	fx, fy := float64(x%passo)/float64(passo), float64(y%passo)/float64(passo)
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	a := hash(uint32(gx), uint32(gy), s)
	b := hash(uint32((gx+1)%periodo), uint32(gy), s)
	c := hash(uint32(gx), uint32((gy+1)%periodo), s)
	d := hash(uint32((gx+1)%periodo), uint32((gy+1)%periodo), s)
	return (a+(b-a)*fx)*(1-fy) + (c+(d-c)*fx)*fy
}

func corCanal(base int, desvio float64) uint8 {
	valor := math.Round(float64(base) + desvio)
	if valor < 0 {
		valor = 0
	}
	if valor > 255 {
		valor = 255
	}
	return uint8(valor)
}

func textura(grama bool) *image.NRGBA {
	imagem := image.NewNRGBA(image.Rect(0, 0, tamanhoTextura, tamanhoTextura))
	for y := 0; y < tamanhoTextura; y++ {
		for x := 0; x < tamanhoTextura; x++ {
			grande := ruidoRepetivel(x, y, 32, 17)
			medio := ruidoRepetivel(x, y, 8, 29)
			fino := ruidoRepetivel(x, y, 2, 43)
			variacao := (grande-0.5)*28 + (medio-0.5)*14 + (fino-0.5)*5
			celula := hash(uint32(x/4), uint32(y/4), 71)
			if grama {
				folha := 0.0
				if celula > 0.94 && x%4 <= 1 && y%4 <= 2 {
					folha = 15
				} else if celula < 0.06 && x%4 == 2 {
					folha = -12
				}
				imagem.SetNRGBA(x, y, color.NRGBA{
					R: corCanal(75, variacao*0.55+folha*0.4),
					G: corCanal(137, variacao+folha),
					B: corCanal(68, variacao*0.45+folha*0.25), A: 255,
				})
			} else {
				pedra := 0.0
				if celula > 0.95 && x%4 <= 2 && y%4 <= 2 {
					pedra = 17
				} else if celula < 0.05 && y%4 == 2 {
					pedra = -12
				}
				imagem.SetNRGBA(x, y, color.NRGBA{
					R: corCanal(131, variacao+pedra),
					G: corCanal(105, variacao*0.82+pedra),
					B: corCanal(78, variacao*0.65+pedra), A: 255,
				})
			}
		}
	}
	return imagem
}

func salvar(caminho string, imagem *image.NRGBA) error {
	arquivo, err := os.Create(caminho)
	if err != nil {
		return err
	}
	defer arquivo.Close()
	return png.Encode(arquivo, imagem)
}

func main() {
	pasta := filepath.Join("assets", "terrain")
	if err := os.MkdirAll(pasta, 0755); err != nil {
		panic(err)
	}
	if err := salvar(filepath.Join(pasta, "grass.png"), textura(true)); err != nil {
		panic(err)
	}
	if err := salvar(filepath.Join(pasta, "soil.png"), textura(false)); err != nil {
		panic(err)
	}
}
