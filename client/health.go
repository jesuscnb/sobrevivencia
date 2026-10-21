package main

type Vida struct {
	Atual  int
	Maxima int
}

func novaVida(maxima int) Vida {
	return Vida{Atual: maxima, Maxima: maxima}
}

func (vida *Vida) ReceberDano(quantidade int) {
	if quantidade <= 0 || vida.Atual == 0 {
		return
	}

	vida.Atual -= quantidade
	if vida.Atual < 0 {
		vida.Atual = 0
	}
}
