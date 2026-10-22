# Primeira estrada visível — plano de implementação

> **Para quem executar:** o autor escreve o código do jogo em passos pequenos, com trechos e explicações do assistente. Diagnósticos auxiliares são temporários e devem ser removidos. Nenhum arquivo de teste permanece no projeto.

**Objetivo:** mostrar uma estrada de terra contínua e suavemente curva que passe perto dos quatro cantos da área inicial de 80 × 80 unidades, atravesse divisas de regiões e acompanhe a superfície do terreno.

**Arquitetura:** um caminho lógico global é calculado uma vez por semente a partir de `marcosIniciais`. Cada região carregada recebe uma pequena máscara de estrada derivada desse mesmo caminho; o shader mistura a textura de solo com a grama onde a máscara indica a faixa. A máscara acompanha o ciclo de vida da região, enquanto o caminho lógico independe das malhas carregadas.

**Stack:** Go, raylib-go 0.60.1, shader GLSL 330, sem nova dependência.

**Especificação:** `docs/superpowers/specs/2026-10-21-mundo-expansivel-design.md`, seções “Estradas e assentamentos”, “Material e aparência” e “Save, banco de dados e conteúdo de IA”.

## Decisão técnica

- **Escolha:** máscara de textura por região, amostrada pelo material do terreno. O shader já recebe posição global e desenha todas as regiões com o mesmo material; a estrada pode seguir o chão e os cortes sem sobrepor outra malha.
- Uma malha fina desenhada sobre o solo exigiria nova projeção após cada corte e poderia apresentar sobreposição visual. Calcular a distância até todos os trechos da estrada para cada pixel no shader teria custo crescente com o tamanho do caminho. A máscara fixa o trabalho na criação da região e acrescenta uma amostragem de textura por pixel.
- Primeiro usar a textura de solo existente com tom próprio para validar a faixa. Uma textura de estrada estilizada e variações de borda pertencem ao acabamento visual antes de considerar esta estrada pronta para lançamento.

## Restrições

- Preservar a semente, os cortes, os recursos, as construções e a leitura dos saves atuais. Este primeiro traçado visual não remove objetos existentes nem altera o formato de save.
- O caminho é o mesmo independentemente da ordem de carregamento. Máscaras vizinhas usam as mesmas coordenadas globais; não podem abrir fendas na divisa.
- Criar máscara apenas para regiões ativas, descarregar sua textura com a região e medir o tempo de geração e o FPS. Não recalcular caminho ou máscara em cada quadro.
- O anel inicial passa perto dos marcos de canto, sem exigir passagem pelos vértices exatos da área. Um acesso ao centro e conexões para outros planos territoriais serão acrescentados após validar o anel.
- Neste corte vertical, árvores ou pedras antigas podem sobrepor visualmente a estrada. A entrega seguinte reserva a faixa antes de gerar recursos de mundos novos, com regras de compatibilidade para saves existentes; lotes de construção usam o mesmo traçado depois disso.

## Mapa de arquivos

- Criar `client/road_path.go`: amostragem determinística da curva fechada pelos quatro marcos e consulta de trechos por área.
- Criar `client/road_mask.go`: gerar a máscara local a partir da distância ao caminho global e carregar/descarregar a textura de região.
- Modificar `client/world.go` e `client/terrain_region.go`: possuir o caminho lógico e a máscara da região; liberar a textura ao descarregar.
- Modificar `client/terrain_material.go` e `client/assets/terrain/terrain.fs`: vincular a máscara da região e misturar o visual da estrada com o terreno.
- Modificar `client/render.go`: passar a máscara da região ao material antes de cada `DrawModel`.

## Foco da revisão

1. Mesma semente e marcos produzem a mesma curva após F5/F9 ou retorno à região; outra ordem de carregamento não a altera.
2. Nos limites X/Z de 16 unidades, a faixa não apresenta uma emenda nem muda de largura.
3. Cortar o terreno sob ou ao lado da estrada não deixa uma malha de estrada flutuando; a faixa só aparece em superfície voltada para cima.
4. Descarregar e recarregar regiões libera e recria máscaras sem vazamento de textura ou travamento.
5. FPS permanece próximo dos 60 observados no cenário atual, inclusive caminhando pela área inicial.

