# Cliente do jogo

Todos os arquivos `.go` desta pasta pertencem ao mesmo `package main`. O Go os compila juntos.

| Arquivo | Responsabilidade |
| --- | --- |
| `main.go` | Abre e fecha a janela. |
| `game.go` | Guarda o estado do jogo e organiza cada quadro. |
| `terrain.go` | Calcula e modifica as densidades do terreno. |
| `terrain_model.go` | Converte as densidades em malha e carrega o modelo 3D. |
| `player.go` | Move a câmera e verifica colisões do jogador. |
| `resources.go` | Posiciona os cubos e calcula o que está na mira. |
| `render.go` | Desenha o mundo e os textos na tela. |

Para executar, abra um terminal nesta pasta e use:

```powershell
go run .
```

Use `go run .` porque `go run main.go` compila apenas aquele arquivo e deixa de fora as funções dos outros arquivos.

Controles atuais: movimento e câmera em primeira pessoa; `E` recolhe um cubo na mira; `C` corta o terreno quando ele está ao alcance; `M` alterna a visualização em arame.
