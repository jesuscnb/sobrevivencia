# Cliente do jogo

Todos os arquivos `.go` desta pasta pertencem ao mesmo `package main`. O Go os compila juntos.

| Arquivo | Responsabilidade |
| --- | --- |
| `main.go` | Abre e fecha a janela. |
| `game.go` | Guarda o estado do jogo e organiza cada quadro. |
| `terrain.go` | Calcula e modifica as densidades do terreno. |
| `terrain_model.go` | Converte as densidades em malha e carrega o modelo 3D. |
| `terrain_material.go` | Carrega um shader e duas texturas compartilhados pelo terreno. |
| `sky.go` | Desenha o céu estilizado conforme a direção da câmera. |
| `lighting.go` | Define a direção do sol compartilhada pelo céu e pelo terreno. |
| `performance.go` | Mede tempo de quadro, regiões ativas e uso aproximado de memória. |
| `player.go` | Move a câmera e verifica colisões do jogador. |
| `resources.go` | Posiciona os cubos e calcula o que está na mira. |
| `render.go` | Desenha o mundo e os textos na tela. |

Para executar, abra um terminal nesta pasta e use:

```powershell
go run .
```

Use `go run .` porque `go run main.go` compila apenas aquele arquivo e deixa de fora as funções dos outros arquivos.

Controles atuais: movimento e câmera em primeira pessoa; `E` recolhe um cubo na mira; `C` corta o terreno quando ele está ao alcance; `M` alterna a visualização em arame.

## Terreno e desempenho

- `F4` mostra ou esconde o painel de desempenho. A mediana, o percentil 95 e o pior quadro usam os últimos 120 quadros; os outros números são atualizados aproximadamente uma vez por segundo.
- `F6` alterna entre o terreno texturizado e a cor antiga. Use a mesma rota com as duas opções para comparar o custo visual do shader.
- `Heap` é a memória em uso pelo Go. `Malhas GPU est.` soma posições e normais das malhas ativas; é uma estimativa que não inclui texturas nem recursos internos do driver.
- As texturas originais ficam em `assets/terrain/`. Para reproduzi-las, execute `go run ./tools/terrain_textures.go` nesta pasta. O shader usa coordenadas globais, por isso o padrão continua entre regiões.
- O céu diurno é desenhado antes da cena 3D com o shader em `assets/sky/sky.fs`. O horizonte acompanha a câmera; o céu não depende dos chunks carregados.