---

### Tarefa 1: Construir o caminho lógico

**Arquivos:** criar `client/road_path.go`; modificar `client/world.go`.

**Interfaces:** `type CaminhoEstrada struct { Pontos []PontoTerritorial }`; `func gerarEstradaInicial(semente int64) CaminhoEstrada`; `func (c CaminhoEstrada) trechosNaArea(area AreaXZ, margem float32) [][2]PontoTerritorial`.

- [ ] Amostrar uma curva Catmull–Rom fechada pelos quatro valores de `marcosIniciais(semente)`, com espaçamento máximo de 1 unidade entre pontos. A curva usa os pontos anteriores e seguintes para manter tangentes contínuas entre trechos.
- [ ] Guardar o caminho em `Mundo` antes de carregar a primeira região. Consultar apenas segmentos cuja caixa X/Z cruza a área ampliada pela largura da estrada.
- [ ] Mostrar temporariamente a linha do caminho em `render.go`, somente sobre terreno carregado, para conferir quatro cantos, curvas e divisas; remover esse desenho quando a máscara estiver visível.
- [ ] Executar `gofmt`, `go build ./...` e verificar visualmente o circuito ao caminhar.

### Tarefa 2: Gerar a máscara por região

**Arquivos:** criar `client/road_mask.go`; modificar `client/terrain_region.go` e `client/world.go`.

**Interfaces:** `func gerarMascaraEstrada(c CaminhoEstrada, coord CoordRegiaoTerreno) rl.Texture2D`; campo `RegiaoTerreno.MascaraEstrada rl.Texture2D`.

- [ ] Para cada região em Y=0, gerar 64 × 64 amostras internas mais uma amostra de borda em cada lado (textura 66 × 66), todas em coordenadas globais; o valor depende da menor distância X/Z aos trechos relevantes. Centro da estrada com largura total de 2,4 unidades e transição lateral de 0,6 unidade. As amostras de borda dão continuidade à interpolação entre regiões.
- [ ] Criar a textura quando a região é carregada e descarregá-la junto com seu modelo. Região sem estrada pode usar máscara vazia compartilhada, evitando textura própria desnecessária.
- [ ] Conferir temporariamente o mesmo ponto global na borda de duas máscaras e que não surgem mais máscaras conforme o jogador se afasta.
- [ ] Executar `gofmt`, `go build ./...` e `go vet ./...` em `client`.

### Tarefa 3: Mostrar a estrada no material existente

**Arquivos:** `client/terrain_material.go`, `client/assets/terrain/terrain.fs`, `client/render.go`.

**Interfaces:** vincular `roadMask` ao shader e a máscara da região ao material no desenho de cada `RegiaoTerreno`.

- [ ] Adicionar uma amostragem de `roadMask` em coordenadas locais da região, passadas ao shader antes de `DrawModel`; misturar solo em tom de estrada sobre superfícies com normal voltada para cima. Manter grama e solo atuais fora da faixa.
- [ ] Remover a linha de depuração. Caminhar pelo circuito e verificar largura, continuidade, curvas, céu e legibilidade da vegetação.
- [ ] Cortar perto da estrada e atravessar a divisa; salvar, recarregar, afastar-se e voltar. Confirmar que a faixa acompanha a superfície e não altera cortes ou construções.
- [ ] Executar `go build ./...`, `go vet ./...` e medir FPS/tempo de carregamento. Se o custo exceder o orçamento atual, reduzir a resolução da máscara ou limitar trechos consultados antes de acrescentar arte.

## Próxima entrega

Antes de povoar lotes, a estrada passa a reservar sua faixa na geração de **mundos novos**; árvores, rochas e grama são posicionadas depois dessa reserva. Saves anteriores mantêm objetos e construções. Depois o caminho gera candidatos a lotes em suas margens, aceitos apenas quando há apoio, inclinação adequada, acesso e espaço livre; limpar vegetação somente nos lotes que receberem uma construção.
