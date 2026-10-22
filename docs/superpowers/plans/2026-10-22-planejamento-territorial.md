# Planejamento territorial — plano de implementação

> **Para quem executar:** o autor escreve o código do jogo em passos pequenos, com explicações e trechos fornecidos pelo assistente. Testes e diagnósticos auxiliares podem ser temporários; não deixar arquivos de teste no projeto.

**Objetivo:** decidir, a partir da semente e de coordenadas globais, onde podem existir assentamentos e por onde conexões territoriais podem cruzar fronteiras, sem depender das regiões carregadas.

**Arquitetura:** uma célula de planejamento cobre 4 × 4 regiões de terreno (64 × 64 unidades). Funções puras calculam seu plano e portais de fronteira usando semente e coordenadas inteiras. Uma consulta por área recupera somente os planos relevantes. Esta entrega define dados lógicos; estrada visível, adaptação do terreno, ocupação e edifícios entram nas entregas seguintes.

**Stack:** Go, gerador atual do jogo, sem nova dependência.

**Especificação:** `docs/superpowers/specs/2026-10-21-mundo-expansivel-design.md`, seções “Estradas e assentamentos”, “Save, banco de dados e conteúdo de IA” e “Entregas verificáveis”.

## Restrições gerais

- Preservar terreno, recursos, construções e saves existentes. O planejamento não altera a versão do gerador de vegetação nem o formato de save nesta entrega.
- Coordenadas negativas devem usar divisão por piso; a célula `(0,0)` cobre X/Z de −8 a 56, e a `(−1,0)` cobre X de −72 a −8.
- Cada portal de fronteira é calculado por uma chave canônica compartilhada: a borda leste de `(x,z)` é a borda oeste de `(x+1,z)`; a norte de `(x,z)` é a sul de `(x,z+1)`.
- O resultado deve ser independente da ordem em que células e regiões são consultadas. Nenhuma varredura do mundo inteiro nem geração por quadro.
- Planos gerados pela semente podem ser reconstruídos. Futuras alterações feitas por jogador ou agente terão IDs próprios e serão persistidas separadamente.
- Usar `const versaoPlanejamentoTerritorial = 1` na chave do gerador. Antes de usar planos em saves, registrar essa versão nos metadados do mundo para preservar resultados antigos quando as regras mudarem.

## Mapa de arquivos

- Criar `client/territory_plan.go`: coordenadas e áreas das células, IDs estáveis, candidatos a assentamento e portais canônicos.
- Criar `client/territory_query.go`: consulta dos planos que intersectam uma `AreaXZ`, sem depender do streaming de malhas.
- Modificar `client/world.go` apenas para expor uma consulta que use `mundo.semente`; sem adicionar planos ao índice de ocupação nesta entrega.
- Não modificar `client/terrain_stream.go`, `client/save.go` nem o renderizador até que a primeira estrada ou assentamento seja materializado.

## Foco da revisão

1. Em X/Z negativos e exatamente nas bordas, a posição pertence a uma única célula de planejamento.
2. Células vizinhas calculam o mesmo portal compartilhado, mesmo quando consultadas em ordem inversa.
3. A mesma semente reproduz os mesmos IDs e posições após descarregar e recarregar regiões; outra semente pode alterar candidatos.
4. Uma consulta perto da fronteira retorna ambos os planos sem duplicá-los.
5. Abrir o jogo continua com o mesmo terreno, saves e FPS, porque esta entrega não desenha nem instancia objetos novos.

---

### Tarefa 1: Endereçar células territoriais

**Arquivos:** criar `client/territory_plan.go`.

**Interfaces:** `type CoordPlanoTerritorial struct { X, Z int64 }`; `func coordenadaPlano(x, z float32) CoordPlanoTerritorial`; `func (coord CoordPlanoTerritorial) area() AreaXZ`. A largura é `const tamanhoPlanoTerritorial = 64`; aplicar `floor((x+8)/64)` e `floor((z+8)/64)`.

- [ ] Implementar as interfaces sem consultar malhas ou objetos carregados.
- [ ] Conferir temporariamente os pontos `−8.01`, `−8`, `55.99` e `56` em X, inclusive combinados com Z negativo; remover o diagnóstico auxiliar.
- [ ] Executar `gofmt` e `go build ./...` dentro de `client`.

### Tarefa 2: Criar decisões reproduzíveis e portais compartilhados

**Arquivos:** `client/territory_plan.go`.

**Interfaces:** `type PlanoTerritorial struct { ID CoordPlanoTerritorial; Area AreaXZ; CentroX, CentroZ float32; CandidatoAssentamento bool }`; `type EixoFronteira uint8` com valores X e Z; `type PortalTerritorial struct { X, Z float32 }`; `func gerarPlanoTerritorial(semente int64, coord CoordPlanoTerritorial) PlanoTerritorial`; `func portalFronteira(semente int64, coord CoordPlanoTerritorial, eixo EixoFronteira) PortalTerritorial`.

- [ ] Definir um hash inteiro estável de `(semente, coord, eixo)` em vez de depender da sequência de chamadas a um gerador aleatório compartilhado.
- [ ] Calcular o centro candidato dentro da margem interna de 16 unidades, para não colocá-lo sobre a divisa; marcar candidato em aproximadamente 1 de cada 8 células pelo hash. `CandidatoAssentamento` é apenas uma proposta, ainda sem reservar terreno.
- [ ] Calcular portais na linha exata da fronteira canônica: eixo X usa `Area.MaxX`, eixo Z usa `Area.MaxZ`; o deslocamento ao longo da linha fica entre 12 e 52 unidades, evitando cantos. O lado oeste/sul usa a chave da célula vizinha correspondente.
- [ ] Conferir temporariamente as duas leituras de cada borda e a reprodução após consultar outras células em ordem diferente; remover o diagnóstico auxiliar.
- [ ] Executar `gofmt` e `go build ./...` em `client`.

### Tarefa 3: Consultar somente a vizinhança necessária

**Arquivos:** criar `client/territory_query.go`; modificar `client/world.go`.

**Interfaces:** `func planosNaArea(semente int64, area AreaXZ) []PlanoTerritorial`; `func (mundo *Mundo) planosNaArea(area AreaXZ) []PlanoTerritorial` delega usando `int64(mundo.semente)`.

- [ ] Enumerar as coordenadas de células que intersectam `area`, calcular cada plano uma vez e devolvê-los em ordem estável por `(X,Z)`. Usar intervalo semiaberto `[Min,Max)`; para uma área pontual (`Min == Max`), retornar a célula do ponto. Área invertida retorna lista vazia.
- [ ] Conferir temporariamente uma área que atravesse X=56 e outra que atravesse X=−8; ambas devem retornar exatamente os dois lados, sem duplicações. Remover o diagnóstico auxiliar.
- [ ] Executar `gofmt`, `go build ./...` e `go vet ./...` em `client` e abrir o jogo. Terreno, recursos, construções, carregamento dinâmico, saves e FPS devem permanecer como antes.

## Saída desta entrega

Ao terminar, o jogo possui um mapa **lógico** consultável por área. A próxima entrega escolhe quais portais serão ligados, traça e mostra uma estrada contínua entre ao menos duas regiões. Só então a estrada passa a influenciar terreno, vegetação e ocupação; depois um assentamento consome um candidato validado. O minimapa lê os mesmos planos, em vez de gerar seu próprio mapa.
