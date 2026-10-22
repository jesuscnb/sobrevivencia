# Duas regiões de terreno — plano de implementação

> **Para quem executar:** seguir este plano em passos pequenos, com revisão visual do autor entre tarefas. O autor escreve o código do jogo; o assistente fornece trechos, localização exata e confere o resultado. Nenhum arquivo de teste temporário fica no projeto.

**Objetivo:** atravessar da região inicial para uma vizinha em X, com terreno e colisão contínuos, preservando o jogo atual e a leitura de saves v4.

**Arquitetura:** cada região de volume tem uma chave inteira `(X,Y,Z)`, seu campo e sua malha. `Mundo` mantém as regiões carregadas e fornece consultas de terreno aos sistemas do jogo. A amostragem usa coordenadas globais para que as duas regiões concordem na borda. Neste marco, as duas regiões são carregadas no início; carregamento dinâmico e texturas pertencem a planos seguintes.

**Stack:** Go 1.27.1; FastNoiseLite; Polyform/marching; raylib-go 0.60.1.

**Especificação:** `docs/superpowers/specs/2026-10-21-mundo-expansivel-design.md`.

## Restrições gerais

- A região inicial continua cobrindo X/Z de −8 a 8 e Y de −4 a 4; cada região tem 16 × 8 × 16 células e 17 × 9 × 17 amostras.
- A segunda região tem chave `(1,0,0)` e cobre X de 8 a 24; a amostra global X=8 é compartilhada.
- Manter a regra de densidade, a semente e a versão do gerador atuais para que a região inicial e os saves v4 preservem sua forma.
- Objetos existentes permanecem na região inicial neste marco. A geração, descarga e persistência de objetos em outras regiões pertencem ao carregamento dinâmico.
- A grade de navegação dos inimigos ainda cobre a região inicial; quando o jogador cruza para a vizinha, os inimigos mantêm o comportamento atual de não encontrar rota fora da grade. Navegação entre regiões virá depois, sem aumentar a grade indefinidamente.
- Usar o carregamento via OBJ atual apenas para estas duas regiões carregadas no início. Antes do streaming, validar uma malha em memória e a propriedade dos buffers na versão instalada de raylib-go.
- Código novo deve ficar separado por responsabilidade; nenhuma mudança de versão de dependência é necessária para este marco.
- Conferir visualmente em execução, além de `go build ./...` e `go vet ./...` dentro de `client`. Testes auxiliares podem ser temporários e removidos depois.

## Mapa de arquivos

- Criar `client/terrain_region.go`: chave, coordenadas, campo e geração por amostras globais.
- Modificar `client/terrain_model.go`: extrair uma malha por região e posicioná-la corretamente.
- Modificar `client/world.go`: guardar regiões e oferecer acesso ao terreno, sem expor uma única malha global.
- Modificar `client/player.go`, `client/resources.go`, `client/placement.go`, `client/construction.go`, `client/enemy.go` e `client/render.go`: consultar ou desenhar as regiões pertinentes.
- Modificar `client/game.go`: travessia e corte sobre o terreno correto; descarregar todos os modelos no encerramento.
- Modificar `client/save.go`, `client/save_load.go` e `client/save_io.go`: preservar a leitura de v4 e preparar alterações com coordenadas globais.

## Foco da revisão

1. X=8 e posições negativas devem escolher a região certa, sem amostras diferentes na borda.
2. O jogador não pode cair nem atravessar uma região sem malha/collision pronta.
3. Raycast de mira e de altura deve escolher a colisão mais próxima, inclusive perto da borda.
4. F5/F9 devem conservar os cortes, recursos e construções de saves v4 da região inicial.
5. Criar, substituir e descarregar as duas malhas não deve causar vazamento ou uso após liberação.

---

### Tarefa 1: Endereçar regiões e gerar o campo por coordenadas globais

**Arquivos:** criar `client/terrain_region.go`; adaptar `client/terrain.go` sem mudar o jogo visível.

**Interfaces propostas:** `type CoordRegiaoTerreno struct { X, Y, Z int32 }`; `type RegiaoTerreno struct { Coord CoordRegiaoTerreno; Campo [17][9][17]float32; Modelo rl.Model }`; `func coordenadaRegiao(x, y, z float32) CoordRegiaoTerreno`; `func preencherCampoRegiao(campo *[17][9][17]float32, coord CoordRegiaoTerreno, ruido *fastnoise.State[float32])`.

- [ ] **Passo 1:** implementar a conversão de posição para região com divisão por piso: `floor((x+8)/16)`, `floor((y+4)/8)`, `floor((z+8)/16)`. `RegiaoGeracao`, usada para distribuir objetos, conserva seu significado próprio.
- [ ] **Passo 2:** gerar cada amostra `(ix,iy,iz)` na posição global `(16*coord.X+ix-8, 8*coord.Y+iy-4, 16*coord.Z+iz-8)` usando `densidadeNoPonto`.
- [ ] **Passo 3:** verificar temporariamente que `(0,0,0)` gera o mesmo campo atual, que a amostra X=8 das regiões 0 e 1 é idêntica e que X=8 pertence à região 1, X=7.99 à 0 e X=−8.01 à −1. Remover o código auxiliar de verificação.
- [ ] **Passo 4:** executar `gofmt` nos arquivos alterados e `go build ./...` em `client`; o jogo original deve abrir como antes.

### Tarefa 2: Fazer `Mundo` possuir a região inicial

**Arquivos:** `client/world.go`, `client/terrain_model.go`, `client/player.go`, `client/resources.go`, `client/placement.go`, `client/construction.go`, `client/enemy.go`, `client/render.go`, `client/game.go`, `client/save.go`, `client/save_load.go`.

**Interfaces propostas:** `Mundo.regioes map[CoordRegiaoTerreno]*RegiaoTerreno`; `func (mundo *Mundo) colisaoTerreno(raio rl.Ray) rl.RayCollision`; `func (mundo *Mundo) alturaTerreno(x, z float32) (float32, bool)`; `func (mundo *Mundo) descarregarTerreno()`.

- [ ] **Passo 1:** construir a região `(0,0,0)` com o campo da tarefa 1 e a malha atual; conferir que ela é visualmente idêntica ao terreno anterior.
- [ ] **Passo 2:** mover o campo e o modelo de `Mundo` para essa região; trocar o acesso direto por métodos de consulta. Para colisão por raio, usar a transformação de cada modelo na posição da região e devolver o acerto mais próximo.
- [ ] **Passo 3:** adaptar movimento, mira, `posicaoNoTerreno`, `pisoTemApoio`, colocação de construções, suporte dos inimigos, desenho, corte, save e limpeza para esses métodos. Uma consulta sem região carregada retorna ausência de chão e impede avançar para o vazio.
- [ ] **Passo 4:** executar `go build ./...`, `go vet ./...` e abrir o jogo. Conferir caminhada, coleta, construção, inimigo, corte e F5/F9 na região inicial.

### Tarefa 3: Mostrar e atravessar a região vizinha

**Arquivos:** `client/world.go`, `client/terrain_model.go`, `client/render.go`, `client/player.go`, `client/game.go`.

**Interfaces propostas:** `func criarModeloTerrenoRegiao(regiao *RegiaoTerreno) rl.Model`; `func (mundo *Mundo) regiaoNaPosicao(x, y, z float32) (*RegiaoTerreno, bool)`.

- [ ] **Passo 1:** criar também `(1,0,0)` na inicialização, com amostras globais. Sua malha usa a mesma geometria local do modelo inicial e uma translação de +16 em X no desenho e nas consultas de colisão.
- [ ] **Passo 2:** percorrer as regiões carregadas para desenhar e consultar a colisão mais próxima; verificar visualmente se os dois terrenos se encostam em X=8.
- [ ] **Passo 3:** substituir o limite fixo `7.5` de `player.go` por um bloqueio baseado na disponibilidade de terreno/collision à frente; manter uma margem do corpo nas bordas externas carregadas. Preservar a direção da câmera quando um passo for bloqueado.
- [ ] **Passo 4:** caminhar da região inicial à vizinha e voltar, olhando para o chão e para a borda. O jogador não deve cair, atravessar o chão nem ficar preso em X=8. Confirmar também `go build ./...` e `go vet ./...`.

### Tarefa 4: Preservar escavação e save na primeira travessia

**Arquivos:** `client/terrain_region.go`, `client/game.go`, `client/save.go`, `client/save_load.go`, `client/save_io.go`.

**Interfaces propostas:** `func (mundo *Mundo) cortarEsfera(x, y, z, raio float32)`; alterações de densidade identificadas por coordenada global inteira.

- [ ] **Passo 1:** ao cortar, modificar todas as regiões carregadas que possuem amostras globais atingidas; reconstruir somente as malhas alteradas. Um corte centrado em X=8 deve aparecer dos dois lados.
- [ ] **Passo 2:** continuar lendo saves v4 com suas coordenadas da região inicial; incluir v4 na validação de `lerPartida` ao elevar `versaoFormatoSave` para 5. Gravar v5 com alterações de terreno em coordenadas globais, mantendo semente e versão do gerador. Não modificar o arquivo salvo até validar toda a entrada.
- [ ] **Passo 3:** criar um save v4 de referência antes da mudança de formato, abrir esse save na versão nova e conferir jogador, recursos, construções, inimigos e cortes. Criar um save novo após um corte na borda, sair, abrir novamente e verificar os dois lados.
- [ ] **Passo 4:** executar `go build ./...`, `go vet ./...` e a rota completa de ida e volta com F5/F9. Verificar que modelos antigos são descarregados uma única vez após substituição.

## Encerramento deste marco

Registrar FPS e tempo de quadro ao caminhar e cortar nas duas regiões. O resultado deste plano é uma travessia correta, ainda com apenas duas malhas carregadas na inicialização. O plano seguinte fará streaming, medição de memória e substituição segura do OBJ temporário; outro plano tratará material, estradas e minimapa.
